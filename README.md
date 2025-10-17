# Serial Data Protocol

A binary serialization protocol optimized for same-machine IPC between languages via FFI.

## Overview

Serial Data Protocol (SDP) enables efficient bulk data transfer from low-level languages (C/Swift/Rust) to high-level orchestrators (Go) via a single FFI call. Designed for scenarios like audio plugin enumeration, device discovery, and configuration data transfer.

**Key Features:**
- Schema-driven code generation
- Flexible field write order (encoder side)
- Rust-like schema syntax with IDE support
- Comprehensive size limits and validation
- No runtime dependencies in generated code

## Documentation

**All project documentation is in five canonical documents:**

1. **[DESIGN_SPEC.md](DESIGN_SPEC.md)** - Complete technical specification
2. **[TESTING_STRATEGY.md](TESTING_STRATEGY.md)** - Testing approach and requirements
3. **[DOCUMENTATION_GUIDELINES.md](DOCUMENTATION_GUIDELINES.md)** - Documentation standards
4. **[IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)** - Step-by-step implementation tasks
5. **[CONSTITUTION.md](CONSTITUTION.md)** - Project governance and rules

## Quick Start

### 1. Define Schema

Create `device.sdp`:

```rust
/// Device represents an audio device.
struct Device {
    /// Unique device identifier.
    id: u32,
    
    /// Human-readable device name.
    name: str,
}
```

### 2. Generate Code

```bash
go run ./cmd/sdp-gen -schema device.sdp -output device -lang go
```

### 3. Use Generated Code

**Encoder (C):**
```c
device_builder_t* builder = NewDeviceBuilder();
SetDeviceID(builder, 42);
SetDeviceName(builder, "Built-in Microphone");
serial_data_t data = FinalizeBuilder(builder);
```

**Decoder (Go):**
```go
import "yourproject/device"

var d device.Device
if err := device.Decode(&d, wireData); err != nil {
    log.Fatal(err)
}
fmt.Printf("Device: %s (ID: %d)\n", d.Name, d.ID)
```

## Status

**Version:** 1.0.0 (In Development)

See [DESIGN_SPEC.md](DESIGN_SPEC.md) for complete specification.

## License

[Add your license here]

## Contributing

Read [CONSTITUTION.md](CONSTITUTION.md) for project governance and rules.

All contributions must:
- Update relevant canonical documents
- Include tests per [TESTING_STRATEGY.md](TESTING_STRATEGY.md)
- Follow [DOCUMENTATION_GUIDELINES.md](DOCUMENTATION_GUIDELINES.md)
