# Byte Mode Safety Guide

## TL;DR

**Byte mode = IPC only. Message mode = everything else.**

---

## The Problem

Byte mode has **NO magic bytes, NO version info, NO type identification.**

```
Byte mode wire format:
[field1][field2][field3]...

That's it. Just raw data.
```

This is intentional for **performance**, but means:
- ❌ Can't detect file corruption
- ❌ Can't validate message type
- ❌ Can't check schema version
- ❌ Can't catch accidental wrong decoder

---

## When Byte Mode is Safe ✅

### ✅ IPC (Inter-Process Communication)

**Requirements**:
1. Both sides import the same generated package
2. Data lifetime is ephemeral (microseconds to milliseconds)
3. Single codebase (guaranteed version match)

**Example - Safe IPC**:
```go
// process_a.go
import "myapp/generated/audiounit"

func sendToProcessB(registry *audiounit.PluginRegistry) {
    data := audiounit.EncodePluginRegistry(registry)  // Byte mode
    ipcChannel.Send(data)  // Same machine, ephemeral
}

// process_b.go  
import "myapp/generated/audiounit"  // Same import!

func receiveFromProcessA() {
    data := ipcChannel.Receive()
    var registry audiounit.PluginRegistry
    audiounit.DecodePluginRegistry(&registry, data)  // Safe!
    // Type safety from shared import, same schema version guaranteed
}
```

**Why it's safe**:
- Both processes compiled from same codebase
- Both import same generated package → same schema version
- Data never persisted → no corruption risk
- Type safety from Go type system

---

## When Byte Mode is Dangerous ❌

### ❌ Persistence (Files, Databases)

**Problem**: No magic bytes to detect corruption or wrong file type.

```go
// ❌ DANGER: Save byte mode to disk
data := audiounit.EncodePluginRegistry(&registry)
os.WriteFile("plugins.bin", data, 0644)

// Months later, file is corrupted
// Or someone tries to decode with wrong type
var plugin audiounit.Plugin  // Wrong type!
audiounit.DecodePlugin(&plugin, data)  // 💥 Silent corruption or panic

// No magic bytes to catch the mistake!
```

**What can go wrong**:
1. **File corruption** - Bit flips, disk errors → undetected
2. **Wrong decoder** - Someone uses `DecodePlugin` instead of `DecodePluginRegistry`
3. **Schema evolution** - Old file, new decoder → undefined behavior
4. **Wrong file type** - Accidentally read a JPEG as SDP → garbage data

### ❌ Network Communication

**Problem**: No validation that received data is actually SDP.

```go
// ❌ DANGER: Send byte mode over network
data := audiounit.EncodePluginRegistry(&registry)
conn.Write(data)

// Receiver has no way to validate:
// - Is this SDP data?
// - What type is it?
// - What version?
// - Is it corrupted?

var decoded audiounit.PluginRegistry
audiounit.DecodePluginRegistry(&decoded, data)  // Hope for the best!
```

**What can go wrong**:
1. **Network corruption** - Packet loss, bit errors → undetected
2. **Version mismatch** - Sender has v1 schema, receiver has v2 → crash
3. **Type confusion** - Router dispatches to wrong handler
4. **Injection attacks** - Attacker sends malformed data

### ❌ Cross-Language Communication

**Problem**: Different languages have different generated packages.

```go
// Go service (generated from schema v1)
data := audiounit.EncodePluginRegistry(&registry)
sendToRustService(data)

// Rust service (generated from schema v2 - oops!)
let decoded = decode_plugin_registry(&data)?;  // 💥 Schema mismatch!

// No header to detect version difference!
```

---

## The Solution: Message Mode ✅

**Message mode adds a 10-30 byte header** with:
1. **Magic bytes** (`"SDP"`) - Detect corruption, validate file type
2. **Version number** - Enable forward/backward compatibility
3. **Type name** - Validate correct decoder before decoding
4. **Payload length** - Frame messages, validate completeness

### ✅ Safe Persistence

```go
// ✅ SAFE: Use message mode for files
data := audiounit.EncodePluginRegistryMessage(&registry)
os.WriteFile("plugins.bin", data, 0644)

// Later, even with corruption or wrong type:
decoded, err := audiounit.DecodeMessage(data)
if err != nil {
    // Catches:
    // - File corruption (bad magic bytes)
    // - Wrong file type (not SDP)
    // - Version mismatch (unsupported version)
    // - Type mismatch (tried to decode as Plugin)
    log.Fatal("Invalid SDP file:", err)
}
registry := decoded.(*audiounit.PluginRegistry)  // Type-safe cast
```

### ✅ Safe Network Communication

```go
// ✅ SAFE: Use message mode for network
data := audiounit.EncodePluginRegistryMessage(&registry)
conn.Write(data)

// Receiver validates before decoding
decoded, err := audiounit.DecodeMessage(data)
if err != nil {
    // Catches all the problems from above
    return err
}

// Dispatcher routes based on type
switch v := decoded.(type) {
case *audiounit.PluginRegistry:
    handleRegistry(v)
case *audiounit.Plugin:
    handlePlugin(v)
default:
    return fmt.Errorf("unknown type: %T", v)
}
```

---

## Decision Matrix

| Use Case | Byte Mode? | Message Mode? | Reason |
|----------|------------|---------------|--------|
| **IPC (same machine)** | ✅ Yes | ⚠️ Overkill | Type safety from imports |
| **IPC (different machines)** | ⚠️ Risky | ✅ Yes | Need version validation |
| **File storage** | ❌ No | ✅ Yes | Need corruption detection |
| **Network (same service)** | ⚠️ Maybe | ✅ Safer | Consider version drift |
| **Network (different services)** | ❌ No | ✅ Yes | Need type validation |
| **Cross-language** | ❌ No | ✅ Yes | Different schema versions |
| **Long-term storage** | ❌ No | ✅ Yes | Schema will evolve |
| **Hot path (performance)** | ✅ Yes | ❌ Overkill | If IPC-only |

---

## Real-World Example: AudioUnit Plugins

### ✅ Good: Byte Mode for IPC

```go
// Host process and plugin process on same Mac
// Both compiled from same codebase

// Host → Plugin (parameter change)
data := audiounit.EncodeParameterUpdate(&update)  // Byte mode, 50 bytes
ipcChannel.Send(data)  // Microsecond latency critical

// Plugin receives and decodes
var update audiounit.ParameterUpdate
audiounit.DecodeParameterUpdate(&update, data)  // ~2µs decode time
```

**Safe because**:
- Same machine, same build
- Ephemeral (parameter lasts milliseconds)
- Type safety from shared import
- Performance critical (real-time audio)

### ❌ Bad: Byte Mode for Presets

```go
// ❌ DON'T DO THIS
// User saves plugin preset to disk
data := audiounit.EncodePluginState(&state)  // Byte mode
os.WriteFile("MyReverb.preset", data, 0644)

// Months later, after plugin update with new schema:
data, _ := os.ReadFile("MyReverb.preset")
var state audiounit.PluginState  // New schema version!
audiounit.DecodePluginState(&state, data)  // 💥 Crash or corruption
```

### ✅ Good: Message Mode for Presets

```go
// ✅ DO THIS
// User saves plugin preset with versioning
data := audiounit.EncodePluginStateMessage(&state)  // Message mode
os.WriteFile("MyReverb.preset", data, 0644)

// Months later, even with schema changes:
decoded, err := audiounit.DecodeMessage(data)
if err != nil {
    if err == ErrUnsupportedVersion {
        return fmt.Errorf("preset from old plugin version, please upgrade")
    }
    return err
}
state := decoded.(*audiounit.PluginState)  // Works!
```

---

## Performance Impact

**Message mode overhead**:
```
Small message (100 bytes):  ~20 bytes = 20% overhead
Medium message (10 KB):     ~24 bytes = 0.24% overhead  
Large message (110 KB):     ~24 bytes = 0.02% overhead

Speed: +5-10% slower (header parsing)
```

**Is it worth it?**
- IPC hot path: No, use byte mode
- Everything else: Yes, use message mode

---

## Summary

**Rule of thumb**:

```
If data leaves the process → Message mode
If data stays in process  → Byte mode (maybe)
If unsure                → Message mode (safer)
```

**Exceptions**:
- Embedded systems with <1KB RAM: Byte mode acceptable
- Real-time audio hot paths: Byte mode for latency
- High-frequency trading: Byte mode for nanoseconds

**Default recommendation**: **Use message mode unless you have a specific reason not to.**

The 0.02% overhead is insurance against:
- Silent data corruption
- Schema evolution bugs  
- Type confusion crashes
- Security vulnerabilities

---

## Checklist

Before using byte mode, verify:

- [ ] Both encoder/decoder import same generated package?
- [ ] Both from same Git commit / build?
- [ ] Data lifetime < 1 second?
- [ ] Never persisted to disk/database?
- [ ] Never sent over network?
- [ ] Never crosses language boundary?
- [ ] Performance is actually critical? (measured, not assumed)

If any answer is "no" or "maybe" → **Use message mode.**

---

**Last Updated**: October 18, 2025  
**SDP Version**: 0.2.0-rc1
