//! Criterion benchmarks for arrays_of_primitives
//!
//! Run with: cargo bench
//! View results: target/criterion/report/index.html

use criterion::{black_box, criterion_group, criterion_main, Criterion};
use sdp_arrays_of_primitives::*;

fn bench_encode_arrays_of_primitives(c: &mut Criterion) {
    let data = ArraysOfPrimitives {
        u8_array: vec![Default::default(), Default::default(), Default::default()],
        u32_array: vec![Default::default(), Default::default(), Default::default()],
        f64_array: vec![Default::default(), Default::default(), Default::default()],
        str_array: vec![Default::default(), Default::default(), Default::default()],
        bool_array: vec![Default::default(), Default::default(), Default::default()],
    };

    c.bench_function("arrays_of_primitives/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_arrays_of_primitives(c: &mut Criterion) {
    let data = ArraysOfPrimitives {
        u8_array: vec![Default::default(), Default::default(), Default::default()],
        u32_array: vec![Default::default(), Default::default(), Default::default()],
        f64_array: vec![Default::default(), Default::default(), Default::default()],
        str_array: vec![Default::default(), Default::default(), Default::default()],
        bool_array: vec![Default::default(), Default::default(), Default::default()],
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("arrays_of_primitives/decode", |bencher| {
        bencher.iter(|| {
            black_box(ArraysOfPrimitives::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_arrays_of_primitives(c: &mut Criterion) {
    let data = ArraysOfPrimitives {
        u8_array: vec![Default::default(), Default::default(), Default::default()],
        u32_array: vec![Default::default(), Default::default(), Default::default()],
        f64_array: vec![Default::default(), Default::default(), Default::default()],
        str_array: vec![Default::default(), Default::default(), Default::default()],
        bool_array: vec![Default::default(), Default::default(), Default::default()],
    };

    c.bench_function("arrays_of_primitives/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(ArraysOfPrimitives::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_item(c: &mut Criterion) {
    let data = Item {
        id: Default::default(),
        name: Default::default(),
    };

    c.bench_function("item/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_item(c: &mut Criterion) {
    let data = Item {
        id: Default::default(),
        name: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("item/decode", |bencher| {
        bencher.iter(|| {
            black_box(Item::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_item(c: &mut Criterion) {
    let data = Item {
        id: Default::default(),
        name: Default::default(),
    };

    c.bench_function("item/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(Item::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_arrays_of_structs(c: &mut Criterion) {
    let data = ArraysOfStructs {
        items: vec![Item::default(), Item::default(), Item::default()],
        count: Default::default(),
    };

    c.bench_function("arrays_of_structs/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_arrays_of_structs(c: &mut Criterion) {
    let data = ArraysOfStructs {
        items: vec![Item::default(), Item::default(), Item::default()],
        count: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("arrays_of_structs/decode", |bencher| {
        bencher.iter(|| {
            black_box(ArraysOfStructs::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_arrays_of_structs(c: &mut Criterion) {
    let data = ArraysOfStructs {
        items: vec![Item::default(), Item::default(), Item::default()],
        count: Default::default(),
    };

    c.bench_function("arrays_of_structs/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(ArraysOfStructs::decode_from_slice(&buf).unwrap());
        });
    });
}

criterion_group!(
    benches,
    bench_encode_arrays_of_primitives,
    bench_decode_arrays_of_primitives,
    bench_roundtrip_arrays_of_primitives,
    bench_encode_item,
    bench_decode_item,
    bench_roundtrip_item,
    bench_encode_arrays_of_structs,
    bench_decode_arrays_of_structs,
    bench_roundtrip_arrays_of_structs
);

criterion_main!(benches);
