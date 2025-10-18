// Generated code benchmarks - AudioUnit schema
// These benchmarks test the performance of generated encode/decode code
// Comparable to the Go benchmarks in benchmarks/comparison_test.go

use criterion::{black_box, criterion_group, criterion_main, Criterion, Throughput};

#[path = "../../../testdata/audiounit/rust/lib.rs"]
mod audiounit;

#[path = "../../../testdata/primitives/rust/lib.rs"]
mod primitives;

use audiounit::*;
use primitives::*;

// Create test data matching the Go benchmarks
fn create_test_primitives() -> AllPrimitives {
    AllPrimitives {
        u8_field: 255,
        u16_field: 65535,
        u32_field: 4_294_967_295,
        u64_field: 18_446_744_073_709_551_615,
        i8_field: -128,
        i16_field: -32768,
        i32_field: -2_147_483_648,
        i64_field: -9_223_372_036_854_775_808,
        f32_field: 3.14159,
        f64_field: 2.718281828459045,
        bool_field: true,
        str_field: "Hello from Rust!".to_string(),
    }
}

fn create_test_audiounit() -> PluginRegistry {
    // Create a registry similar to what's in the Go benchmarks
    // 2 plugins with varying numbers of parameters
    let param1 = Parameter {
        address: 0,
        display_name: "Volume".to_string(),
        identifier: "volume".to_string(),
        unit: "dB".to_string(),
        min_value: -96.0,
        max_value: 6.0,
        default_value: 0.0,
        current_value: -12.0,
        raw_flags: 0x01,
        is_writable: true,
        can_ramp: true,
    };
    
    let param2 = Parameter {
        address: 1,
        display_name: "Pan".to_string(),
        identifier: "pan".to_string(),
        unit: "%".to_string(),
        min_value: -100.0,
        max_value: 100.0,
        default_value: 0.0,
        current_value: 0.0,
        raw_flags: 0x01,
        is_writable: true,
        can_ramp: true,
    };
    
    let plugin1 = Plugin {
        name: "Test Synth".to_string(),
        manufacturer_id: "TEST".to_string(),
        component_type: "aumu".to_string(),
        component_subtype: "test".to_string(),
        parameters: vec![param1.clone(), param2.clone()],
    };
    
    let plugin2 = Plugin {
        name: "Test Effect".to_string(),
        manufacturer_id: "TEST".to_string(),
        component_type: "aumf".to_string(),
        component_subtype: "tsfx".to_string(),
        parameters: vec![param1.clone()],
    };
    
    PluginRegistry {
        plugins: vec![plugin1, plugin2],
        total_plugin_count: 2,
        total_parameter_count: 3,
    }
}

fn bench_primitives_encode(c: &mut Criterion) {
    let data = create_test_primitives();
    
    c.bench_function("primitives_encode", |b| {
        b.iter(|| {
            let mut buf = Vec::new();
            black_box(&data).encode(&mut buf).unwrap();
            black_box(buf);
        });
    });
}

fn bench_primitives_decode(c: &mut Criterion) {
    let data = create_test_primitives();
    let mut encoded = Vec::new();
    data.encode(&mut encoded).unwrap();
    
    c.bench_function("primitives_decode", |b| {
        b.iter(|| {
            let decoded = AllPrimitives::decode(&mut black_box(&encoded[..])).unwrap();
            black_box(decoded);
        });
    });
}

fn bench_primitives_roundtrip(c: &mut Criterion) {
    let data = create_test_primitives();
    
    c.bench_function("primitives_roundtrip", |b| {
        b.iter(|| {
            let mut buf = Vec::new();
            black_box(&data).encode(&mut buf).unwrap();
            let decoded = AllPrimitives::decode(&mut black_box(&buf[..])).unwrap();
            black_box(decoded);
        });
    });
}

fn bench_audiounit_encode(c: &mut Criterion) {
    let data = create_test_audiounit();
    
    // Count total bytes for throughput measurement
    let mut size_buf = Vec::new();
    data.encode(&mut size_buf).unwrap();
    
    let mut group = c.benchmark_group("audiounit_encode");
    group.throughput(Throughput::Bytes(size_buf.len() as u64));
    
    group.bench_function("encode", |b| {
        b.iter(|| {
            let mut buf = Vec::new();
            black_box(&data).encode(&mut buf).unwrap();
            black_box(buf);
        });
    });
    group.finish();
}

fn bench_audiounit_decode(c: &mut Criterion) {
    let data = create_test_audiounit();
    let mut encoded = Vec::new();
    data.encode(&mut encoded).unwrap();
    
    let mut group = c.benchmark_group("audiounit_decode");
    group.throughput(Throughput::Bytes(encoded.len() as u64));
    
    group.bench_function("decode", |b| {
        b.iter(|| {
            let decoded = PluginRegistry::decode(&mut black_box(&encoded[..])).unwrap();
            black_box(decoded);
        });
    });
    group.finish();
}

fn bench_audiounit_roundtrip(c: &mut Criterion) {
    let data = create_test_audiounit();
    
    let mut size_buf = Vec::new();
    data.encode(&mut size_buf).unwrap();
    
    let mut group = c.benchmark_group("audiounit_roundtrip");
    group.throughput(Throughput::Bytes(size_buf.len() as u64));
    
    group.bench_function("roundtrip", |b| {
        b.iter(|| {
            let mut buf = Vec::new();
            black_box(&data).encode(&mut buf).unwrap();
            let decoded = PluginRegistry::decode(&mut black_box(&buf[..])).unwrap();
            black_box(decoded);
        });
    });
    group.finish();
}

criterion_group!(
    benches,
    bench_primitives_encode,
    bench_primitives_decode,
    bench_primitives_roundtrip,
    bench_audiounit_encode,
    bench_audiounit_decode,
    bench_audiounit_roundtrip
);
criterion_main!(benches);
