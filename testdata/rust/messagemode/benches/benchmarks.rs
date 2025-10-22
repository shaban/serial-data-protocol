// Rust message mode benchmarks
// Measures encode/decode performance and compares to Go/C++ implementations

use criterion::{black_box, criterion_group, criterion_main, Criterion, BenchmarkId};
use sdp_point::{Point, Rectangle};
use sdp_point::{encode_point_message, decode_point_message};
use sdp_point::{encode_rectangle_message, decode_rectangle_message};
use sdp_point::decode_message;

fn bench_point_encode(c: &mut Criterion) {
    let point = Point {
        x: 3.14,
        y: 2.71,
    };
    
    c.bench_function("Point: encode_message", |b| {
        b.iter(|| {
            black_box(encode_point_message(black_box(&point)))
        })
    });
}

fn bench_point_decode(c: &mut Criterion) {
    let point = Point { x: 3.14, y: 2.71 };
    let encoded = encode_point_message(&point);
    
    c.bench_function("Point: decode_message", |b| {
        b.iter(|| {
            black_box(decode_point_message(black_box(&encoded)).unwrap())
        })
    });
}

fn bench_point_roundtrip(c: &mut Criterion) {
    let point = Point { x: 3.14, y: 2.71 };
    
    c.bench_function("Point: roundtrip (encode + decode)", |b| {
        b.iter(|| {
            let encoded = encode_point_message(black_box(&point));
            black_box(decode_point_message(&encoded).unwrap())
        })
    });
}

fn bench_rectangle_encode(c: &mut Criterion) {
    let rect = Rectangle {
        top_left: Point { x: 10.0, y: 20.0 },
        width: 100.0,
        height: 50.0,
    };
    
    c.bench_function("Rectangle: encode_message", |b| {
        b.iter(|| {
            black_box(encode_rectangle_message(black_box(&rect)))
        })
    });
}

fn bench_rectangle_decode(c: &mut Criterion) {
    let rect = Rectangle {
        top_left: Point { x: 10.0, y: 20.0 },
        width: 100.0,
        height: 50.0,
    };
    let encoded = encode_rectangle_message(&rect);
    
    c.bench_function("Rectangle: decode_message", |b| {
        b.iter(|| {
            black_box(decode_rectangle_message(black_box(&encoded)).unwrap())
        })
    });
}

fn bench_rectangle_roundtrip(c: &mut Criterion) {
    let rect = Rectangle {
        top_left: Point { x: 10.0, y: 20.0 },
        width: 100.0,
        height: 50.0,
    };
    
    c.bench_function("Rectangle: roundtrip (encode + decode)", |b| {
        b.iter(|| {
            let encoded = encode_rectangle_message(black_box(&rect));
            black_box(decode_rectangle_message(&encoded).unwrap())
        })
    });
}

fn bench_dispatcher_point(c: &mut Criterion) {
    let point = Point { x: 3.14, y: 2.71 };
    let encoded = encode_point_message(&point);
    
    c.bench_function("Dispatcher: decode_message (Point)", |b| {
        b.iter(|| {
            black_box(decode_message(black_box(&encoded)).unwrap())
        })
    });
}

fn bench_dispatcher_rectangle(c: &mut Criterion) {
    let rect = Rectangle {
        top_left: Point { x: 10.0, y: 20.0 },
        width: 100.0,
        height: 50.0,
    };
    let encoded = encode_rectangle_message(&rect);
    
    c.bench_function("Dispatcher: decode_message (Rectangle)", |b| {
        b.iter(|| {
            black_box(decode_message(black_box(&encoded)).unwrap())
        })
    });
}

criterion_group!(
    benches,
    bench_point_encode,
    bench_point_decode,
    bench_point_roundtrip,
    bench_rectangle_encode,
    bench_rectangle_decode,
    bench_rectangle_roundtrip,
    bench_dispatcher_point,
    bench_dispatcher_rectangle,
);

criterion_main!(benches);
