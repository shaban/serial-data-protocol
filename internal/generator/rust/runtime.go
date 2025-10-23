package rust

// Embedded runtime - these are the static wire format primitives that get
// embedded into every generated Rust crate, making them fully self-contained.

const wireRuntime = `//! Wire format encoding/decoding primitives
//!
//! Low-level functions for reading/writing SDP wire format.
//! All integers are little-endian.

use byteorder::{LittleEndian, ReadBytesExt, WriteBytesExt};
use std::io::{self, Read, Write};

/// Wire format errors
#[derive(Debug)]
pub enum Error {
    /// I/O error during encode/decode
    Io(io::Error),
    /// Invalid UTF-8 in string field
    InvalidUtf8(std::string::FromUtf8Error),
    /// Array length exceeds maximum (prevents DoS)
    ArrayTooLarge { size: u32, max: u32 },
    /// Buffer too small for expected data
    UnexpectedEof,
    /// Invalid boolean value (must be 0 or 1)
    InvalidBool(u8),
}

impl From<io::Error> for Error {
    fn from(e: io::Error) -> Self {
        Error::Io(e)
    }
}

impl From<std::string::FromUtf8Error> for Error {
    fn from(e: std::string::FromUtf8Error) -> Self {
        Error::InvalidUtf8(e)
    }
}

impl std::fmt::Display for Error {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Error::Io(e) => write!(f, "I/O error: {}", e),
            Error::InvalidUtf8(e) => write!(f, "Invalid UTF-8: {}", e),
            Error::ArrayTooLarge { size, max } => {
                write!(f, "Array too large: {} > {} max", size, max)
            }
            Error::UnexpectedEof => write!(f, "Unexpected end of buffer"),
            Error::InvalidBool(v) => write!(f, "Invalid boolean value: {}", v),
        }
    }
}

impl std::error::Error for Error {}

pub type Result<T> = std::result::Result<T, Error>;

/// Maximum array size (prevents DoS attacks)
const MAX_ARRAY_SIZE: u32 = 10_000_000;

/// Encoder for SDP wire format
pub struct Encoder<W: Write> {
    pub writer: W,
}

impl<W: Write> Encoder<W> {
    pub fn new(writer: W) -> Self {
        Self { writer }
    }

    /// Encode a boolean (1 byte: 0 or 1)
    pub fn write_bool(&mut self, value: bool) -> Result<()> {
        self.writer.write_u8(if value { 1 } else { 0 })?;
        Ok(())
    }

    /// Encode an 8-bit unsigned integer
    pub fn write_u8(&mut self, value: u8) -> Result<()> {
        self.writer.write_u8(value)?;
        Ok(())
    }

    /// Encode a 16-bit unsigned integer (little-endian)
    pub fn write_u16(&mut self, value: u16) -> Result<()> {
        self.writer.write_u16::<LittleEndian>(value)?;
        Ok(())
    }

    /// Encode a 32-bit unsigned integer (little-endian)
    pub fn write_u32(&mut self, value: u32) -> Result<()> {
        self.writer.write_u32::<LittleEndian>(value)?;
        Ok(())
    }

    /// Encode a 64-bit unsigned integer (little-endian)
    pub fn write_u64(&mut self, value: u64) -> Result<()> {
        self.writer.write_u64::<LittleEndian>(value)?;
        Ok(())
    }

    /// Encode an 8-bit signed integer
    pub fn write_i8(&mut self, value: i8) -> Result<()> {
        self.writer.write_i8(value)?;
        Ok(())
    }

    /// Encode a 16-bit signed integer (little-endian)
    pub fn write_i16(&mut self, value: i16) -> Result<()> {
        self.writer.write_i16::<LittleEndian>(value)?;
        Ok(())
    }

    /// Encode a 32-bit signed integer (little-endian)
    pub fn write_i32(&mut self, value: i32) -> Result<()> {
        self.writer.write_i32::<LittleEndian>(value)?;
        Ok(())
    }

    /// Encode a 64-bit signed integer (little-endian)
    pub fn write_i64(&mut self, value: i64) -> Result<()> {
        self.writer.write_i64::<LittleEndian>(value)?;
        Ok(())
    }

    /// Encode a 32-bit IEEE 754 float (little-endian)
    pub fn write_f32(&mut self, value: f32) -> Result<()> {
        self.writer.write_f32::<LittleEndian>(value)?;
        Ok(())
    }

    /// Encode a 64-bit IEEE 754 float (little-endian)
    pub fn write_f64(&mut self, value: f64) -> Result<()> {
        self.writer.write_f64::<LittleEndian>(value)?;
        Ok(())
    }

    /// Encode a string (u32 length + UTF-8 bytes)
    pub fn write_string(&mut self, value: &str) -> Result<()> {
        let bytes = value.as_bytes();
        self.write_u32(bytes.len() as u32)?;
        self.writer.write_all(bytes)?;
        Ok(())
    }

    /// Encode a byte array (u32 length + bytes)
    pub fn write_bytes(&mut self, value: &[u8]) -> Result<()> {
        self.write_u32(value.len() as u32)?;
        self.writer.write_all(value)?;
        Ok(())
    }
}

/// Decoder for SDP wire format
pub struct Decoder<R: Read> {
    pub reader: R,
}

impl<R: Read> Decoder<R> {
    pub fn new(reader: R) -> Self {
        Self { reader }
    }

    /// Decode a boolean (1 byte: 0 or 1)
    pub fn read_bool(&mut self) -> Result<bool> {
        match self.reader.read_u8()? {
            0 => Ok(false),
            1 => Ok(true),
            v => Err(Error::InvalidBool(v)),
        }
    }

    /// Decode an 8-bit unsigned integer
    pub fn read_u8(&mut self) -> Result<u8> {
        Ok(self.reader.read_u8()?)
    }

    /// Decode a 16-bit unsigned integer (little-endian)
    pub fn read_u16(&mut self) -> Result<u16> {
        Ok(self.reader.read_u16::<LittleEndian>()?)
    }

    /// Decode a 32-bit unsigned integer (little-endian)
    pub fn read_u32(&mut self) -> Result<u32> {
        Ok(self.reader.read_u32::<LittleEndian>()?)
    }

    /// Decode a 64-bit unsigned integer (little-endian)
    pub fn read_u64(&mut self) -> Result<u64> {
        Ok(self.reader.read_u64::<LittleEndian>()?)
    }

    /// Decode an 8-bit signed integer
    pub fn read_i8(&mut self) -> Result<i8> {
        Ok(self.reader.read_i8()?)
    }

    /// Decode a 16-bit signed integer (little-endian)
    pub fn read_i16(&mut self) -> Result<i16> {
        Ok(self.reader.read_i16::<LittleEndian>()?)
    }

    /// Decode a 32-bit signed integer (little-endian)
    pub fn read_i32(&mut self) -> Result<i32> {
        Ok(self.reader.read_i32::<LittleEndian>()?)
    }

    /// Decode a 64-bit signed integer (little-endian)
    pub fn read_i64(&mut self) -> Result<i64> {
        Ok(self.reader.read_i64::<LittleEndian>()?)
    }

    /// Decode a 32-bit IEEE 754 float (little-endian)
    pub fn read_f32(&mut self) -> Result<f32> {
        Ok(self.reader.read_f32::<LittleEndian>()?)
    }

    /// Decode a 64-bit IEEE 754 float (little-endian)
    pub fn read_f64(&mut self) -> Result<f64> {
        Ok(self.reader.read_f64::<LittleEndian>()?)
    }

    /// Decode a string (u32 length + UTF-8 bytes)
    pub fn read_string(&mut self) -> Result<String> {
        let len = self.read_u32()?;
        if len > MAX_ARRAY_SIZE {
            return Err(Error::ArrayTooLarge {
                size: len,
                max: MAX_ARRAY_SIZE,
            });
        }
        let mut buf = vec![0u8; len as usize];
        self.reader.read_exact(&mut buf)?;
        Ok(String::from_utf8(buf)?)
    }

    /// Decode a byte array (u32 length + bytes)
    pub fn read_bytes(&mut self) -> Result<Vec<u8>> {
        let len = self.read_u32()?;
        if len > MAX_ARRAY_SIZE {
            return Err(Error::ArrayTooLarge {
                size: len,
                max: MAX_ARRAY_SIZE,
            });
        }
        let mut buf = vec![0u8; len as usize];
        self.reader.read_exact(&mut buf)?;
        Ok(buf)
    }
}
`

const wireSliceRuntime = `//! Zero-copy wire format encoding/decoding
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
    InvalidUtf8(std::str::Utf8Error),
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

// Note: std::str::Utf8Error for slice operations (zero-copy validation)
// Different from std::string::FromUtf8Error used in Reader API (owned data)
impl From<std::str::Utf8Error> for SliceError {
    fn from(e: std::str::Utf8Error) -> Self {
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
    // Optimize: validate UTF-8 in-place first (zero-copy), then allocate
    // This is faster than bytes.to_vec() + String::from_utf8()
    let s = std::str::from_utf8(bytes)?.to_string();
    
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
`
