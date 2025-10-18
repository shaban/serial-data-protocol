//! Zero-copy wire format encoding/decoding
//!
//! This module provides direct byte slice operations, avoiding the overhead
//! of Read/Write traits. This is analogous to Go's wire package which works
//! directly on []byte slices.

use byteorder::{ByteOrder, LittleEndian};

/// Wire format errors
#[derive(Debug)]
pub enum Error {
    /// Buffer too small for the requested operation
    BufferTooSmall { needed: usize, available: usize },
    /// Invalid UTF-8 in string field
    InvalidUtf8(std::string::FromUtf8Error),
    /// Array length exceeds maximum (prevents DoS)
    ArrayTooLarge { size: u32, max: u32 },
    /// Invalid boolean value (must be 0 or 1)
    InvalidBool(u8),
}

impl std::fmt::Display for Error {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Error::BufferTooSmall { needed, available } => {
                write!(f, "Buffer too small: needed {}, got {}", needed, available)
            }
            Error::InvalidUtf8(e) => write!(f, "Invalid UTF-8: {}", e),
            Error::ArrayTooLarge { size, max } => {
                write!(f, "Array too large: {} > {} max", size, max)
            }
            Error::InvalidBool(v) => write!(f, "Invalid boolean value: {}", v),
        }
    }
}

impl std::error::Error for Error {}

impl From<std::string::FromUtf8Error> for Error {
    fn from(e: std::string::FromUtf8Error) -> Self {
        Error::InvalidUtf8(e)
    }
}

pub type Result<T> = std::result::Result<T, Error>;

/// Maximum array size (prevents DoS attacks)
const MAX_ARRAY_SIZE: u32 = 10_000_000;

// ============================================================================
// ENCODING - Direct byte slice operations (like Go's wire.Encode* functions)
// ============================================================================

/// Encode a boolean at the given offset (1 byte: 0 or 1)
#[inline]
pub fn encode_bool(buf: &mut [u8], offset: usize, value: bool) -> Result<()> {
    if offset >= buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + 1,
            available: buf.len(),
        });
    }
    buf[offset] = if value { 1 } else { 0 };
    Ok(())
}

/// Encode an 8-bit unsigned integer at the given offset
#[inline]
pub fn encode_u8(buf: &mut [u8], offset: usize, value: u8) -> Result<()> {
    if offset >= buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + 1,
            available: buf.len(),
        });
    }
    buf[offset] = value;
    Ok(())
}

/// Encode a 16-bit unsigned integer at the given offset (little-endian)
#[inline]
pub fn encode_u16(buf: &mut [u8], offset: usize, value: u16) -> Result<()> {
    if offset + 2 > buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + 2,
            available: buf.len(),
        });
    }
    LittleEndian::write_u16(&mut buf[offset..], value);
    Ok(())
}

/// Encode a 32-bit unsigned integer at the given offset (little-endian)
#[inline]
pub fn encode_u32(buf: &mut [u8], offset: usize, value: u32) -> Result<()> {
    if offset + 4 > buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + 4,
            available: buf.len(),
        });
    }
    LittleEndian::write_u32(&mut buf[offset..], value);
    Ok(())
}

/// Encode a 64-bit unsigned integer at the given offset (little-endian)
#[inline]
pub fn encode_u64(buf: &mut [u8], offset: usize, value: u64) -> Result<()> {
    if offset + 8 > buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + 8,
            available: buf.len(),
        });
    }
    LittleEndian::write_u64(&mut buf[offset..], value);
    Ok(())
}

/// Encode an 8-bit signed integer at the given offset
#[inline]
pub fn encode_i8(buf: &mut [u8], offset: usize, value: i8) -> Result<()> {
    encode_u8(buf, offset, value as u8)
}

/// Encode a 16-bit signed integer at the given offset (little-endian)
#[inline]
pub fn encode_i16(buf: &mut [u8], offset: usize, value: i16) -> Result<()> {
    encode_u16(buf, offset, value as u16)
}

/// Encode a 32-bit signed integer at the given offset (little-endian)
#[inline]
pub fn encode_i32(buf: &mut [u8], offset: usize, value: i32) -> Result<()> {
    encode_u32(buf, offset, value as u32)
}

/// Encode a 64-bit signed integer at the given offset (little-endian)
#[inline]
pub fn encode_i64(buf: &mut [u8], offset: usize, value: i64) -> Result<()> {
    encode_u64(buf, offset, value as u64)
}

/// Encode a 32-bit float at the given offset (little-endian, IEEE 754)
#[inline]
pub fn encode_f32(buf: &mut [u8], offset: usize, value: f32) -> Result<()> {
    encode_u32(buf, offset, value.to_bits())
}

/// Encode a 64-bit float at the given offset (little-endian, IEEE 754)
#[inline]
pub fn encode_f64(buf: &mut [u8], offset: usize, value: f64) -> Result<()> {
    encode_u64(buf, offset, value.to_bits())
}

/// Encode a string: u32 length + UTF-8 bytes
/// Returns the number of bytes written
pub fn encode_string(buf: &mut [u8], offset: usize, value: &str) -> Result<usize> {
    let bytes = value.as_bytes();
    let len = bytes.len() as u32;
    
    // Need 4 bytes for length + string bytes
    let total = 4 + bytes.len();
    if offset + total > buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + total,
            available: buf.len(),
        });
    }
    
    // Write length
    encode_u32(buf, offset, len)?;
    
    // Write string bytes
    buf[offset + 4..offset + 4 + bytes.len()].copy_from_slice(bytes);
    
    Ok(total)
}

/// Encode bytes: u32 length + raw bytes
/// Returns the number of bytes written
pub fn encode_bytes(buf: &mut [u8], offset: usize, value: &[u8]) -> Result<usize> {
    let len = value.len() as u32;
    
    let total = 4 + value.len();
    if offset + total > buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + total,
            available: buf.len(),
        });
    }
    
    encode_u32(buf, offset, len)?;
    buf[offset + 4..offset + 4 + value.len()].copy_from_slice(value);
    
    Ok(total)
}

// ============================================================================
// DECODING - Direct byte slice operations (like Go's wire.Decode* functions)
// ============================================================================

/// Decode a boolean from the given offset
#[inline]
pub fn decode_bool(buf: &[u8], offset: usize) -> Result<bool> {
    if offset >= buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + 1,
            available: buf.len(),
        });
    }
    match buf[offset] {
        0 => Ok(false),
        1 => Ok(true),
        v => Err(Error::InvalidBool(v)),
    }
}

/// Decode an 8-bit unsigned integer from the given offset
#[inline]
pub fn decode_u8(buf: &[u8], offset: usize) -> Result<u8> {
    if offset >= buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + 1,
            available: buf.len(),
        });
    }
    Ok(buf[offset])
}

/// Decode a 16-bit unsigned integer from the given offset (little-endian)
#[inline]
pub fn decode_u16(buf: &[u8], offset: usize) -> Result<u16> {
    if offset + 2 > buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + 2,
            available: buf.len(),
        });
    }
    Ok(LittleEndian::read_u16(&buf[offset..]))
}

/// Decode a 32-bit unsigned integer from the given offset (little-endian)
#[inline]
pub fn decode_u32(buf: &[u8], offset: usize) -> Result<u32> {
    if offset + 4 > buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + 4,
            available: buf.len(),
        });
    }
    Ok(LittleEndian::read_u32(&buf[offset..]))
}

/// Decode a 64-bit unsigned integer from the given offset (little-endian)
#[inline]
pub fn decode_u64(buf: &[u8], offset: usize) -> Result<u64> {
    if offset + 8 > buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + 8,
            available: buf.len(),
        });
    }
    Ok(LittleEndian::read_u64(&buf[offset..]))
}

/// Decode an 8-bit signed integer from the given offset
#[inline]
pub fn decode_i8(buf: &[u8], offset: usize) -> Result<i8> {
    Ok(decode_u8(buf, offset)? as i8)
}

/// Decode a 16-bit signed integer from the given offset (little-endian)
#[inline]
pub fn decode_i16(buf: &[u8], offset: usize) -> Result<i16> {
    Ok(decode_u16(buf, offset)? as i16)
}

/// Decode a 32-bit signed integer from the given offset (little-endian)
#[inline]
pub fn decode_i32(buf: &[u8], offset: usize) -> Result<i32> {
    Ok(decode_u32(buf, offset)? as i32)
}

/// Decode a 64-bit signed integer from the given offset (little-endian)
#[inline]
pub fn decode_i64(buf: &[u8], offset: usize) -> Result<i64> {
    Ok(decode_u64(buf, offset)? as i64)
}

/// Decode a 32-bit float from the given offset (little-endian, IEEE 754)
#[inline]
pub fn decode_f32(buf: &[u8], offset: usize) -> Result<f32> {
    Ok(f32::from_bits(decode_u32(buf, offset)?))
}

/// Decode a 64-bit float from the given offset (little-endian, IEEE 754)
#[inline]
pub fn decode_f64(buf: &[u8], offset: usize) -> Result<f64> {
    Ok(f64::from_bits(decode_u64(buf, offset)?))
}

/// Decode a string: u32 length + UTF-8 bytes
/// Returns (String, bytes_consumed)
pub fn decode_string(buf: &[u8], offset: usize) -> Result<(String, usize)> {
    let len = decode_u32(buf, offset)? as usize;
    
    if len > MAX_ARRAY_SIZE as usize {
        return Err(Error::ArrayTooLarge {
            size: len as u32,
            max: MAX_ARRAY_SIZE,
        });
    }
    
    let total = 4 + len;
    if offset + total > buf.len() {
        return Err(Error::BufferTooSmall {
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
pub fn decode_bytes(buf: &[u8], offset: usize) -> Result<(Vec<u8>, usize)> {
    let len = decode_u32(buf, offset)? as usize;
    
    if len > MAX_ARRAY_SIZE as usize {
        return Err(Error::ArrayTooLarge {
            size: len as u32,
            max: MAX_ARRAY_SIZE,
        });
    }
    
    let total = 4 + len;
    if offset + total > buf.len() {
        return Err(Error::BufferTooSmall {
            needed: offset + total,
            available: buf.len(),
        });
    }
    
    let bytes = buf[offset + 4..offset + 4 + len].to_vec();
    
    Ok((bytes, total))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_u32_roundtrip() {
        let mut buf = [0u8; 4];
        encode_u32(&mut buf, 0, 0x12345678).unwrap();
        assert_eq!(buf, [0x78, 0x56, 0x34, 0x12]); // Little-endian
        assert_eq!(decode_u32(&buf, 0).unwrap(), 0x12345678);
    }

    #[test]
    fn test_string_roundtrip() {
        let mut buf = [0u8; 100];
        let s = "Hello, Rust!";
        let written = encode_string(&mut buf, 0, s).unwrap();
        assert_eq!(written, 4 + s.len());
        
        let (decoded, consumed) = decode_string(&buf, 0).unwrap();
        assert_eq!(decoded, s);
        assert_eq!(consumed, written);
    }

    #[test]
    fn test_bool_roundtrip() {
        let mut buf = [0u8; 2];
        encode_bool(&mut buf, 0, true).unwrap();
        encode_bool(&mut buf, 1, false).unwrap();
        assert_eq!(buf, [1, 0]);
        assert_eq!(decode_bool(&buf, 0).unwrap(), true);
        assert_eq!(decode_bool(&buf, 1).unwrap(), false);
    }

    #[test]
    fn test_buffer_too_small() {
        let mut buf = [0u8; 2];
        let err = encode_u32(&mut buf, 0, 42).unwrap_err();
        match err {
            Error::BufferTooSmall { needed, available } => {
                assert_eq!(needed, 4);
                assert_eq!(available, 2);
            }
            _ => panic!("Expected BufferTooSmall error"),
        }
    }

    #[test]
    fn test_invalid_bool() {
        let buf = [2u8]; // Invalid: must be 0 or 1
        let err = decode_bool(&buf, 0).unwrap_err();
        match err {
            Error::InvalidBool(v) => assert_eq!(v, 2),
            _ => panic!("Expected InvalidBool error"),
        }
    }
}
