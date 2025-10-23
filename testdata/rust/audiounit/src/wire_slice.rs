//! Zero-copy wire format encoding/decoding
//!
//! This module provides direct byte slice operations, avoiding the overhead
//! of Read/Write traits. This is analogous to Go's wire package which works
//! directly on []byte slices.

use byteorder::{ByteOrder, LittleEndian};

/// Wire format errors for slice operations
#[derive(Debug)]
pub enum SliceError {
    /// Buffer too small for the requested operation
    BufferTooSmall { needed: usize, available: usize },
    /// Invalid UTF-8 in string field
    InvalidUtf8(std::string::FromUtf8Error),
    /// Array length exceeds maximum (prevents DoS)
    ArrayTooLarge { size: u32, max: u32 },
    /// Invalid boolean value (must be 0 or 1)
    InvalidBool(u8),
}

impl std::fmt::Display for SliceError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            SliceError::BufferTooSmall { needed, available } => {
                write!(f, "Buffer too small: needed {}, got {}", needed, available)
            }
            SliceError::InvalidUtf8(e) => write!(f, "Invalid UTF-8: {}", e),
            SliceError::ArrayTooLarge { size, max } => {
                write!(f, "Array too large: {} > {} max", size, max)
            }
            SliceError::InvalidBool(v) => write!(f, "Invalid boolean value: {}", v),
        }
    }
}

impl std::error::Error for SliceError {}

impl From<std::string::FromUtf8Error> for SliceError {
    fn from(e: std::string::FromUtf8Error) -> Self {
        SliceError::InvalidUtf8(e)
    }
}

pub type SliceResult<T> = std::result::Result<T, SliceError>;

/// Maximum array size (prevents DoS attacks)
const MAX_ARRAY_SIZE: u32 = 10_000_000;

/// Check if buffer has enough space at the given offset
/// This is a helper for bulk operations that need bounds checking
#[inline]
pub fn check_bounds(buf: &[u8], offset: usize, size: usize) -> SliceResult<()> {
    if offset + size > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + size,
            available: buf.len(),
        });
    }
    Ok(())
}

// ============================================================================
// ENCODING - Direct byte slice operations
// ============================================================================

/// Encode a boolean at the given offset (1 byte: 0 or 1)
#[inline]
pub fn encode_bool(buf: &mut [u8], offset: usize, value: bool) -> SliceResult<()> {
    if offset >= buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + 1,
            available: buf.len(),
        });
    }
    buf[offset] = if value { 1 } else { 0 };
    Ok(())
}

/// Encode an 8-bit unsigned integer at the given offset
#[inline]
pub fn encode_u8(buf: &mut [u8], offset: usize, value: u8) -> SliceResult<()> {
    if offset >= buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + 1,
            available: buf.len(),
        });
    }
    buf[offset] = value;
    Ok(())
}

/// Encode a 16-bit unsigned integer at the given offset (little-endian)
#[inline]
pub fn encode_u16(buf: &mut [u8], offset: usize, value: u16) -> SliceResult<()> {
    if offset + 2 > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + 2,
            available: buf.len(),
        });
    }
    LittleEndian::write_u16(&mut buf[offset..], value);
    Ok(())
}

/// Encode a 32-bit unsigned integer at the given offset (little-endian)
#[inline]
pub fn encode_u32(buf: &mut [u8], offset: usize, value: u32) -> SliceResult<()> {
    if offset + 4 > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + 4,
            available: buf.len(),
        });
    }
    LittleEndian::write_u32(&mut buf[offset..], value);
    Ok(())
}

/// Encode a 64-bit unsigned integer at the given offset (little-endian)
#[inline]
pub fn encode_u64(buf: &mut [u8], offset: usize, value: u64) -> SliceResult<()> {
    if offset + 8 > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + 8,
            available: buf.len(),
        });
    }
    LittleEndian::write_u64(&mut buf[offset..], value);
    Ok(())
}

/// Encode an 8-bit signed integer at the given offset
#[inline]
pub fn encode_i8(buf: &mut [u8], offset: usize, value: i8) -> SliceResult<()> {
    encode_u8(buf, offset, value as u8)
}

/// Encode a 16-bit signed integer at the given offset (little-endian)
#[inline]
pub fn encode_i16(buf: &mut [u8], offset: usize, value: i16) -> SliceResult<()> {
    encode_u16(buf, offset, value as u16)
}

/// Encode a 32-bit signed integer at the given offset (little-endian)
#[inline]
pub fn encode_i32(buf: &mut [u8], offset: usize, value: i32) -> SliceResult<()> {
    encode_u32(buf, offset, value as u32)
}

/// Encode a 64-bit signed integer at the given offset (little-endian)
#[inline]
pub fn encode_i64(buf: &mut [u8], offset: usize, value: i64) -> SliceResult<()> {
    encode_u64(buf, offset, value as u64)
}

/// Encode a 32-bit float at the given offset (little-endian, IEEE 754)
#[inline]
pub fn encode_f32(buf: &mut [u8], offset: usize, value: f32) -> SliceResult<()> {
    encode_u32(buf, offset, value.to_bits())
}

/// Encode a 64-bit float at the given offset (little-endian, IEEE 754)
#[inline]
pub fn encode_f64(buf: &mut [u8], offset: usize, value: f64) -> SliceResult<()> {
    encode_u64(buf, offset, value.to_bits())
}

/// Encode a string: u32 length + UTF-8 bytes
/// Returns the number of bytes written
pub fn encode_string(buf: &mut [u8], offset: usize, value: &str) -> SliceResult<usize> {
    let bytes = value.as_bytes();
    let len = bytes.len() as u32;
    
    let total = 4 + bytes.len();
    if offset + total > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + total,
            available: buf.len(),
        });
    }
    
    encode_u32(buf, offset, len)?;
    buf[offset + 4..offset + 4 + bytes.len()].copy_from_slice(bytes);
    
    Ok(total)
}

/// Encode bytes: u32 length + raw bytes
/// Returns the number of bytes written
pub fn encode_bytes(buf: &mut [u8], offset: usize, value: &[u8]) -> SliceResult<usize> {
    let len = value.len() as u32;
    
    let total = 4 + value.len();
    if offset + total > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + total,
            available: buf.len(),
        });
    }
    
    encode_u32(buf, offset, len)?;
    buf[offset + 4..offset + 4 + value.len()].copy_from_slice(value);
    
    Ok(total)
}

// ============================================================================
// DECODING - Direct byte slice operations
// ============================================================================

/// Decode a boolean from the given offset
#[inline]
pub fn decode_bool(buf: &[u8], offset: usize) -> SliceResult<bool> {
    if offset >= buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + 1,
            available: buf.len(),
        });
    }
    match buf[offset] {
        0 => Ok(false),
        1 => Ok(true),
        v => Err(SliceError::InvalidBool(v)),
    }
}

/// Decode an 8-bit unsigned integer from the given offset
#[inline]
pub fn decode_u8(buf: &[u8], offset: usize) -> SliceResult<u8> {
    if offset >= buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + 1,
            available: buf.len(),
        });
    }
    Ok(buf[offset])
}

/// Decode a 16-bit unsigned integer from the given offset (little-endian)
#[inline]
pub fn decode_u16(buf: &[u8], offset: usize) -> SliceResult<u16> {
    if offset + 2 > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + 2,
            available: buf.len(),
        });
    }
    Ok(LittleEndian::read_u16(&buf[offset..]))
}

/// Decode a 32-bit unsigned integer from the given offset (little-endian)
#[inline]
pub fn decode_u32(buf: &[u8], offset: usize) -> SliceResult<u32> {
    if offset + 4 > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + 4,
            available: buf.len(),
        });
    }
    Ok(LittleEndian::read_u32(&buf[offset..]))
}

/// Decode a 64-bit unsigned integer from the given offset (little-endian)
#[inline]
pub fn decode_u64(buf: &[u8], offset: usize) -> SliceResult<u64> {
    if offset + 8 > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + 8,
            available: buf.len(),
        });
    }
    Ok(LittleEndian::read_u64(&buf[offset..]))
}

/// Decode an 8-bit signed integer from the given offset
#[inline]
pub fn decode_i8(buf: &[u8], offset: usize) -> SliceResult<i8> {
    Ok(decode_u8(buf, offset)? as i8)
}

/// Decode a 16-bit signed integer from the given offset (little-endian)
#[inline]
pub fn decode_i16(buf: &[u8], offset: usize) -> SliceResult<i16> {
    Ok(decode_u16(buf, offset)? as i16)
}

/// Decode a 32-bit signed integer from the given offset (little-endian)
#[inline]
pub fn decode_i32(buf: &[u8], offset: usize) -> SliceResult<i32> {
    Ok(decode_u32(buf, offset)? as i32)
}

/// Decode a 64-bit signed integer from the given offset (little-endian)
#[inline]
pub fn decode_i64(buf: &[u8], offset: usize) -> SliceResult<i64> {
    Ok(decode_u64(buf, offset)? as i64)
}

/// Decode a 32-bit float from the given offset (little-endian, IEEE 754)
#[inline]
pub fn decode_f32(buf: &[u8], offset: usize) -> SliceResult<f32> {
    Ok(f32::from_bits(decode_u32(buf, offset)?))
}

/// Decode a 64-bit float from the given offset (little-endian, IEEE 754)
#[inline]
pub fn decode_f64(buf: &[u8], offset: usize) -> SliceResult<f64> {
    Ok(f64::from_bits(decode_u64(buf, offset)?))
}

/// Decode a string: u32 length + UTF-8 bytes
/// Returns (String, bytes_consumed)
pub fn decode_string(buf: &[u8], offset: usize) -> SliceResult<(String, usize)> {
    let len = decode_u32(buf, offset)? as usize;
    
    if len > MAX_ARRAY_SIZE as usize {
        return Err(SliceError::ArrayTooLarge {
            size: len as u32,
            max: MAX_ARRAY_SIZE,
        });
    }
    
    let total = 4 + len;
    if offset + total > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + total,
            available: buf.len(),
        });
    }
    
    let bytes = &buf[offset + 4..offset + 4 + len];
    let s = String::from_utf8(bytes.to_vec())?;
    
    Ok((s, total))
}

/// Decode bytes: u32 length + raw bytes
/// Returns (Vec<u8>, bytes_consumed)
pub fn decode_bytes(buf: &[u8], offset: usize) -> SliceResult<(Vec<u8>, usize)> {
    let len = decode_u32(buf, offset)? as usize;
    
    if len > MAX_ARRAY_SIZE as usize {
        return Err(SliceError::ArrayTooLarge {
            size: len as u32,
            max: MAX_ARRAY_SIZE,
        });
    }
    
    let total = 4 + len;
    if offset + total > buf.len() {
        return Err(SliceError::BufferTooSmall {
            needed: offset + total,
            available: buf.len(),
        });
    }
    
    let bytes = buf[offset + 4..offset + 4 + len].to_vec();
    
    Ok((bytes, total))
}
