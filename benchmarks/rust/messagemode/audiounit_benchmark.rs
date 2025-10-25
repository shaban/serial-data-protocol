// Rust AudioUnit Message Mode Benchmark
// Measures encoding/decoding performance with real-world AudioUnit data
// Schema: PluginRegistry (62 plugins, 1,759 parameters, ~110KB)
//
// Matches C++ benchmark: benchmarks/cpp/messagemode/bench_audiounit.cpp

use criterion::{black_box, criterion_group, criterion_main, Criterion};
use std::fs;
use sdp_parameter::{PluginRegistry, encode_plugin_registry_message, decode_plugin_registry_message, decode_message, Message};

// Load .sdpb file (path from environment or default)
fn load_audiounit_data() -> PluginRegistry {
    let path = std::env::var("AUDIOUNIT_DATA")
        .unwrap_or_else(|_| "../../../testdata/binaries/audiounit.sdpb".to_string());
    
    let data = fs::read(&path)
        .unwrap_or_else(|e| panic!("Failed to load {}: {}", path, e));
    
    // Decode to get PluginRegistry struct (byte mode)
    PluginRegistry::decode_from_slice(&data)
        .expect("Failed to decode AudioUnit data")
}

fn bench_encode_byte_mode(c: &mut Criterion) {
    let registry = load_audiounit_data();
    
    c.bench_function("AudioUnit: Byte mode encode", |b| {
        b.iter(|| {
            let size = registry.encoded_size();
            let mut buf = vec![0u8; size];
            black_box(registry.encode_to_slice(&mut buf).unwrap())
        })
    });
}

fn bench_encode_message_mode(c: &mut Criterion) {
    let registry = load_audiounit_data();
    
    c.bench_function("AudioUnit: Message mode encode", |b| {
        b.iter(|| {
            black_box(encode_plugin_registry_message(black_box(&registry)))
        })
    });
}

fn bench_decode_byte_mode(c: &mut Criterion) {
    let registry = load_audiounit_data();
    let size = registry.encoded_size();
    let mut encoded = vec![0u8; size];
    registry.encode_to_slice(&mut encoded).unwrap();
    
    c.bench_function("AudioUnit: Byte mode decode", |b| {
        b.iter(|| {
            black_box(PluginRegistry::decode_from_slice(black_box(&encoded)).unwrap())
        })
    });
}

fn bench_decode_message_mode(c: &mut Criterion) {
    let registry = load_audiounit_data();
    let encoded = encode_plugin_registry_message(&registry);
    
    c.bench_function("AudioUnit: Message mode decode", |b| {
        b.iter(|| {
            black_box(decode_plugin_registry_message(black_box(&encoded)).unwrap())
        })
    });
}

fn bench_roundtrip_byte_mode(c: &mut Criterion) {
    let registry = load_audiounit_data();
    
    c.bench_function("AudioUnit: Byte mode roundtrip", |b| {
        b.iter(|| {
            let size = registry.encoded_size();
            let mut buf = vec![0u8; size];
            registry.encode_to_slice(&mut buf).unwrap();
            black_box(PluginRegistry::decode_from_slice(&buf).unwrap())
        })
    });
}

fn bench_roundtrip_message_mode(c: &mut Criterion) {
    let registry = load_audiounit_data();
    
    c.bench_function("AudioUnit: Message mode roundtrip", |b| {
        b.iter(|| {
            let encoded = encode_plugin_registry_message(black_box(&registry));
            black_box(decode_plugin_registry_message(&encoded).unwrap())
        })
    });
}

fn bench_dispatcher(c: &mut Criterion) {
    let registry = load_audiounit_data();
    let encoded = encode_plugin_registry_message(&registry);
    
    c.bench_function("AudioUnit: Dispatcher decode_message", |b| {
        b.iter(|| {
            match black_box(decode_message(black_box(&encoded)).unwrap()) {
                Message::Parameter(_) => unreachable!(),
                Message::Plugin(_) => unreachable!(),
                Message::PluginRegistry(r) => black_box(r),
            }
        })
    });
}

criterion_group!(
    benches,
    bench_encode_byte_mode,
    bench_encode_message_mode,
    bench_decode_byte_mode,
    bench_decode_message_mode,
    bench_roundtrip_byte_mode,
    bench_roundtrip_message_mode,
    bench_dispatcher,
);

criterion_main!(benches);
