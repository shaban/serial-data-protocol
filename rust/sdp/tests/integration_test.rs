//! Integration tests for Rust-generated SDP code
//!
//! Tests roundtrip encoding/decoding of generated types.

// Include generated code from testdata
// Paths are relative to the Cargo workspace root (rust/)
#[path = "../../../testdata/primitives/rust/lib.rs"]
mod primitives;

#[path = "../../../testdata/audiounit/rust/lib.rs"]
mod audiounit;

#[path = "../../../testdata/arrays/rust/lib.rs"]
mod arrays;

#[path = "../../../testdata/optional/rust/lib.rs"]
mod optional;

#[test]
fn test_primitives_roundtrip() {
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
        str_field: "Hello, SDP!".to_string(),
    };

    // Encode
    let mut buf = Vec::new();
    original.encode(&mut buf).expect("encode failed");

    // Decode
    let decoded = primitives::AllPrimitives::decode(&mut &buf[..]).expect("decode failed");

    // Verify
    assert_eq!(decoded, original);
}

#[test]
fn test_audiounit_roundtrip() {
    let original = audiounit::PluginRegistry {
        plugins: vec![
            audiounit::Plugin {
                name: "Reverb".to_string(),
                manufacturer_id: "APPL".to_string(),
                component_type: "aufx".to_string(),
                component_subtype: "rvb ".to_string(),
                parameters: vec![
                    audiounit::Parameter {
                        address: 1000,
                        display_name: "Mix".to_string(),
                        identifier: "mix".to_string(),
                        unit: "%".to_string(),
                        min_value: 0.0,
                        max_value: 100.0,
                        default_value: 50.0,
                        current_value: 75.0,
                        raw_flags: 0x0001,
                        is_writable: true,
                        can_ramp: true,
                    },
                    audiounit::Parameter {
                        address: 1001,
                        display_name: "Room Size".to_string(),
                        identifier: "room_size".to_string(),
                        unit: "m".to_string(),
                        min_value: 1.0,
                        max_value: 100.0,
                        default_value: 10.0,
                        current_value: 25.0,
                        raw_flags: 0x0001,
                        is_writable: true,
                        can_ramp: true,
                    },
                ],
            },
        ],
        total_plugin_count: 1,
        total_parameter_count: 2,
    };

    // Encode
    let mut buf = Vec::new();
    original.encode(&mut buf).expect("encode failed");

    println!("Encoded {} bytes", buf.len());

    // Decode
    let decoded = audiounit::PluginRegistry::decode(&mut &buf[..]).expect("decode failed");

    // Verify
    assert_eq!(decoded.plugins.len(), 1);
    assert_eq!(decoded.plugins[0].name, "Reverb");
    assert_eq!(decoded.plugins[0].parameters.len(), 2);
    assert_eq!(decoded.plugins[0].parameters[0].display_name, "Mix");
    assert_eq!(decoded.total_plugin_count, 1);
    assert_eq!(decoded.total_parameter_count, 2);
}

#[test]
fn test_arrays_roundtrip() {
    let original = arrays::ArraysOfPrimitives {
        u8_array: vec![1, 2, 3, 4, 5],
        u32_array: vec![100, 200, 300],
        f64_array: vec![1.1, 2.2, 3.3],
        str_array: vec!["one".to_string(), "two".to_string(), "three".to_string()],
        bool_array: vec![true, false, true],
    };

    // Encode
    let mut buf = Vec::new();
    original.encode(&mut buf).expect("encode failed");

    // Decode
    let decoded = arrays::ArraysOfPrimitives::decode(&mut &buf[..]).expect("decode failed");

    // Verify
    assert_eq!(decoded, original);
}

#[test]
fn test_optional_fields() {
    // Test with Some values
    let with_metadata = optional::Request {
        id: 42,
        metadata: Some(optional::Metadata {
            user_id: 1001,
            username: "alice".to_string(),
        }),
    };

    let mut buf = Vec::new();
    with_metadata.encode(&mut buf).expect("encode failed");
    let decoded = optional::Request::decode(&mut &buf[..]).expect("decode failed");
    assert_eq!(decoded, with_metadata);

    // Test with None values
    let without_metadata = optional::Request {
        id: 100,
        metadata: None,
    };

    let mut buf2 = Vec::new();
    without_metadata.encode(&mut buf2).expect("encode failed");
    let decoded2 = optional::Request::decode(&mut &buf2[..]).expect("decode failed");
    assert_eq!(decoded2, without_metadata);

    // Verify size difference (None should be smaller)
    assert!(buf2.len() < buf.len(), "None encoding should be smaller");
}

#[test]
fn test_empty_arrays() {
    let empty = arrays::ArraysOfPrimitives {
        u8_array: vec![],
        u32_array: vec![],
        f64_array: vec![],
        str_array: vec![],
        bool_array: vec![],
    };

    let mut buf = Vec::new();
    empty.encode(&mut buf).expect("encode failed");

    // Should only encode 5 × u32(0) for lengths = 20 bytes
    assert_eq!(buf.len(), 20);

    let decoded = arrays::ArraysOfPrimitives::decode(&mut &buf[..]).expect("decode failed");
    assert_eq!(decoded, empty);
}
