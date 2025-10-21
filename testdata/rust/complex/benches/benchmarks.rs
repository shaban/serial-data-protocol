//! Criterion benchmarks for parameter
//!
//! Run with: cargo bench
//! View results: target/criterion/report/index.html

use criterion::{black_box, criterion_group, criterion_main, Criterion};
use sdp_parameter::*;

fn bench_encode_parameter(c: &mut Criterion) {
    let data = Parameter {
        id: Default::default(),
        name: Default::default(),
        value: Default::default(),
        min: Default::default(),
        max: Default::default(),
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
        id: Default::default(),
        name: Default::default(),
        value: Default::default(),
        min: Default::default(),
        max: Default::default(),
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
        id: Default::default(),
        name: Default::default(),
        value: Default::default(),
        min: Default::default(),
        max: Default::default(),
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
        id: Default::default(),
        name: Default::default(),
        manufacturer: Default::default(),
        version: Default::default(),
        enabled: Default::default(),
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
        id: Default::default(),
        name: Default::default(),
        manufacturer: Default::default(),
        version: Default::default(),
        enabled: Default::default(),
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
        id: Default::default(),
        name: Default::default(),
        manufacturer: Default::default(),
        version: Default::default(),
        enabled: Default::default(),
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

fn bench_encode_audio_device(c: &mut Criterion) {
    let data = AudioDevice {
        device_id: Default::default(),
        device_name: Default::default(),
        sample_rate: Default::default(),
        buffer_size: Default::default(),
        input_channels: Default::default(),
        output_channels: Default::default(),
        is_default: Default::default(),
        active_plugins: vec![Plugin::default(), Plugin::default(), Plugin::default()],
    };

    c.bench_function("audio_device/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_audio_device(c: &mut Criterion) {
    let data = AudioDevice {
        device_id: Default::default(),
        device_name: Default::default(),
        sample_rate: Default::default(),
        buffer_size: Default::default(),
        input_channels: Default::default(),
        output_channels: Default::default(),
        is_default: Default::default(),
        active_plugins: vec![Plugin::default(), Plugin::default(), Plugin::default()],
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("audio_device/decode", |bencher| {
        bencher.iter(|| {
            black_box(AudioDevice::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_audio_device(c: &mut Criterion) {
    let data = AudioDevice {
        device_id: Default::default(),
        device_name: Default::default(),
        sample_rate: Default::default(),
        buffer_size: Default::default(),
        input_channels: Default::default(),
        output_channels: Default::default(),
        is_default: Default::default(),
        active_plugins: vec![Plugin::default(), Plugin::default(), Plugin::default()],
    };

    c.bench_function("audio_device/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(AudioDevice::decode_from_slice(&buf).unwrap());
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
    bench_encode_audio_device,
    bench_decode_audio_device,
    bench_roundtrip_audio_device
);

criterion_main!(benches);
