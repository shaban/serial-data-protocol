// Wire format encoding/decoding benchmarks
// These benchmark the low-level primitives to establish baseline performance

use criterion::{black_box, criterion_group, criterion_main, Criterion, Throughput};
use sdp::wire::{Decoder, Encoder};
use std::io::Cursor;

fn bench_encode_u32(c: &mut Criterion) {
    c.bench_function("encode_u32", |b| {
        b.iter(|| {
            let mut buf = Vec::with_capacity(4);
            let mut enc = Encoder::new(&mut buf);
            enc.write_u32(black_box(0x12345678)).unwrap();
        });
    });
}

fn bench_decode_u32(c: &mut Criterion) {
    let data = vec![0x78, 0x56, 0x34, 0x12]; // Little-endian
    c.bench_function("decode_u32", |b| {
        b.iter(|| {
            let mut cursor = Cursor::new(&data);
            let mut dec = Decoder::new(&mut cursor);
            black_box(dec.read_u32().unwrap());
        });
    });
}

fn bench_encode_string(c: &mut Criterion) {
    let s = "Hello, World! This is a test string.";
    let mut group = c.benchmark_group("encode_string");
    group.throughput(Throughput::Bytes(s.len() as u64));
    
    group.bench_function("encode", |b| {
        b.iter(|| {
            let mut buf = Vec::with_capacity(100);
            let mut enc = Encoder::new(&mut buf);
            enc.write_string(black_box(s)).unwrap();
        });
    });
    group.finish();
}

fn bench_decode_string(c: &mut Criterion) {
    // Pre-encode the string
    let s = "Hello, World! This is a test string.";
    let mut data = Vec::new();
    {
        let mut enc = Encoder::new(&mut data);
        enc.write_string(s).unwrap();
    }
    
    let mut group = c.benchmark_group("decode_string");
    group.throughput(Throughput::Bytes(s.len() as u64));
    
    group.bench_function("decode", |b| {
        b.iter(|| {
            let mut cursor = Cursor::new(&data);
            let mut dec = Decoder::new(&mut cursor);
            black_box(dec.read_string().unwrap());
        });
    });
    group.finish();
}

fn bench_encode_array(c: &mut Criterion) {
    let arr: Vec<u32> = (0..1000).collect();
    let mut group = c.benchmark_group("encode_array");
    group.throughput(Throughput::Elements(1000));
    
    group.bench_function("1000_u32", |b| {
        b.iter(|| {
            let mut buf = Vec::with_capacity(4004); // 4 + 1000*4
            let mut enc = Encoder::new(&mut buf);
            enc.write_u32(black_box(arr.len() as u32)).unwrap();
            for &item in black_box(&arr) {
                enc.write_u32(item).unwrap();
            }
        });
    });
    group.finish();
}

fn bench_decode_array(c: &mut Criterion) {
    // Pre-encode array
    let arr: Vec<u32> = (0..1000).collect();
    let mut data = Vec::new();
    {
        let mut enc = Encoder::new(&mut data);
        enc.write_u32(arr.len() as u32).unwrap();
        for &item in &arr {
            enc.write_u32(item).unwrap();
        }
    }
    
    let mut group = c.benchmark_group("decode_array");
    group.throughput(Throughput::Elements(1000));
    
    group.bench_function("1000_u32", |b| {
        b.iter(|| {
            let mut cursor = Cursor::new(&data);
            let mut dec = Decoder::new(&mut cursor);
            let len = dec.read_u32().unwrap() as usize;
            let mut result = Vec::with_capacity(len);
            for _ in 0..len {
                result.push(dec.read_u32().unwrap());
            }
            black_box(result);
        });
    });
    group.finish();
}

criterion_group!(
    benches,
    bench_encode_u32,
    bench_decode_u32,
    bench_encode_string,
    bench_decode_string,
    bench_encode_array,
    bench_decode_array
);
criterion_main!(benches);
