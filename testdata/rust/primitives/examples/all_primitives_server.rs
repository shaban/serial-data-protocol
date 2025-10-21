//! Cross-platform test helper for all_primitives
//! Usage:
//!   cargo run --release --example crossplatform_helper encode-all_primitives > output.bin
//!   cargo run --release --example crossplatform_helper decode-all_primitives input.bin

use std::env;
use std::fs;
use std::io::{self, Write};

use sdp_all_primitives::*;

// Type alias to avoid conflict with wire format Result
type StdResult<T> = std::result::Result<T, Box<dyn std::error::Error>>;

fn encode_all_primitives() -> StdResult<()> {
    let data = AllPrimitives {
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

fn decode_all_primitives(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = AllPrimitives::decode_from_slice(&file_data)?;

    // Basic validation
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

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn main() -> StdResult<()> {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        eprintln!("Usage: {} <command> [args]", args[0]);
        eprintln!("Commands:");
        eprintln!("  encode-all_primitives - Encode AllPrimitives to stdout");
        eprintln!("  decode-all_primitives <file> - Decode AllPrimitives from file");
        std::process::exit(1);
    }

    match args[1].as_str() {
        "encode-all_primitives" => encode_all_primitives()?,
        "decode-all_primitives" => {
            if args.len() < 3 {
                eprintln!("Error: decode-all_primitives requires filename argument");
                std::process::exit(1);
            }
            decode_all_primitives(&args[2])?;
        }
        cmd => {
            eprintln!("Unknown command: {}", cmd);
            std::process::exit(1);
        }
    }

    Ok(())
}
