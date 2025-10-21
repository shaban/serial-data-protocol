//! Cross-platform test helper for arrays_of_primitives
//! Usage:
//!   cargo run --release --example crossplatform_helper encode-arrays_of_primitives > output.bin
//!   cargo run --release --example crossplatform_helper decode-arrays_of_primitives input.bin
//!   cargo run --release --example crossplatform_helper encode-item > output.bin
//!   cargo run --release --example crossplatform_helper decode-item input.bin
//!   cargo run --release --example crossplatform_helper encode-arrays_of_structs > output.bin
//!   cargo run --release --example crossplatform_helper decode-arrays_of_structs input.bin

use std::env;
use std::fs;
use std::io::{self, Write};

use sdp_arrays_of_primitives::*;

// Type alias to avoid conflict with wire format Result
type StdResult<T> = std::result::Result<T, Box<dyn std::error::Error>>;

fn encode_arrays_of_primitives() -> StdResult<()> {
    let data = ArraysOfPrimitives {
        u8_array: vec![255, 255, 255],
        u32_array: vec![4_294_967_295, 4_294_967_295, 4_294_967_295],
        f64_array: vec![2.718281828459045, 2.718281828459045, 2.718281828459045],
        str_array: vec!["Hello from Rust!".to_string(), "Hello from Rust!".to_string(), "Hello from Rust!".to_string()],
        bool_array: vec![true, true, true],
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_arrays_of_primitives(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = ArraysOfPrimitives::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.u8_array.len() > 0;
    ok = ok && decoded.u32_array.len() > 0;
    ok = ok && decoded.f64_array.len() > 0;
    ok = ok && decoded.str_array.len() > 0;
    ok = ok && decoded.bool_array.len() > 0;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_item() -> StdResult<()> {
    let data = Item {
        id: 4_294_967_295,
        name: "Hello from Rust!".to_string(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_item(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = Item::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.id == 4_294_967_295;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_arrays_of_structs() -> StdResult<()> {
    let data = ArraysOfStructs {
        items: vec![Default::default(), Default::default(), Default::default()],
        count: 4_294_967_295,
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_arrays_of_structs(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = ArraysOfStructs::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.items.len() > 0;
    ok = ok && decoded.count == 4_294_967_295;

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
        eprintln!("  encode-arrays_of_primitives - Encode ArraysOfPrimitives to stdout");
        eprintln!("  decode-arrays_of_primitives <file> - Decode ArraysOfPrimitives from file");
        eprintln!("  encode-item - Encode Item to stdout");
        eprintln!("  decode-item <file> - Decode Item from file");
        eprintln!("  encode-arrays_of_structs - Encode ArraysOfStructs to stdout");
        eprintln!("  decode-arrays_of_structs <file> - Decode ArraysOfStructs from file");
        std::process::exit(1);
    }

    match args[1].as_str() {
        "encode-arrays_of_primitives" => encode_arrays_of_primitives()?,
        "decode-arrays_of_primitives" => {
            if args.len() < 3 {
                eprintln!("Error: decode-arrays_of_primitives requires filename argument");
                std::process::exit(1);
            }
            decode_arrays_of_primitives(&args[2])?;
        }
        "encode-item" => encode_item()?,
        "decode-item" => {
            if args.len() < 3 {
                eprintln!("Error: decode-item requires filename argument");
                std::process::exit(1);
            }
            decode_item(&args[2])?;
        }
        "encode-arrays_of_structs" => encode_arrays_of_structs()?,
        "decode-arrays_of_structs" => {
            if args.len() < 3 {
                eprintln!("Error: decode-arrays_of_structs requires filename argument");
                std::process::exit(1);
            }
            decode_arrays_of_structs(&args[2])?;
        }
        cmd => {
            eprintln!("Unknown command: {}", cmd);
            std::process::exit(1);
        }
    }

    Ok(())
}
