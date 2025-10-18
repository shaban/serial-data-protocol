# Rust API Design: Mode-Appropriate APIs

## Overview

Different SDP modes have different performance requirements. The generated Rust API should provide the right tool for each use case.

## Mode-API Mapping

### 1. Byte Mode (IPC) → Slice API

**Use Case:** Same-machine IPC, shared memory, audio plugins

**Why Slice API:**
- Maximum performance (4x faster encoding)
- Pre-allocated buffers common in IPC
- Zero-copy friendly
- Matches the "raw speed" philosophy

**Generated API:**
```rust
impl AllPrimitives {
    /// Encode to a pre-allocated buffer (byte mode)
    /// Returns the number of bytes written
    pub fn encode_to_slice(&self, buf: &mut [u8]) -> Result<usize> {
        use sdp::wire_slice;
        let mut offset = 0;
        
        wire_slice::encode_u32(buf, offset, self.u32_field)?;
        offset += 4;
        
        let written = wire_slice::encode_string(buf, offset, &self.str_field)?;
        offset += written;
        
        Ok(offset)
    }
    
    /// Decode from a byte slice (byte mode)
    pub fn decode_from_slice(buf: &[u8]) -> Result<Self> {
        use sdp::wire_slice;
        let mut offset = 0;
        
        let u32_field = wire_slice::decode_u32(buf, offset)?;
        offset += 4;
        
        let (str_field, consumed) = wire_slice::decode_string(buf, offset)?;
        offset += consumed;
        
        Ok(Self { u32_field, str_field })
    }
    
    /// Calculate exact size needed for encoding
    pub fn encoded_size(&self) -> usize {
        4 +  // u32_field
        4 + self.str_field.len()  // string: length + bytes
    }
}
```

**Usage (Audio Plugin IPC):**
```rust
// IPC scenario: pre-allocated shared memory buffer
let mut shared_buffer = [0u8; 1024];

// Encode directly to shared memory
let size = params.encode_to_slice(&mut shared_buffer)?;

// Other process decodes from same buffer
let decoded = AllPrimitives::decode_from_slice(&shared_buffer[..size])?;
```

### 2. Message Mode → Trait API

**Use Case:** Network protocols, file storage, type discrimination

**Why Trait API:**
- Needs to write headers (type ID, size)
- Often writes to files/sockets (io::Write)
- Composability more important than raw speed
- 10-byte overhead dwarfs trait overhead

**Generated API:**
```rust
impl ErrorMsg {
    /// Type ID for message discrimination
    pub const TYPE_ID: u64 = 0x1234567890ABCDEF; // FNV-1a hash
    
    /// Encode with message mode header (type ID + size + payload)
    pub fn encode_message<W: Write>(&self, writer: &mut W) -> Result<()> {
        use sdp::wire::{Encoder};
        use byteorder::{LittleEndian, WriteBytesExt};
        
        // Calculate payload size
        let payload_size = self.encoded_size();
        
        // Write header
        writer.write_u64::<LittleEndian>(Self::TYPE_ID)?;
        writer.write_u32::<LittleEndian>(payload_size as u32)?;
        
        // Write payload using trait API
        let mut enc = Encoder::new(writer);
        enc.write_u32(self.code)?;
        enc.write_string(&self.text)?;
        
        Ok(())
    }
    
    /// Decode message payload (after type ID is checked)
    pub fn decode_message<R: Read>(reader: &mut R) -> Result<Self> {
        use sdp::wire::Decoder;
        use byteorder::{LittleEndian, ReadBytesExt};
        
        // Read header
        let type_id = reader.read_u64::<LittleEndian>()?;
        if type_id != Self::TYPE_ID {
            return Err(Error::WrongMessageType { 
                expected: Self::TYPE_ID, 
                got: type_id 
            });
        }
        
        let _size = reader.read_u32::<LittleEndian>()?;
        
        // Decode payload
        let mut dec = Decoder::new(reader);
        let code = dec.read_u32()?;
        let text = dec.read_string()?;
        
        Ok(Self { code, text })
    }
}

/// Message dispatcher (like Go's DispatchMessage)
pub fn dispatch_message<R: Read>(reader: &mut R) -> Result<Message> {
    use byteorder::{LittleEndian, ReadBytesExt};
    
    let type_id = reader.read_u64::<LittleEndian>()?;
    
    match type_id {
        ErrorMsg::TYPE_ID => Ok(Message::Error(ErrorMsg::decode_message(reader)?)),
        DataMsg::TYPE_ID => Ok(Message::Data(DataMsg::decode_message(reader)?)),
        _ => Err(Error::UnknownMessageType(type_id)),
    }
}

pub enum Message {
    Error(ErrorMsg),
    Data(DataMsg),
}
```

**Usage (Network Protocol):**
```rust
// Send over network
let mut conn = TcpStream::connect("localhost:8080")?;
error_msg.encode_message(&mut conn)?;

// Receive and dispatch
let msg = dispatch_message(&mut conn)?;
match msg {
    Message::Error(e) => eprintln!("Error {}: {}", e.code, e.text),
    Message::Data(d) => println!("Data: {} bytes", d.payload.len()),
}
```

### 3. Streaming I/O → Trait API (Convenience)

**Use Case:** File I/O, compression, network sockets

**Why Trait API:**
- Composability with stdlib (`io::Write`, `io::Read`)
- Works with any stream type
- Convenience over maximum performance

**Generated API:**
```rust
impl AllPrimitives {
    /// Encode to any writer (convenience for files, sockets, etc.)
    pub fn encode_to_writer<W: Write>(&self, writer: &mut W) -> Result<()> {
        use sdp::wire::Encoder;
        let mut enc = Encoder::new(writer);
        
        enc.write_u32(self.u32_field)?;
        enc.write_string(&self.str_field)?;
        
        Ok(())
    }
    
    /// Decode from any reader
    pub fn decode_from_reader<R: Read>(reader: &mut R) -> Result<Self> {
        use sdp::wire::Decoder;
        let mut dec = Decoder::new(reader);
        
        let u32_field = dec.read_u32()?;
        let str_field = dec.read_string()?;
        
        Ok(Self { u32_field, str_field })
    }
}
```

**Usage (File I/O):**
```rust
// Write to file
let mut file = File::create("data.sdp")?;
plugin.encode_to_writer(&mut file)?;

// Read from file
let mut file = File::open("data.sdp")?;
let plugin = Plugin::decode_from_reader(&mut file)?;

// With compression
let mut file = File::create("data.sdp.gz")?;
let mut gz = GzEncoder::new(file, Compression::default());
plugin.encode_to_writer(&mut gz)?;
```

## Complete Generated API Surface

Each struct generates:

```rust
impl MyStruct {
    // BYTE MODE (IPC) - Slice API - FASTEST
    pub fn encode_to_slice(&self, buf: &mut [u8]) -> Result<usize>;
    pub fn decode_from_slice(buf: &[u8]) -> Result<Self>;
    pub fn encoded_size(&self) -> usize;
    
    // STREAMING - Trait API - COMPOSABLE
    pub fn encode_to_writer<W: Write>(&self, writer: &mut W) -> Result<()>;
    pub fn decode_from_reader<R: Read>(reader: &mut R) -> Result<Self>;
}
```

Each message generates (in addition to above):

```rust
impl MyMessage {
    pub const TYPE_ID: u64 = ...;
    
    // MESSAGE MODE - With headers
    pub fn encode_message<W: Write>(&self, writer: &mut W) -> Result<()>;
    pub fn decode_message<R: Read>(reader: &mut R) -> Result<Self>;
}

// Module-level dispatcher
pub fn dispatch_message<R: Read>(reader: &mut R) -> Result<MessageEnum>;
```

## Performance Characteristics

### Byte Mode (Slice API)
- **Encode:** ~7.5 ns/op (string), ~0.3 ns/op (u32)
- **Decode:** ~37 ns/op (string), ~0.3 ns/op (u32)
- **Use when:** Performance critical, IPC, audio/video

### Streaming (Trait API)
- **Encode:** ~30 ns/op (string), ~1 ns/op (u32)
- **Decode:** ~38 ns/op (string), ~0.3 ns/op (u32)
- **Use when:** Files, sockets, compression layers

### Message Mode (Trait API + Headers)
- **Encode:** Streaming + 10 bytes overhead
- **Decode:** Streaming + type dispatch
- **Use when:** Multiple message types, storage, protocols

## Recommendations

1. **Default to slice API for byte mode**
   - Matches SDP's "high performance IPC" mission
   - 4x faster encoding
   - Most users want this for audio plugins

2. **Provide trait API for convenience**
   - Not everyone needs maximum performance
   - Essential for streaming use cases
   - Better Rust ecosystem integration

3. **Message mode always uses trait API**
   - Already has 10-byte header overhead
   - Trait overhead (3-4ns) is negligible
   - Needs to compose with io::Write for headers

4. **Make it obvious in docs**
   - Show performance table in generated code docs
   - Recommend slice API for hot paths
   - Explain trade-offs clearly

## Migration Path

1. **Phase 1:** Add slice API alongside existing trait API
2. **Phase 2:** Update benchmarks to use slice API
3. **Phase 3:** Re-run Go vs Rust benchmarks
4. **Phase 4:** Document performance recommendations
5. **Phase 5:** Consider making slice API the default

Expected result: **Rust matches or beats Go** for byte mode! 🚀
