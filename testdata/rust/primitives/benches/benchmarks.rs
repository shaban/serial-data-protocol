//! Criterion benchmarks for all_primitives
//!
//! Run with: cargo bench
//! View results: target/criterion/report/index.html

use criterion::{black_box, criterion_group, criterion_main, Criterion};
use sdp_all_primitives::*;

fn bench_encode_all_primitives(c: &mut Criterion) {
    let data = AllPrimitives {
        u8_field: Default::default(),
        u16_field: Default::default(),
        u32_field: Default::default(),
        u64_field: Default::default(),
        i8_field: Default::default(),
        i16_field: Default::default(),
        i32_field: Default::default(),
        i64_field: Default::default(),
        f32_field: Default::default(),
        f64_field: Default::default(),
        bool_field: Default::default(),
        str_field: Default::default(),
    };

    c.bench_function("all_primitives/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_all_primitives(c: &mut Criterion) {
    let data = AllPrimitives {
        u8_field: Default::default(),
        u16_field: Default::default(),
        u32_field: Default::default(),
        u64_field: Default::default(),
        i8_field: Default::default(),
        i16_field: Default::default(),
        i32_field: Default::default(),
        i64_field: Default::default(),
        f32_field: Default::default(),
        f64_field: Default::default(),
        bool_field: Default::default(),
        str_field: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("all_primitives/decode", |bencher| {
        bencher.iter(|| {
            black_box(AllPrimitives::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_all_primitives(c: &mut Criterion) {
    let data = AllPrimitives {
        u8_field: Default::default(),
        u16_field: Default::default(),
        u32_field: Default::default(),
        u64_field: Default::default(),
        i8_field: Default::default(),
        i16_field: Default::default(),
        i32_field: Default::default(),
        i64_field: Default::default(),
        f32_field: Default::default(),
        f64_field: Default::default(),
        bool_field: Default::default(),
        str_field: Default::default(),
    };

    c.bench_function("all_primitives/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(AllPrimitives::decode_from_slice(&buf).unwrap());
        });
    });
}

criterion_group!(
    benches,
    bench_encode_all_primitives,
    bench_decode_all_primitives,
    bench_roundtrip_all_primitives
);

criterion_main!(benches);
