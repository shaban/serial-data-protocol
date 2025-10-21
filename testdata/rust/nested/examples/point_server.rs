//! Cross-platform test helper for point
//! Usage:
//!   cargo run --release --example crossplatform_helper encode-point > output.bin
//!   cargo run --release --example crossplatform_helper decode-point input.bin
//!   cargo run --release --example crossplatform_helper encode-rectangle > output.bin
//!   cargo run --release --example crossplatform_helper decode-rectangle input.bin
//!   cargo run --release --example crossplatform_helper encode-scene > output.bin
//!   cargo run --release --example crossplatform_helper decode-scene input.bin

use std::env;
use std::fs;
use std::io::{self, Write};

use sdp_point::*;

// Type alias to avoid conflict with wire format Result
type StdResult<T> = std::result::Result<T, Box<dyn std::error::Error>>;

fn encode_point() -> StdResult<()> {
    let data = Point {
        x: 3.14159,
        y: 3.14159,
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_point(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = Point::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && (decoded.x - 3.14159).abs() < 0.0001;
    ok = ok && (decoded.y - 3.14159).abs() < 0.0001;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_rectangle() -> StdResult<()> {
    let data = Rectangle {
        top_left: Default::default(),
        bottom_right: Default::default(),
        color: 4_294_967_295,
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_rectangle(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = Rectangle::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.color == 4_294_967_295;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_scene() -> StdResult<()> {
    let data = Scene {
        name: "Hello from Rust!".to_string(),
        main_rect: Default::default(),
        count: 4_294_967_295,
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_scene(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = Scene::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
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
        eprintln!("  encode-point - Encode Point to stdout");
        eprintln!("  decode-point <file> - Decode Point from file");
        eprintln!("  encode-rectangle - Encode Rectangle to stdout");
        eprintln!("  decode-rectangle <file> - Decode Rectangle from file");
        eprintln!("  encode-scene - Encode Scene to stdout");
        eprintln!("  decode-scene <file> - Decode Scene from file");
        std::process::exit(1);
    }

    match args[1].as_str() {
        "encode-point" => encode_point()?,
        "decode-point" => {
            if args.len() < 3 {
                eprintln!("Error: decode-point requires filename argument");
                std::process::exit(1);
            }
            decode_point(&args[2])?;
        }
        "encode-rectangle" => encode_rectangle()?,
        "decode-rectangle" => {
            if args.len() < 3 {
                eprintln!("Error: decode-rectangle requires filename argument");
                std::process::exit(1);
            }
            decode_rectangle(&args[2])?;
        }
        "encode-scene" => encode_scene()?,
        "decode-scene" => {
            if args.len() < 3 {
                eprintln!("Error: decode-scene requires filename argument");
                std::process::exit(1);
            }
            decode_scene(&args[2])?;
        }
        cmd => {
            eprintln!("Unknown command: {}", cmd);
            std::process::exit(1);
        }
    }

    Ok(())
}
