# SDP Language Implementation Quick Reference

**Use this for tomorrow's C and Swift implementations**

---

## 🎯 Implementation Order (4 hours each)

### Hour 1: Setup + Types
- [ ] Create `internal/generator/{lang}/` directory structure
- [ ] Implement type generator (structs → native types)
- [ ] Write 10 unit tests for type generation
- [ ] Verify generated code compiles

### Hour 2: Encoder
- [ ] Implement size calculator generator
- [ ] Implement encoder generator (all types)
- [ ] Write 10 unit tests for encoder generation
- [ ] Verify generated encoders compile

### Hour 3: Decoder + Errors
- [ ] Implement decoder generator (all types)
- [ ] Implement error types generator
- [ ] Write 10 unit tests for decoder generation
- [ ] Verify generated decoders compile

### Hour 4: Integration + Performance
- [ ] Set up TestMain with test schemas
- [ ] Add 11 wire format tests
- [ ] Add 8 roundtrip tests
- [ ] Add performance benchmarks
- [ ] Test with plugins.json

---

## 📝 Wire Format Cheat Sheet

| Type | Size | Encode | Decode |
|------|------|--------|--------|
| `u8` | 1 | `buf[off] = val` | `val = data[off]` |
| `u16` | 2 | Write LE uint16 | Read LE uint16 |
| `u32` | 4 | Write LE uint32 | Read LE uint32 |
| `u64` | 8 | Write LE uint64 | Read LE uint64 |
| `i8` | 1 | `buf[off] = (u8)val` | `val = (i8)data[off]` |
| `i16` | 2 | Write LE, cast from i16 | Read LE, cast to i16 |
| `i32` | 4 | Write LE, cast from i32 | Read LE, cast to i32 |
| `i64` | 8 | Write LE, cast from i64 | Read LE, cast to i64 |
| `f32` | 4 | Write bits as LE u32 | Read LE u32 as bits |
| `f64` | 8 | Write bits as LE u64 | Read LE u64 as bits |
| `bool` | 1 | `buf[off] = val?1:0` | `val = data[off]!=0` |
| `str` | Var | `[u32 len][UTF-8]` | Read len, read bytes |
| `[]T` | Var | `[u32 count][elems]` | Read count, decode each |

---

## 🔧 Encoder Pattern (Copy-Paste Template)

```
Public API:
    1. Calculate size
    2. Allocate buffer
    3. Call helper encoder
    4. Return buffer

Helper Encoder:
    For each field:
        1. Write to buffer at offset
        2. Increment offset by field size
```

---

## 🔍 Decoder Pattern (Copy-Paste Template)

```
Public API:
    1. Create DecodeContext
    2. Call helper decoder
    3. Return error or success

Helper Decoder:
    For each field:
        1. Check bounds (offset + size <= len)
        2. Read from buffer at offset
        3. Increment offset by field size
        4. Check array limits if needed
```

---

## ⚠️ Must-Have Error Checks

```
Decode every field:
    if (offset + SIZE > len) return ErrUnexpectedEOF;

Decode array:
    if (count > MaxArrayElements) return ErrArrayTooLarge;
    ctx.total += count;
    if (ctx.total > MaxTotalElements) return ErrTooManyElements;

Decode string:
    if (offset + 4 > len) return ErrUnexpectedEOF;
    len = read_u32_le(data + offset);
    if (offset + 4 + len > data_len) return ErrUnexpectedEOF;
```

---

## 🎯 Performance Targets

| Test | Target (M1 Base) |
|------|------------------|
| Simple struct (2 fields) | < 30 ns encode, < 25 ns decode |
| Nested struct (3 levels) | < 25 ns encode, < 20 ns decode |
| Small array (5 elements) | < 60 ns encode, < 140 ns decode |
| Real-world (1,759 params) | < 150 µs roundtrip |

**Must beat:** Protocol Buffers (1,300 µs) by at least **8×**

---

## ✅ Test Checklist

Wire Format Tests (11):
- [ ] All 12 primitive types with hand-crafted binary
- [ ] Truncated data detection
- [ ] Invalid string length detection
- [ ] 3-level nested structs
- [ ] Primitive arrays
- [ ] Empty arrays
- [ ] Oversized array rejection
- [ ] Struct arrays

Roundtrip Tests (8):
- [ ] All primitives with max/min values + Unicode
- [ ] Empty string edge case
- [ ] Nested structs with negative floats
- [ ] All array types with Unicode
- [ ] Empty arrays
- [ ] Arrays of structs
- [ ] Complex nested + arrays
- [ ] Large data (1000 elements)

---

## 🚀 C-Specific Quick Notes

**Memory:**
```c
uint8_t* encode_X(X* src, uint32_t* out_size) {
    uint32_t size = calculate_X_size(src);
    uint8_t* buf = malloc(size);
    // ... encode ...
    *out_size = size;
    return buf;  // Caller must free()
}

int decode_X(X* dst, uint8_t* data, uint32_t len) {
    // ... decode ...
    return 0;  // 0 = success, error code otherwise
}
```

**Little-Endian Helpers:**
```c
void write_u32_le(uint8_t* buf, uint32_t v) {
    buf[0]=v; buf[1]=v>>8; buf[2]=v>>16; buf[3]=v>>24;
}
uint32_t read_u32_le(uint8_t* buf) {
    return buf[0]|(buf[1]<<8)|(buf[2]<<16)|(buf[3]<<24);
}
```

---

## 🚀 Swift-Specific Quick Notes

**Memory:**
```swift
func encodeX(_ src: X) throws -> Data {
    let size = calculateXSize(src)
    var data = Data(capacity: size)
    // ... encode ...
    return data
}

func decodeX(_ dst: inout X, from data: Data) throws {
    var offset = 0
    // ... decode ...
}
```

**Little-Endian Helpers:**
```swift
extension Data {
    mutating func appendLE(_ value: UInt32) {
        var v = value.littleEndian
        withUnsafeBytes(of: &v) { append(contentsOf: $0) }
    }
    func readU32LE(at offset: Int) -> UInt32 {
        withUnsafeBytes { 
            $0.load(fromByteOffset: offset, as: UInt32.self)
        }.littleEndian
    }
}
```

---

## 📊 Final Validation

Before calling it done:
- [ ] All 238+ tests pass
- [ ] Benchmarks show 8-10× faster than Protocol Buffers
- [ ] plugins.json roundtrip works perfectly
- [ ] No memory leaks (C: valgrind, Swift: Instruments)
- [ ] Cross-language compatibility (Go ↔ C ↔ Swift)

---

**Time Budget:**
- C: 4 hours (careful with memory)
- Swift: 3 hours (easier than C)
- Total: 7 hours → Both done in one day! 🎉

