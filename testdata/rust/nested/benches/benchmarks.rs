//! Criterion benchmarks for point
//!
//! Run with: cargo bench
//! View results: target/criterion/report/index.html

use criterion::{black_box, criterion_group, criterion_main, Criterion};
use sdp_point::*;

fn bench_encode_point(c: &mut Criterion) {
    let data = Point {
        x: Default::default(),
        y: Default::default(),
    };

    c.bench_function("point/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_point(c: &mut Criterion) {
    let data = Point {
        x: Default::default(),
        y: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("point/decode", |bencher| {
        bencher.iter(|| {
            black_box(Point::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_point(c: &mut Criterion) {
    let data = Point {
        x: Default::default(),
        y: Default::default(),
    };

    c.bench_function("point/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(Point::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_rectangle(c: &mut Criterion) {
    let data = Rectangle {
        top_left: Point::default(),
        bottom_right: Point::default(),
        color: Default::default(),
    };

    c.bench_function("rectangle/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_rectangle(c: &mut Criterion) {
    let data = Rectangle {
        top_left: Point::default(),
        bottom_right: Point::default(),
        color: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("rectangle/decode", |bencher| {
        bencher.iter(|| {
            black_box(Rectangle::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_rectangle(c: &mut Criterion) {
    let data = Rectangle {
        top_left: Point::default(),
        bottom_right: Point::default(),
        color: Default::default(),
    };

    c.bench_function("rectangle/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(Rectangle::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_scene(c: &mut Criterion) {
    let data = Scene {
        name: Default::default(),
        main_rect: Rectangle::default(),
        count: Default::default(),
    };

    c.bench_function("scene/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_scene(c: &mut Criterion) {
    let data = Scene {
        name: Default::default(),
        main_rect: Rectangle::default(),
        count: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("scene/decode", |bencher| {
        bencher.iter(|| {
            black_box(Scene::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_scene(c: &mut Criterion) {
    let data = Scene {
        name: Default::default(),
        main_rect: Rectangle::default(),
        count: Default::default(),
    };

    c.bench_function("scene/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(Scene::decode_from_slice(&buf).unwrap());
        });
    });
}

criterion_group!(
    benches,
    bench_encode_point,
    bench_decode_point,
    bench_roundtrip_point,
    bench_encode_rectangle,
    bench_decode_rectangle,
    bench_roundtrip_rectangle,
    bench_encode_scene,
    bench_decode_scene,
    bench_roundtrip_scene
);

criterion_main!(benches);
