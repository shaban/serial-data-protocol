//! Cross-platform test helper for parameter
//! Usage:
//!   cargo run --release --example crossplatform_helper encode-parameter > output.bin
//!   cargo run --release --example crossplatform_helper decode-parameter input.bin
//!   cargo run --release --example crossplatform_helper encode-plugin > output.bin
//!   cargo run --release --example crossplatform_helper decode-plugin input.bin
//!   cargo run --release --example crossplatform_helper encode-plugin_registry > output.bin
//!   cargo run --release --example crossplatform_helper decode-plugin_registry input.bin

use std::env;
use std::fs;
use std::io::{self, Write};

use sdp_parameter::*;

// Type alias to avoid conflict with wire format Result
type StdResult<T> = std::result::Result<T, Box<dyn std::error::Error>>;

fn encode_parameter() -> StdResult<()> {
    let data = Parameter {
        address: 18_446_744_073_709_551_615,
        display_name: "Hello from Rust!".to_string(),
        identifier: "Hello from Rust!".to_string(),
        unit: "Hello from Rust!".to_string(),
        min_value: 3.14159,
        max_value: 3.14159,
        default_value: 3.14159,
        current_value: 3.14159,
        raw_flags: 4_294_967_295,
        is_writable: true,
        can_ramp: true,
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
    ok = ok && decoded.address == 18_446_744_073_709_551_615;
    ok = ok && (decoded.min_value - 3.14159).abs() < 0.0001;
    ok = ok && (decoded.max_value - 3.14159).abs() < 0.0001;
    ok = ok && (decoded.default_value - 3.14159).abs() < 0.0001;
    ok = ok && (decoded.current_value - 3.14159).abs() < 0.0001;
    ok = ok && decoded.raw_flags == 4_294_967_295;
    ok = ok && decoded.is_writable == true;
    ok = ok && decoded.can_ramp == true;

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
        name: "Hello from Rust!".to_string(),
        manufacturer_id: "Hello from Rust!".to_string(),
        component_type: "Hello from Rust!".to_string(),
        component_subtype: "Hello from Rust!".to_string(),
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
    ok = ok && decoded.parameters.len() > 0;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_plugin_registry() -> StdResult<()> {
    let data = PluginRegistry {
        plugins: vec![Default::default(), Default::default(), Default::default()],
        total_plugin_count: 4_294_967_295,
        total_parameter_count: 4_294_967_295,
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_plugin_registry(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = PluginRegistry::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.plugins.len() > 0;
    ok = ok && decoded.total_plugin_count == 4_294_967_295;
    ok = ok && decoded.total_parameter_count == 4_294_967_295;

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
        eprintln!("  encode-plugin_registry - Encode PluginRegistry to stdout");
        eprintln!("  decode-plugin_registry <file> - Decode PluginRegistry from file");
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
        "encode-plugin_registry" => encode_plugin_registry()?,
        "decode-plugin_registry" => {
            if args.len() < 3 {
                eprintln!("Error: decode-plugin_registry requires filename argument");
                std::process::exit(1);
            }
            decode_plugin_registry(&args[2])?;
        }
        cmd => {
            eprintln!("Unknown command: {}", cmd);
            std::process::exit(1);
        }
    }

    Ok(())
}
