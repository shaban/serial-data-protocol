//! Criterion benchmarks for request
//!
//! Run with: cargo bench
//! View results: target/criterion/report/index.html

use criterion::{black_box, criterion_group, criterion_main, Criterion};
use sdp_request::*;

fn bench_encode_request(c: &mut Criterion) {
    let data = Request {
        id: Default::default(),
        metadata: Some(Metadata::default()),
    };

    c.bench_function("request/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_request(c: &mut Criterion) {
    let data = Request {
        id: Default::default(),
        metadata: Some(Metadata::default()),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("request/decode", |bencher| {
        bencher.iter(|| {
            black_box(Request::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_request(c: &mut Criterion) {
    let data = Request {
        id: Default::default(),
        metadata: Some(Metadata::default()),
    };

    c.bench_function("request/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(Request::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_metadata(c: &mut Criterion) {
    let data = Metadata {
        user_id: Default::default(),
        username: Default::default(),
    };

    c.bench_function("metadata/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_metadata(c: &mut Criterion) {
    let data = Metadata {
        user_id: Default::default(),
        username: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("metadata/decode", |bencher| {
        bencher.iter(|| {
            black_box(Metadata::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_metadata(c: &mut Criterion) {
    let data = Metadata {
        user_id: Default::default(),
        username: Default::default(),
    };

    c.bench_function("metadata/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(Metadata::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_config(c: &mut Criterion) {
    let data = Config {
        name: Default::default(),
        database: Some(DatabaseConfig::default()),
        cache: Some(CacheConfig::default()),
    };

    c.bench_function("config/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_config(c: &mut Criterion) {
    let data = Config {
        name: Default::default(),
        database: Some(DatabaseConfig::default()),
        cache: Some(CacheConfig::default()),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("config/decode", |bencher| {
        bencher.iter(|| {
            black_box(Config::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_config(c: &mut Criterion) {
    let data = Config {
        name: Default::default(),
        database: Some(DatabaseConfig::default()),
        cache: Some(CacheConfig::default()),
    };

    c.bench_function("config/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(Config::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_database_config(c: &mut Criterion) {
    let data = DatabaseConfig {
        host: Default::default(),
        port: Default::default(),
    };

    c.bench_function("database_config/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_database_config(c: &mut Criterion) {
    let data = DatabaseConfig {
        host: Default::default(),
        port: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("database_config/decode", |bencher| {
        bencher.iter(|| {
            black_box(DatabaseConfig::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_database_config(c: &mut Criterion) {
    let data = DatabaseConfig {
        host: Default::default(),
        port: Default::default(),
    };

    c.bench_function("database_config/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(DatabaseConfig::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_cache_config(c: &mut Criterion) {
    let data = CacheConfig {
        size_mb: Default::default(),
        ttl_seconds: Default::default(),
    };

    c.bench_function("cache_config/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_cache_config(c: &mut Criterion) {
    let data = CacheConfig {
        size_mb: Default::default(),
        ttl_seconds: Default::default(),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("cache_config/decode", |bencher| {
        bencher.iter(|| {
            black_box(CacheConfig::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_cache_config(c: &mut Criterion) {
    let data = CacheConfig {
        size_mb: Default::default(),
        ttl_seconds: Default::default(),
    };

    c.bench_function("cache_config/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(CacheConfig::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_document(c: &mut Criterion) {
    let data = Document {
        id: Default::default(),
        tags: Some(TagList::default()),
    };

    c.bench_function("document/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_document(c: &mut Criterion) {
    let data = Document {
        id: Default::default(),
        tags: Some(TagList::default()),
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("document/decode", |bencher| {
        bencher.iter(|| {
            black_box(Document::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_document(c: &mut Criterion) {
    let data = Document {
        id: Default::default(),
        tags: Some(TagList::default()),
    };

    c.bench_function("document/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(Document::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_encode_tag_list(c: &mut Criterion) {
    let data = TagList {
        items: vec![Default::default(), Default::default(), Default::default()],
    };

    c.bench_function("tag_list/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_tag_list(c: &mut Criterion) {
    let data = TagList {
        items: vec![Default::default(), Default::default(), Default::default()],
    };

    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("tag_list/decode", |bencher| {
        bencher.iter(|| {
            black_box(TagList::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_tag_list(c: &mut Criterion) {
    let data = TagList {
        items: vec![Default::default(), Default::default(), Default::default()],
    };

    c.bench_function("tag_list/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(TagList::decode_from_slice(&buf).unwrap());
        });
    });
}

criterion_group!(
    benches,
    bench_encode_request,
    bench_decode_request,
    bench_roundtrip_request,
    bench_encode_metadata,
    bench_decode_metadata,
    bench_roundtrip_metadata,
    bench_encode_config,
    bench_decode_config,
    bench_roundtrip_config,
    bench_encode_database_config,
    bench_decode_database_config,
    bench_roundtrip_database_config,
    bench_encode_cache_config,
    bench_decode_cache_config,
    bench_roundtrip_cache_config,
    bench_encode_document,
    bench_decode_document,
    bench_roundtrip_document,
    bench_encode_tag_list,
    bench_decode_tag_list,
    bench_roundtrip_tag_list
);

criterion_main!(benches);
