use criterion::{black_box, criterion_group, criterion_main, Criterion};
use sdp_parameter::*;
use std::fs;

fn load_audiounit_data() -> (PluginRegistry, Vec<u8>) {
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    let binary_path = format!("{}/../../../testdata/binaries/audiounit.sdpb", manifest_dir);
    let binary = fs::read(&binary_path)
        .unwrap_or_else(|_| panic!("Failed to read audiounit.sdpb at {}", binary_path));
    let data = PluginRegistry::decode_from_slice(&binary)
        .expect("Failed to decode audiounit.sdpb");
    (data, binary)
}

fn bench_encode(c: &mut Criterion) {
    let (registry, _) = load_audiounit_data();
    let size = registry.encoded_size();
    let mut buf = vec![0u8; size];
    
    c.bench_function("rustexp_audiounit_encode", |b| {
        b.iter(|| {
            black_box(registry.encode_to_slice(&mut buf).unwrap())
        })
    });
}

fn bench_decode(c: &mut Criterion) {
    let (_, binary) = load_audiounit_data();
    
    c.bench_function("rustexp_audiounit_decode", |b| {
        b.iter(|| {
            let decoded = PluginRegistry::decode_from_slice(&binary).unwrap();
            black_box(decoded)
        })
    });
}

fn bench_roundtrip(c: &mut Criterion) {
    let (registry, _) = load_audiounit_data();
    let size = registry.encoded_size();
    let mut buf = vec![0u8; size];
    
    c.bench_function("rustexp_audiounit_roundtrip", |b| {
        b.iter(|| {
            let size = registry.encode_to_slice(&mut buf).unwrap();
            let decoded = PluginRegistry::decode_from_slice(&buf[..size]).unwrap();
            black_box(decoded)
        })
    });
}

criterion_group!(benches, bench_encode, bench_decode, bench_roundtrip);
criterion_main!(benches);
