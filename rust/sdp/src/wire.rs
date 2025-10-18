//! Wire format encoding/decoding primitives
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
    writer: W,
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
    reader: R,
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_bool_roundtrip() {
        let mut buf = Vec::new();
        let mut enc = Encoder::new(&mut buf);
        enc.write_bool(true).unwrap();
        enc.write_bool(false).unwrap();

        let mut dec = Decoder::new(&buf[..]);
        assert_eq!(dec.read_bool().unwrap(), true);
        assert_eq!(dec.read_bool().unwrap(), false);
    }

    #[test]
    fn test_integers_roundtrip() {
        let mut buf = Vec::new();
        let mut enc = Encoder::new(&mut buf);
        enc.write_u8(0xFF).unwrap();
        enc.write_u16(0xABCD).unwrap();
        enc.write_u32(0x12345678).unwrap();
        enc.write_u64(0x123456789ABCDEF0).unwrap();

        let mut dec = Decoder::new(&buf[..]);
        assert_eq!(dec.read_u8().unwrap(), 0xFF);
        assert_eq!(dec.read_u16().unwrap(), 0xABCD);
        assert_eq!(dec.read_u32().unwrap(), 0x12345678);
        assert_eq!(dec.read_u64().unwrap(), 0x123456789ABCDEF0);
    }

    #[test]
    fn test_string_roundtrip() {
        let mut buf = Vec::new();
        let mut enc = Encoder::new(&mut buf);
        enc.write_string("Hello, SDP!").unwrap();

        let mut dec = Decoder::new(&buf[..]);
        assert_eq!(dec.read_string().unwrap(), "Hello, SDP!");
    }

    #[test]
    fn test_little_endian() {
        let mut buf = Vec::new();
        let mut enc = Encoder::new(&mut buf);
        enc.write_u32(0x12345678).unwrap();

        // Verify little-endian byte order
        assert_eq!(buf, vec![0x78, 0x56, 0x34, 0x12]);
    }
}
