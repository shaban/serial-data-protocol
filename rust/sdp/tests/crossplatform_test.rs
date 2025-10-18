//! Cross-language wire format compatibility tests
//!
//! Tests that Go-encoded data can be decoded by Rust and vice versa.
//! This validates that the wire format is truly language-agnostic.

use std::process::Command;
use std::fs;
use std::path::Path;

#[path = "../../../testdata/primitives/rust/lib.rs"]
mod primitives;

#[path = "../../../testdata/audiounit/rust/lib.rs"]
mod audiounit;

/// Helper: Run Go code to encode a message and return the binary output
fn go_encode_primitives() -> Vec<u8> {
    // Use Go to encode a primitives message (runs from workspace root)
    let output = Command::new("go")
        .args(&["run", "testdata/crossplatform_helper.go", "encode-primitives"])
        .current_dir("../..")
        .output()
        .expect("Failed to run Go encoder");
    
    assert!(output.status.success(), "Go encoder failed: {}", String::from_utf8_lossy(&output.stderr));
    output.stdout
}

/// Helper: Send Rust-encoded data to Go for decoding
fn go_decode_primitives(data: &[u8]) -> bool {
    let temp_file = "/tmp/rust_encoded.bin";
    fs::write(temp_file, data).expect("Failed to write temp file");
    
    let output = Command::new("go")
        .args(&["run", "testdata/crossplatform_helper.go", "decode-primitives", temp_file])
        .current_dir("../..")
        .output()
        .expect("Failed to run Go decoder");
    
    output.status.success()
}

#[test]
fn test_go_to_rust_primitives() {
    // Get Go-encoded data
    let go_data = go_encode_primitives();
    
    println!("Go encoded {} bytes", go_data.len());
    
    // Decode with Rust
    let decoded = primitives::AllPrimitives::decode_from_slice(&go_data)
        .expect("Rust failed to decode Go-encoded data");
    
    // Verify values (these match what Go encodes)
    assert_eq!(decoded.u8_field, 255);
    assert_eq!(decoded.u16_field, 65535);
    assert_eq!(decoded.u32_field, 4_294_967_295);
    assert_eq!(decoded.u64_field, 18_446_744_073_709_551_615);
    assert_eq!(decoded.i8_field, -128);
    assert_eq!(decoded.i16_field, -32768);
    assert_eq!(decoded.i32_field, -2_147_483_648);
    assert_eq!(decoded.i64_field, -9_223_372_036_854_775_808);
    assert!((decoded.f32_field - 3.14159).abs() < 0.0001);
    assert!((decoded.f64_field - 2.718281828459045).abs() < 0.0000001);
    assert_eq!(decoded.bool_field, true);
    assert_eq!(decoded.str_field, "Hello from Go!");
}

#[test]
fn test_rust_to_go_primitives() {
    // Create data in Rust
    let original = primitives::AllPrimitives {
        u8_field: 255,
        u16_field: 65535,
        u32_field: 4_294_967_295,
        u64_field: 18_446_744_073_709_551_615,
        i8_field: -128,
        i16_field: -32768,
        i32_field: -2_147_483_648,
        i64_field: -9_223_372_036_854_775_808,
        f32_field: 3.14159,
        f64_field: 2.718281828459045,
        bool_field: true,
        str_field: "Hello from Rust!".to_string(),
    };
    
    // Encode with Rust
    let mut buf = vec![0u8; original.encoded_size()];
    original.encode_to_slice(&mut buf).expect("Rust encode failed");
    
    println!("Rust encoded {} bytes", buf.len());
    
    // Send to Go for decoding
    let go_success = go_decode_primitives(&buf);
    assert!(go_success, "Go failed to decode Rust-encoded data");
}

#[test]
fn test_wire_format_is_identical() {
    // Create same data in Rust
    let rust_data = primitives::AllPrimitives {
        u8_field: 255,
        u16_field: 65535,
        u32_field: 4_294_967_295,
        u64_field: 18_446_744_073_709_551_615,
        i8_field: -128,
        i16_field: -32768,
        i32_field: -2_147_483_648,
        i64_field: -9_223_372_036_854_775_808,
        f32_field: 3.14159,
        f64_field: 2.718281828459045,
        bool_field: true,
        str_field: "Hello from Go!".to_string(), // Same string as Go
    };
    
    // Encode with Rust
    let mut rust_buf = vec![0u8; rust_data.encoded_size()];
    rust_data.encode_to_slice(&mut rust_buf).expect("Rust encode failed");
    
    // Get Go-encoded data
    let go_buf = go_encode_primitives();
    
    // Wire formats should be IDENTICAL
    assert_eq!(rust_buf.len(), go_buf.len(), "Buffer lengths don't match");
    assert_eq!(rust_buf, go_buf, "Wire formats are not identical!");
    
    println!("✓ Go and Rust produce identical wire format ({} bytes)", rust_buf.len());
}
