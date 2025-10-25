use criterion::{black_box, criterion_group, criterion_main, Criterion};
use sdp_arrays_of_primitives::*;
use std::fs;

fn load_test_data() -> (ArraysOfPrimitives, Vec<u8>) {
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    let binary_path = format!("{}/../../../testdata/binaries/arrays_primitives.sdpb", manifest_dir);
    let binary = fs::read(&binary_path)
        .unwrap_or_else(|_| panic!("Failed to read arrays_primitives.sdpb at {}", binary_path));
    let data = ArraysOfPrimitives::decode_from_slice(&binary)
        .expect("Failed to decode arrays_primitives.sdpb");
    (data, binary)
}

fn bench_encode(c: &mut Criterion) {
    let (data, _) = load_test_data();
    let mut buf = vec![0u8; 100000];
    c.bench_function("rustexp_arrays_encode", |b| {
        b.iter(|| {
            let size = data.encode_to_slice(&mut buf).unwrap();
            black_box(size)
        })
    });
}

fn bench_decode(c: &mut Criterion) {
    let (_, binary) = load_test_data();
    c.bench_function("rustexp_arrays_decode", |b| {
        b.iter(|| {
            let decoded = ArraysOfPrimitives::decode_from_slice(&binary).unwrap();
            black_box(decoded)
        })
    });
}

fn bench_roundtrip(c: &mut Criterion) {
    let (data, _) = load_test_data();
    let mut buf = vec![0u8; 100000];
    c.bench_function("rustexp_arrays_roundtrip", |b| {
        b.iter(|| {
            let size = data.encode_to_slice(&mut buf).unwrap();
            let decoded = ArraysOfPrimitives::decode_from_slice(&buf[..size]).unwrap();
            black_box(decoded)
        })
    });
}

criterion_group!(benches, bench_encode, bench_decode, bench_roundtrip);
criterion_main!(benches);
