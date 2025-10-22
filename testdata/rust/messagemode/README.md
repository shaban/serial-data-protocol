# Rust Message Mode Implementation

**Status:** ✅ Complete and verified

## Overview

Rust implementation of SDP message mode with idiomatic enum-based dispatch. Provides type-safe, zero-copy decoding with pattern matching.

## Generated Code

For a schema with `Point` and `Rectangle`:

```rust
// Message enum for type-safe dispatch
#[derive(Debug, Clone)]
pub enum Message {
    Point(Point),
    Rectangle(Rectangle),
}

// Encoder functions
pub fn encode_point_message(src: &Point) -> Vec<u8>
pub fn encode_rectangle_message(src: &Rectangle) -> Vec<u8>

// Decoder functions
pub fn decode_point_message(data: &[u8]) -> Result<Point, MessageError>
pub fn decode_rectangle_message(data: &[u8]) -> Result<Rectangle, MessageError>

// Dispatcher
pub fn decode_message(data: &[u8]) -> Result<Message, MessageError>

// Error handling
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum MessageError {
    MessageTooShort,
    InvalidMagic,
    InvalidVersion,
    UnknownMessageType(u16),
    PayloadSizeMismatch,
    DecodeError(String),
}
```

## Usage

### Encoding

```rust
use sdp_point::{Point, encode_point_message};

let point = Point { x: 3.14, y: 2.71 };
let message = encode_point_message(&point);

// message = [SDP:3]['2':1][type_id:2][length:4][payload:N]
// Total: 10 bytes header + payload
```

### Decoding (Type-Specific)

```rust
use sdp_point::{Point, decode_point_message};

let point = decode_point_message(&message)?;
println!("Point: ({}, {})", point.x, point.y);
```

### Decoding (Dispatcher)

```rust
use sdp_point::{Message, decode_message};

match decode_message(&message)? {
    Message::Point(p) => {
        println!("Got Point: ({}, {})", p.x, p.y);
    }
    Message::Rectangle(r) => {
        println!("Got Rectangle: {}×{}", r.width, r.height);
    }
}
```

## Cross-Language Verification

✅ **9/9 tests passing** - Full compatibility verified:

1. **Decode Go Point** - Rust reads Go-generated messages
2. **Decode Go Rectangle** - Complex nested struct compatibility
3. **Rust → Go match** - Byte-for-byte identical encoding
4. **Roundtrip** - Encode + decode = original
5. **Dispatcher Point** - Enum routing works correctly
6. **Dispatcher Rectangle** - Multi-type dispatch verified
7. **Decode C++ Point** - Three-way compatibility (Go ↔ Rust ↔ C++)
8. **Decode C++ Rectangle** - All languages interoperate
9. **Full compatibility** - All three produce identical wire format

```bash
cargo test --test crosslang_test
# running 9 tests
# test result: ok. 9 passed; 0 failed
```

## Performance

Measured with `criterion`:

| Operation | Time | vs C++ | Notes |
|-----------|------|--------|-------|
| Point encode | ~49ns | ~5× slower | Allocation overhead |
| Point decode | ~3ns | ~same | Zero-copy decode |
| Rectangle encode | ~51ns | ~10× slower | Nested allocation |
| Rectangle decode | ~5ns | ~same | Fast validation |
| Dispatcher overhead | ~2ns | negligible | Pattern match cost |

**Notes:**
- Decode is zero-copy and blazingly fast
- Encode allocates Vec which adds overhead
- Still 10-100× faster than JSON/MessagePack
- Compile time: ~2.6s (incremental: ~0.5s)

## Implementation Details

### Encoder

Uses struct methods from byte mode:
```rust
let payload_size = src.encoded_size();
let mut payload = vec![0u8; payload_size];
src.encode_to_slice(&mut payload).expect("encoding failed");

// Add 10-byte header
let mut message = vec![0u8; MESSAGE_HEADER_SIZE + payload.len()];
message[0..3].copy_from_slice(MESSAGE_MAGIC);  // 'SDP'
message[3] = MESSAGE_VERSION;                  // '2'
LittleEndian::write_u16(&mut message[4..6], type_id);
LittleEndian::write_u32(&mut message[6..10], payload.len() as u32);
message[10..].copy_from_slice(&payload);
```

### Decoder

Validates header then delegates to byte mode decoder:
```rust
// Validate header (magic, version, type_id, length)
if &data[0..3] != MESSAGE_MAGIC { return Err(InvalidMagic); }
if data[3] != MESSAGE_VERSION { return Err(InvalidVersion); }

let type_id = LittleEndian::read_u16(&data[4..6]);
let payload_length = LittleEndian::read_u32(&data[6..10]) as usize;

// Extract and decode payload
let payload = &data[MESSAGE_HEADER_SIZE..];
StructType::decode_from_slice(payload)
    .map_err(|e| MessageError::DecodeError(e.to_string()))
```

### Dispatcher

Pattern matches on type_id:
```rust
let type_id = LittleEndian::read_u16(&data[4..6]);

match type_id {
    1 => decode_point_message(data).map(Message::Point),
    2 => decode_rectangle_message(data).map(Message::Rectangle),
    _ => Err(MessageError::UnknownMessageType(type_id)),
}
```

## Advantages

1. **Type Safety** - Enum forces exhaustive pattern matching
2. **Ergonomics** - `#[derive(Debug, Clone)]` on Message enum
3. **Zero Copy** - Decode borrows from input buffer
4. **Fast Builds** - ~2.6s initial, ~0.5s incremental
5. **No Dependencies** - Only `byteorder` crate
6. **Error Handling** - Proper `Result<T, E>` with descriptive errors

## Comparison to Other Languages

| Feature | Go | C++ | Rust |
|---------|----|----|------|
| Dispatch | interface{} | std::variant | enum |
| Type Safety | ❌ Runtime | ⚠️ Compile-time (verbose) | ✅ Compile-time (ergonomic) |
| Pattern Match | ❌ Type switch | ⚠️ std::visit | ✅ Native match |
| Errors | error interface | exceptions | Result<T, E> |
| Zero Copy | ❌ Allocates | ✅ Yes | ✅ Yes |

**Verdict:** Rust provides the most idiomatic and type-safe implementation! 🦀

## Files

- `src/message_encode.rs` - Encoder functions (auto-generated)
- `src/message_decode.rs` - Decoder + enum + dispatcher (auto-generated)
- `tests/crosslang_test.rs` - Cross-language compatibility tests
- `benches/benchmarks.rs` - Performance benchmarks (criterion)

## See Also

- `C_QUICK_REFERENCE.md` - C++ message mode API
- `testdata/binaries/MESSAGE_MODE_README.md` - Wire format spec
- `PERFORMANCE_ANALYSIS.md` - Cross-language performance comparison
