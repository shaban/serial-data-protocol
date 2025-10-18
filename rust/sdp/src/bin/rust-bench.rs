// Cross-language benchmark helper
// This binary is called by Go benchmarks to measure Rust encode/decode performance
// Usage:
//   rust-bench encode-primitives <iterations>
//   rust-bench decode-primitives <file> <iterations>

use std::env;
use std::fs;
use std::time::Instant;

#[path = "../../../../testdata/primitives/rust/lib.rs"]
mod primitives;

#[path = "../../../../testdata/audiounit/rust/lib.rs"]
mod audiounit;

use primitives::AllPrimitives;
use audiounit::{PluginRegistry, Plugin, Parameter};

fn create_test_primitives() -> AllPrimitives {
    AllPrimitives {
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
    }
}

fn create_test_audiounit() -> PluginRegistry {
    let param1 = Parameter {
        address: 0,
        display_name: "Volume".to_string(),
        identifier: "volume".to_string(),
        unit: "dB".to_string(),
        min_value: -96.0,
        max_value: 6.0,
        default_value: 0.0,
        current_value: -12.0,
        raw_flags: 0x01,
        is_writable: true,
        can_ramp: true,
    };
    
    let param2 = Parameter {
        address: 1,
        display_name: "Pan".to_string(),
        identifier: "pan".to_string(),
        unit: "%".to_string(),
        min_value: -100.0,
        max_value: 100.0,
        default_value: 0.0,
        current_value: 0.0,
        raw_flags: 0x01,
        is_writable: true,
        can_ramp: true,
    };
    
    let plugin1 = Plugin {
        name: "Test Synth".to_string(),
        manufacturer_id: "TEST".to_string(),
        component_type: "aumu".to_string(),
        component_subtype: "test".to_string(),
        parameters: vec![param1.clone(), param2.clone()],
    };
    
    let plugin2 = Plugin {
        name: "Test Effect".to_string(),
        manufacturer_id: "TEST".to_string(),
        component_type: "aumf".to_string(),
        component_subtype: "tsfx".to_string(),
        parameters: vec![param1.clone()],
    };
    
    PluginRegistry {
        plugins: vec![plugin1, plugin2],
        total_plugin_count: 2,
        total_parameter_count: 3,
    }
}

fn bench_encode_primitives(iterations: usize) {
    let data = create_test_primitives();
    
    let start = Instant::now();
    for _ in 0..iterations {
        let mut buf = vec![0u8; data.encoded_size()];
        data.encode_to_slice(&mut buf).unwrap();
        // Prevent optimization from removing the encode
        std::hint::black_box(buf);
    }
    let duration = start.elapsed();
    
    // Output timing in nanoseconds per operation
    let ns_per_op = duration.as_nanos() / iterations as u128;
    println!("{}", ns_per_op);
}

fn bench_decode_primitives(filename: &str, iterations: usize) {
    let data = fs::read(filename).expect("Failed to read file");
    
    let start = Instant::now();
    for _ in 0..iterations {
        let decoded = AllPrimitives::decode_from_slice(&data).unwrap();
        std::hint::black_box(decoded);
    }
    let duration = start.elapsed();
    
    let ns_per_op = duration.as_nanos() / iterations as u128;
    println!("{}", ns_per_op);
}

fn bench_encode_audiounit(iterations: usize) {
    let data = create_test_audiounit();
    
    let start = Instant::now();
    for _ in 0..iterations {
        let mut buf = vec![0u8; data.encoded_size()];
        data.encode_to_slice(&mut buf).unwrap();
        std::hint::black_box(buf);
    }
    let duration = start.elapsed();
    
    let ns_per_op = duration.as_nanos() / iterations as u128;
    println!("{}", ns_per_op);
}

fn bench_decode_audiounit(filename: &str, iterations: usize) {
    let data = fs::read(filename).expect("Failed to read file");
    
    let start = Instant::now();
    for _ in 0..iterations {
        let decoded = PluginRegistry::decode_from_slice(&data).unwrap();
        std::hint::black_box(decoded);
    }
    let duration = start.elapsed();
    
    let ns_per_op = duration.as_nanos() / iterations as u128;
    println!("{}", ns_per_op);
}

fn main() {
    let args: Vec<String> = env::args().collect();
    
    if args.len() < 2 {
        eprintln!("Usage: {} <command> [args]", args[0]);
        eprintln!("Commands:");
        eprintln!("  encode-primitives <iterations>");
        eprintln!("  decode-primitives <file> <iterations>");
        eprintln!("  encode-audiounit <iterations>");
        eprintln!("  decode-audiounit <file> <iterations>");
        std::process::exit(1);
    }
    
    let command = &args[1];
    
    match command.as_str() {
        "encode-primitives" => {
            let iterations = args[2].parse().expect("Invalid iterations");
            bench_encode_primitives(iterations);
        }
        "decode-primitives" => {
            let filename = &args[2];
            let iterations = args[3].parse().expect("Invalid iterations");
            bench_decode_primitives(filename, iterations);
        }
        "encode-audiounit" => {
            let iterations = args[2].parse().expect("Invalid iterations");
            bench_encode_audiounit(iterations);
        }
        "decode-audiounit" => {
            let filename = &args[2];
            let iterations = args[3].parse().expect("Invalid iterations");
            bench_decode_audiounit(filename, iterations);
        }
        _ => {
            eprintln!("Unknown command: {}", command);
            std::process::exit(1);
        }
    }
}
