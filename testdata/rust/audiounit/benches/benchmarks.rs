//! Criterion benchmarks for parameter
//!
//! Run with: cargo bench
//! View results: target/criterion/report/index.html

use criterion::{black_box, criterion_group, criterion_main, Criterion};
use sdp_parameter::*;

fn bench_encode_parameter(c: &mut Criterion) {
    let data = Parameter {
        address: Default::default(),
        display_name: Default::default(),
        identifier: Default::default(),
        unit: Default::default(),
        min_value: Default::default(),
        max_value: Default::default(),
        default_value: Default::default(),
        current_value: Default::default(),
        raw_flags: Default::default(),
        is_writable: Default::default(),
        can_ramp: Default::default(),
    };

    c.bench_function("parameter/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_parameter(c: &mut Criterion) {
    let data = Parameter {
        address: Default::default(),
        display_name: Default::default(),
        identifier: Default::default(),
        unit: Default::default(),
        min_value: Default::default(),
        max_value: Default::default(),
        default_value: Default::default(),
        current_value: Default::default(),
        raw_flags: Default::default(),
        is_writable: Default::default(),
        can_ramp: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("parameter/decode", |bencher| {
        bencher.iter(|| {
            black_box(Parameter::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_parameter(c: &mut Criterion) {
    let data = Parameter {
        address: Default::default(),
        display_name: Default::default(),
        identifier: Default::default(),
        unit: Default::default(),
        min_value: Default::default(),
        max_value: Default::default(),
        default_value: Default::default(),
        current_value: Default::default(),
        raw_flags: Default::default(),
        is_writable: Default::default(),
        can_ramp: Default::default(),
    };

    c.bench_function("parameter/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(Parameter::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_plugin(c: &mut Criterion) {
    let data = Plugin {
        name: Default::default(),
        manufacturer_id: Default::default(),
        component_type: Default::default(),
        component_subtype: Default::default(),
        parameters: vec![Parameter::default(), Parameter::default(), Parameter::default()],
    };

    c.bench_function("plugin/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_plugin(c: &mut Criterion) {
    let data = Plugin {
        name: Default::default(),
        manufacturer_id: Default::default(),
        component_type: Default::default(),
        component_subtype: Default::default(),
        parameters: vec![Parameter::default(), Parameter::default(), Parameter::default()],
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("plugin/decode", |bencher| {
        bencher.iter(|| {
            black_box(Plugin::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_plugin(c: &mut Criterion) {
    let data = Plugin {
        name: Default::default(),
        manufacturer_id: Default::default(),
        component_type: Default::default(),
        component_subtype: Default::default(),
        parameters: vec![Parameter::default(), Parameter::default(), Parameter::default()],
    };

    c.bench_function("plugin/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(Plugin::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_plugin_registry(c: &mut Criterion) {
    let data = PluginRegistry {
        plugins: vec![Plugin::default(), Plugin::default(), Plugin::default()],
        total_plugin_count: Default::default(),
        total_parameter_count: Default::default(),
    };

    c.bench_function("plugin_registry/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_plugin_registry(c: &mut Criterion) {
    let data = PluginRegistry {
        plugins: vec![Plugin::default(), Plugin::default(), Plugin::default()],
        total_plugin_count: Default::default(),
        total_parameter_count: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("plugin_registry/decode", |bencher| {
        bencher.iter(|| {
            black_box(PluginRegistry::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_plugin_registry(c: &mut Criterion) {
    let data = PluginRegistry {
        plugins: vec![Plugin::default(), Plugin::default(), Plugin::default()],
        total_plugin_count: Default::default(),
        total_parameter_count: Default::default(),
    };

    c.bench_function("plugin_registry/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(PluginRegistry::decode_from_slice(&buf).unwrap());
        });
    });
}

criterion_group!(
    benches,
    bench_encode_parameter,
    bench_decode_parameter,
    bench_roundtrip_parameter,
    bench_encode_plugin,
    bench_decode_plugin,
    bench_roundtrip_plugin,
    bench_encode_plugin_registry,
    bench_decode_plugin_registry,
    bench_roundtrip_plugin_registry
);

criterion_main!(benches);
