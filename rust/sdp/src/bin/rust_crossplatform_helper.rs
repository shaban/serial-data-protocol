//! Cross-platform test helper for Rust
//! Usage:
//!   rust_crossplatform_helper encode-primitives > output.bin
//!   rust_crossplatform_helper decode-primitives input.bin

use std::env;
use std::fs;
use std::io::{self, Write};

#[path = "../../../../testdata/primitives/rust/lib.rs"]
mod primitives;

fn encode_primitives() -> Result<(), Box<dyn std::error::Error>> {
    let data = primitives::AllPrimitives {
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

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_primitives(filename: &str) -> Result<(), Box<dyn std::error::Error>> {
    let file_data = fs::read(filename)?;

    let decoded = primitives::AllPrimitives::decode_from_slice(&file_data)?;

    // Verify values
    let mut ok = true;
    ok = ok && decoded.u8_field == 255;
    ok = ok && decoded.u16_field == 65535;
    ok = ok && decoded.u32_field == 4_294_967_295;
    ok = ok && decoded.u64_field == 18_446_744_073_709_551_615;
    ok = ok && decoded.i8_field == -128;
    ok = ok && decoded.i16_field == -32768;
    ok = ok && decoded.i32_field == -2_147_483_648;
    ok = ok && decoded.i64_field == -9_223_372_036_854_775_808;
    ok = ok && (decoded.f32_field - 3.14159).abs() < 0.0001;
    ok = ok && (decoded.f64_field - 2.718281828459045).abs() < 0.0000001;
    ok = ok && decoded.bool_field == true;
    // Accept from Go or Swift
    ok = ok && (decoded.str_field == "Hello from Go!" || decoded.str_field == "Hello from Swift!");

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        eprintln!("Usage: {} <command> [args]", args[0]);
        eprintln!("Commands:");
        eprintln!("  encode-primitives - Encode primitives and output binary to stdout");
        eprintln!("  decode-primitives <file> - Decode primitives from file");
        std::process::exit(1);
    }

    match args[1].as_str() {
        "encode-primitives" => encode_primitives()?,
        "decode-primitives" => {
            if args.len() < 3 {
                eprintln!("Error: decode-primitives requires filename argument");
                std::process::exit(1);
            }
            decode_primitives(&args[2])?;
        }
        cmd => {
            eprintln!("Unknown command: {}", cmd);
            std::process::exit(1);
        }
    }

    Ok(())
}
