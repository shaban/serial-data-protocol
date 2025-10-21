//! Cross-platform test helper for request
//! Usage:
//!   cargo run --release --example crossplatform_helper encode-request > output.bin
//!   cargo run --release --example crossplatform_helper decode-request input.bin
//!   cargo run --release --example crossplatform_helper encode-metadata > output.bin
//!   cargo run --release --example crossplatform_helper decode-metadata input.bin
//!   cargo run --release --example crossplatform_helper encode-config > output.bin
//!   cargo run --release --example crossplatform_helper decode-config input.bin
//!   cargo run --release --example crossplatform_helper encode-database_config > output.bin
//!   cargo run --release --example crossplatform_helper decode-database_config input.bin
//!   cargo run --release --example crossplatform_helper encode-cache_config > output.bin
//!   cargo run --release --example crossplatform_helper decode-cache_config input.bin
//!   cargo run --release --example crossplatform_helper encode-document > output.bin
//!   cargo run --release --example crossplatform_helper decode-document input.bin
//!   cargo run --release --example crossplatform_helper encode-tag_list > output.bin
//!   cargo run --release --example crossplatform_helper decode-tag_list input.bin

use std::env;
use std::fs;
use std::io::{self, Write};

use sdp_request::*;

// Type alias to avoid conflict with wire format Result
type StdResult<T> = std::result::Result<T, Box<dyn std::error::Error>>;

fn encode_request() -> StdResult<()> {
    let data = Request {
        id: 4_294_967_295,
        metadata: Some(Default::default()),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_request(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = Request::decode_from_slice(&file_data)?;

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

fn encode_metadata() -> StdResult<()> {
    let data = Metadata {
        user_id: 18_446_744_073_709_551_615,
        username: "Hello from Rust!".to_string(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_metadata(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = Metadata::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.user_id == 18_446_744_073_709_551_615;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_config() -> StdResult<()> {
    let data = Config {
        name: "Hello from Rust!".to_string(),
        database: Some(Default::default()),
        cache: Some(Default::default()),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_config(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = Config::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_database_config() -> StdResult<()> {
    let data = DatabaseConfig {
        host: "Hello from Rust!".to_string(),
        port: 65535,
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_database_config(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = DatabaseConfig::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.port == 65535;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_cache_config() -> StdResult<()> {
    let data = CacheConfig {
        size_mb: 4_294_967_295,
        ttl_seconds: 4_294_967_295,
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_cache_config(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = CacheConfig::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.size_mb == 4_294_967_295;
    ok = ok && decoded.ttl_seconds == 4_294_967_295;

    if !ok {
        eprintln!("Validation failed");
        eprintln!("Decoded: {:?}", decoded);
        std::process::exit(1);
    }

    eprintln!("✓ Rust successfully decoded and validated");
    Ok(())
}

fn encode_document() -> StdResult<()> {
    let data = Document {
        id: 4_294_967_295,
        tags: Some(Default::default()),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_document(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = Document::decode_from_slice(&file_data)?;

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

fn encode_tag_list() -> StdResult<()> {
    let data = TagList {
        items: vec!["Hello from Rust!".to_string(), "Hello from Rust!".to_string(), "Hello from Rust!".to_string()],
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf)?;

    io::stdout().write_all(&buf)?;
    Ok(())
}

fn decode_tag_list(filename: &str) -> StdResult<()> {
    let file_data = fs::read(filename)?;
    let decoded = TagList::decode_from_slice(&file_data)?;

    // Basic validation
    let mut ok = true;
    ok = ok && decoded.items.len() > 0;

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
        eprintln!("  encode-request - Encode Request to stdout");
        eprintln!("  decode-request <file> - Decode Request from file");
        eprintln!("  encode-metadata - Encode Metadata to stdout");
        eprintln!("  decode-metadata <file> - Decode Metadata from file");
        eprintln!("  encode-config - Encode Config to stdout");
        eprintln!("  decode-config <file> - Decode Config from file");
        eprintln!("  encode-database_config - Encode DatabaseConfig to stdout");
        eprintln!("  decode-database_config <file> - Decode DatabaseConfig from file");
        eprintln!("  encode-cache_config - Encode CacheConfig to stdout");
        eprintln!("  decode-cache_config <file> - Decode CacheConfig from file");
        eprintln!("  encode-document - Encode Document to stdout");
        eprintln!("  decode-document <file> - Decode Document from file");
        eprintln!("  encode-tag_list - Encode TagList to stdout");
        eprintln!("  decode-tag_list <file> - Decode TagList from file");
        std::process::exit(1);
    }

    match args[1].as_str() {
        "encode-request" => encode_request()?,
        "decode-request" => {
            if args.len() < 3 {
                eprintln!("Error: decode-request requires filename argument");
                std::process::exit(1);
            }
            decode_request(&args[2])?;
        }
        "encode-metadata" => encode_metadata()?,
        "decode-metadata" => {
            if args.len() < 3 {
                eprintln!("Error: decode-metadata requires filename argument");
                std::process::exit(1);
            }
            decode_metadata(&args[2])?;
        }
        "encode-config" => encode_config()?,
        "decode-config" => {
            if args.len() < 3 {
                eprintln!("Error: decode-config requires filename argument");
                std::process::exit(1);
            }
            decode_config(&args[2])?;
        }
        "encode-database_config" => encode_database_config()?,
        "decode-database_config" => {
            if args.len() < 3 {
                eprintln!("Error: decode-database_config requires filename argument");
                std::process::exit(1);
            }
            decode_database_config(&args[2])?;
        }
        "encode-cache_config" => encode_cache_config()?,
        "decode-cache_config" => {
            if args.len() < 3 {
                eprintln!("Error: decode-cache_config requires filename argument");
                std::process::exit(1);
            }
            decode_cache_config(&args[2])?;
        }
        "encode-document" => encode_document()?,
        "decode-document" => {
            if args.len() < 3 {
                eprintln!("Error: decode-document requires filename argument");
                std::process::exit(1);
            }
            decode_document(&args[2])?;
        }
        "encode-tag_list" => encode_tag_list()?,
        "decode-tag_list" => {
            if args.len() < 3 {
                eprintln!("Error: decode-tag_list requires filename argument");
                std::process::exit(1);
            }
            decode_tag_list(&args[2])?;
        }
        cmd => {
            eprintln!("Unknown command: {}", cmd);
            std::process::exit(1);
        }
    }

    Ok(())
}
