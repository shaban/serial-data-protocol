//! Serial Data Protocol - Core Library
//!
//! High-performance binary serialization for low-latency IPC.
//!
//! ## Features
//!
//! - **Zero-copy decoding**: Direct buffer access without intermediate allocations
//! - **Explicit endianness**: Little-endian wire format for consistent cross-platform behavior
//! - **Optional fields**: Efficient presence flags for nullable data
//! - **Message mode**: Frame-based I/O with length prefixes
//! - **Streaming mode**: Continuous data flow without framing overhead
//!
//! ## APIs
//!
//! Two wire format APIs are provided:
//!
//! - **`wire`**: Stream-based API using `Read`/`Write` traits (flexible, composable)
//! - **`wire_slice`**: Direct byte slice API (faster, zero-abstraction overhead)
//!
//! Choose `wire_slice` for maximum performance in hot paths.
//!
//! ## Wire Format
//!
//! SDP uses a simple, efficient binary encoding:
//! - All integers are little-endian
//! - Optional fields use presence flags (1 byte bitmap per 8 fields)
//! - Strings are length-prefixed: `u32_length + utf8_bytes`
//! - Arrays are length-prefixed: `u32_length + elements`
//!
//! ## Example
//!
//! ```rust,ignore
//! use sdp::{Encoder, Decoder};
//!
//! // Generated from schema:
//! // struct AudioPlugin {
//! //     name: string
//! //     latency: u32
//! // }
//!
//! let plugin = AudioPlugin {
//!     name: "Reverb".to_string(),
//!     latency: 512,
//! };
//!
//! // Encode
//! let mut buf = Vec::new();
//! plugin.encode(&mut buf)?;
//!
//! // Decode
//! let decoded = AudioPlugin::decode(&buf)?;
//! assert_eq!(decoded.name, "Reverb");
//! assert_eq!(decoded.latency, 512);
//! ```

pub mod wire;
pub mod wire_slice;

pub use wire::{Encoder, Decoder, Error, Result};

/// Wire format version (semver-compatible)
pub const VERSION: &str = "0.2.0-rc1";
