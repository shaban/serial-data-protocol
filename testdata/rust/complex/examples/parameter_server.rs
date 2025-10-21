//! Cross-platform test helper for parameter
//! Usage:
//!   cargo run --release --example crossplatform_helper encode-parameter > output.bin
//!   cargo run --release --example crossplatform_helper decode-parameter input.bin
//!   cargo run --release --example crossplatform_helper encode-plugin > output.bin
//!   cargo run --release --example crossplatform_helper decode-plugin input.bin
//!   cargo run --release --example crossplatform_helper encode-audio_device > output.bin
//!   cargo run --release --example crossplatform_helper decode-audio_device input.bin

use std::env;
use std::fs;
use std::io::{self, Write};

use sdp_parameter::*;

// Type alias to avoid conflict with wire format Result
type StdResult<T> = std::result::Result<T, Box<dyn std::error::Error>>;

fn encode_parameter() -> StdResult<()> {
    let data = Parameter {
        id: 4_294_967_295,
        name: "Hello from Rust!".to_string(),
        value: 3.14159,
        min: 3.14159,
        max: 3.14159,
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_parameter(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = Parameter::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.id == 4_294_967_295;
    ok = ok && (decoded.value - 3.14159).abs() < 0.0001;
    ok = ok && (decoded.min - 3.14159).abs() < 0.0001;
    ok = ok && (decoded.max - 3.14159).abs() < 0.0001;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_plugin() -> StdResult<()> {
    let data = Plugin {
        id: 4_294_967_295,
        name: "Hello from Rust!".to_string(),
        manufacturer: "Hello from Rust!".to_string(),
        version: 4_294_967_295,
        enabled: true,
        parameters: vec![Default::default(), Default::default(), Default::default()],
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_plugin(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = Plugin::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.id == 4_294_967_295;
    ok = ok && decoded.version == 4_294_967_295;
    ok = ok && decoded.enabled == true;
    ok = ok && decoded.parameters.len() > 0;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_audio_device() -> StdResult<()> {
    let data = AudioDevice {
        device_id: 4_294_967_295,
        device_name: "Hello from Rust!".to_string(),
        sample_rate: 4_294_967_295,
        buffer_size: 4_294_967_295,
        input_channels: 65535,
        output_channels: 65535,
        is_default: true,
        active_plugins: vec![Default::default(), Default::default(), Default::default()],
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_audio_device(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = AudioDevice::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.device_id == 4_294_967_295;
    ok = ok && decoded.sample_rate == 4_294_967_295;
    ok = ok && decoded.buffer_size == 4_294_967_295;
    ok = ok && decoded.input_channels == 65535;
    ok = ok && decoded.output_channels == 65535;
    ok = ok && decoded.is_default == true;
    ok = ok && decoded.active_plugins.len() > 0;

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
        eprintln!("  encode-parameter - Encode Parameter to stdout");
        eprintln!("  decode-parameter <file> - Decode Parameter from file");
        eprintln!("  encode-plugin - Encode Plugin to stdout");
        eprintln!("  decode-plugin <file> - Decode Plugin from file");
        eprintln!("  encode-audio_device - Encode AudioDevice to stdout");
        eprintln!("  decode-audio_device <file> - Decode AudioDevice from file");
        std::process::exit(1);
    }

    match args[1].as_str() {
        "encode-parameter" => encode_parameter()?,
        "decode-parameter" => {
            if args.len() < 3 {
                eprintln!("Error: decode-parameter requires filename argument");
                std::process::exit(1);
            }
            decode_parameter(&args[2])?;
        }
        "encode-plugin" => encode_plugin()?,
        "decode-plugin" => {
            if args.len() < 3 {
                eprintln!("Error: decode-plugin requires filename argument");
                std::process::exit(1);
            }
            decode_plugin(&args[2])?;
        }
        "encode-audio_device" => encode_audio_device()?,
        "decode-audio_device" => {
            if args.len() < 3 {
                eprintln!("Error: decode-audio_device requires filename argument");
                std::process::exit(1);
            }
            decode_audio_device(&args[2])?;
        }
        cmd => {
            eprintln!("Unknown command: {}", cmd);
            std::process::exit(1);
        }
    }

    Ok(())
}
