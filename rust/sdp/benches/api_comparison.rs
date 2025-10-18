// Wire format API comparison: Read/Write traits vs direct byte slices
// This demonstrates the performance difference between the two approaches

use criterion::{black_box, criterion_group, criterion_main, Criterion, Throughput};
use sdp::wire::{Decoder, Encoder};
use sdp::wire_slice;
use std::io::Cursor;

const TEST_STRING: &str = "Hello, World! This is a test string for benchmarking.";

// ============================================================================
// Trait-based API (wire module with Read/Write)
// ============================================================================

fn bench_trait_encode_u32(c: &mut Criterion) {
    c.bench_function("trait/encode_u32", |b| {
        b.iter(|| {
            let mut buf = Vec::with_capacity(4);
            let mut enc = Encoder::new(&mut buf);
            enc.write_u32(black_box(0x12345678)).unwrap();
        });
    });
}

fn bench_trait_decode_u32(c: &mut Criterion) {
    let data = vec![0x78, 0x56, 0x34, 0x12];
    c.bench_function("trait/decode_u32", |b| {
        b.iter(|| {
            let mut cursor = Cursor::new(&data);
            let mut dec = Decoder::new(&mut cursor);
            black_box(dec.read_u32().unwrap());
        });
    });
}

fn bench_trait_encode_string(c: &mut Criterion) {
    let mut group = c.benchmark_group("trait/string");
    group.throughput(Throughput::Bytes(TEST_STRING.len() as u64));
    
    group.bench_function("encode", |b| {
        b.iter(|| {
            let mut buf = Vec::with_capacity(100);
            let mut enc = Encoder::new(&mut buf);
            enc.write_string(black_box(TEST_STRING)).unwrap();
        });
    });
    group.finish();
}

fn bench_trait_decode_string(c: &mut Criterion) {
    let mut data = Vec::new();
    {
        let mut enc = Encoder::new(&mut data);
        enc.write_string(TEST_STRING).unwrap();
    }
    
    let mut group = c.benchmark_group("trait/string");
    group.throughput(Throughput::Bytes(TEST_STRING.len() as u64));
    
    group.bench_function("decode", |b| {
        b.iter(|| {
            let mut cursor = Cursor::new(&data);
            let mut dec = Decoder::new(&mut cursor);
            black_box(dec.read_string().unwrap());
        });
    });
    group.finish();
}

// ============================================================================
// Slice-based API (wire_slice module with &[u8]/&mut [u8])
// ============================================================================

fn bench_slice_encode_u32(c: &mut Criterion) {
    c.bench_function("slice/encode_u32", |b| {
        b.iter(|| {
            let mut buf = [0u8; 4];
            wire_slice::encode_u32(&mut buf, 0, black_box(0x12345678)).unwrap();
        });
    });
}

fn bench_slice_decode_u32(c: &mut Criterion) {
    let data = [0x78u8, 0x56, 0x34, 0x12];
    c.bench_function("slice/decode_u32", |b| {
        b.iter(|| {
            black_box(wire_slice::decode_u32(&data, 0).unwrap());
        });
    });
}

fn bench_slice_encode_string(c: &mut Criterion) {
    let mut group = c.benchmark_group("slice/string");
    group.throughput(Throughput::Bytes(TEST_STRING.len() as u64));
    
    group.bench_function("encode", |b| {
        b.iter(|| {
            let mut buf = [0u8; 100];
            wire_slice::encode_string(&mut buf, 0, black_box(TEST_STRING)).unwrap();
        });
    });
    group.finish();
}

fn bench_slice_decode_string(c: &mut Criterion) {
    let mut buf = [0u8; 100];
    wire_slice::encode_string(&mut buf, 0, TEST_STRING).unwrap();
    
    let mut group = c.benchmark_group("slice/string");
    group.throughput(Throughput::Bytes(TEST_STRING.len() as u64));
    
    group.bench_function("decode", |b| {
        b.iter(|| {
            black_box(wire_slice::decode_string(&buf, 0).unwrap());
        });
    });
    group.finish();
}

// ============================================================================
// Complex roundtrip: Multiple values
// ============================================================================

fn bench_trait_complex_roundtrip(c: &mut Criterion) {
    c.bench_function("trait/complex_roundtrip", |b| {
        b.iter(|| {
            let mut buf = Vec::new();
            let mut enc = Encoder::new(&mut buf);
            
            // Encode multiple values
            enc.write_u32(black_box(42)).unwrap();
            enc.write_f64(black_box(3.14159)).unwrap();
            enc.write_bool(black_box(true)).unwrap();
            enc.write_string(black_box("test")).unwrap();
            
            // Decode them back
            let mut cursor = Cursor::new(&buf);
            let mut dec = Decoder::new(&mut cursor);
            
            black_box(dec.read_u32().unwrap());
            black_box(dec.read_f64().unwrap());
            black_box(dec.read_bool().unwrap());
            black_box(dec.read_string().unwrap());
        });
    });
}

fn bench_slice_complex_roundtrip(c: &mut Criterion) {
    c.bench_function("slice/complex_roundtrip", |b| {
        b.iter(|| {
            let mut buf = [0u8; 100];
            let mut offset = 0;
            
            // Encode multiple values
            wire_slice::encode_u32(&mut buf, offset, black_box(42)).unwrap();
            offset += 4;
            wire_slice::encode_f64(&mut buf, offset, black_box(3.14159)).unwrap();
            offset += 8;
            wire_slice::encode_bool(&mut buf, offset, black_box(true)).unwrap();
            offset += 1;
            let written = wire_slice::encode_string(&mut buf, offset, black_box("test")).unwrap();
            offset += written;
            
            // Decode them back
            let mut read_offset = 0;
            black_box(wire_slice::decode_u32(&buf, read_offset).unwrap());
            read_offset += 4;
            black_box(wire_slice::decode_f64(&buf, read_offset).unwrap());
            read_offset += 8;
            black_box(wire_slice::decode_bool(&buf, read_offset).unwrap());
            read_offset += 1;
            let (s, consumed) = wire_slice::decode_string(&buf, read_offset).unwrap();
            black_box(s);
        });
    });
}

criterion_group!(
    benches,
    bench_trait_encode_u32,
    bench_trait_decode_u32,
    bench_trait_encode_string,
    bench_trait_decode_string,
    bench_slice_encode_u32,
    bench_slice_decode_u32,
    bench_slice_encode_string,
    bench_slice_decode_string,
    bench_trait_complex_roundtrip,
    bench_slice_complex_roundtrip
);
criterion_main!(benches);
