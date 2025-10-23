// Rust byte mode array benchmarks
// Measures encode/decode performance for primitive arrays with bulk optimization
// Uses canonical binary fixture from testdata/binaries/arrays_primitives.sdpb

use sdp_arrays_of_primitives::*;
use std::time::Instant;
use std::fs;

fn load_test_data() -> (ArraysOfPrimitives, Vec<u8>) {
    // Load canonical .sdpb binary
    let binary_path = "../../../testdata/binaries/arrays_primitives.sdpb";
    let binary = fs::read(binary_path)
        .expect("Failed to read arrays_primitives.sdpb - run from benchmarks/rust/bytemode");
    
    // Decode to get struct for encode benchmarks
    let data = ArraysOfPrimitives::decode_from_slice(&binary)
        .expect("Failed to decode arrays_primitives.sdpb");
    
    (data, binary)
}

fn bench_encode(iterations: usize, data: &ArraysOfPrimitives) -> (u128, usize) {
    let mut buf = vec![0u8; 100000]; // Pre-allocate buffer
    let start = Instant::now();
    
    for _ in 0..iterations {
        let size = data.encode_to_slice(&mut buf).unwrap();
        std::hint::black_box(size);
    }
    
    let elapsed = start.elapsed().as_nanos();
    let per_op = elapsed / iterations as u128;
    (per_op, iterations)
}

fn bench_decode(iterations: usize, encoded: &[u8]) -> (u128, usize) {
    let start = Instant::now();
    
    for _ in 0..iterations {
        let decoded = ArraysOfPrimitives::decode_from_slice(encoded).unwrap();
        std::hint::black_box(decoded);
    }
    
    let elapsed = start.elapsed().as_nanos();
    let per_op = elapsed / iterations as u128;
    (per_op, iterations)
}

fn bench_roundtrip(iterations: usize, data: &ArraysOfPrimitives) -> (u128, usize) {
    let mut buf = vec![0u8; 100000];
    let start = Instant::now();
    
    for _ in 0..iterations {
        let size = data.encode_to_slice(&mut buf).unwrap();
        let decoded = ArraysOfPrimitives::decode_from_slice(&buf[..size]).unwrap();
        std::hint::black_box(decoded);
    }
    
    let elapsed = start.elapsed().as_nanos();
    let per_op = elapsed / iterations as u128;
    (per_op, iterations)
}

fn main() {
    println!("=== Rust SDP Byte Mode: Arrays Benchmark ===");
    println!("Schema: arrays.sdp (ArraysOfPrimitives)");
    println!("Data: testdata/binaries/arrays_primitives.sdpb (canonical)");
    println!();
    
    // Load canonical test data
    let (data, binary) = load_test_data();
    
    println!("Loaded {} bytes from canonical fixture", binary.len());
    println!("u8_array: {} elements", data.u8_array.len());
    println!("u32_array: {} elements", data.u32_array.len());
    println!("f64_array: {} elements", data.f64_array.len());
    println!("str_array: {} elements", data.str_array.len());
    println!("bool_array: {} elements", data.bool_array.len());
    println!();
    
    // Warm up
    bench_encode(100, &data);
    bench_decode(100, &binary);
    
    // Actual benchmarks
    let (encode_ns, encode_iters) = bench_encode(10000, &data);
    let (decode_ns, decode_iters) = bench_decode(10000, &binary);
    let (roundtrip_ns, roundtrip_iters) = bench_roundtrip(5000, &data);
    
    println!("Encode:    {:>8} ns/op  ({} iterations)", encode_ns, encode_iters);
    println!("Decode:    {:>8} ns/op  ({} iterations)", decode_ns, decode_iters);
    println!("Roundtrip: {:>8} ns/op  ({} iterations)", roundtrip_ns, roundtrip_iters);
    
    // Verify encoded size matches
    let mut buf = vec![0u8; 100000];
    let size = data.encode_to_slice(&mut buf).unwrap();
    println!();
    println!("Encoded size: {} bytes (canonical: {})", size, binary.len());
    if size != binary.len() {
        println!("⚠️  WARNING: Encoded size doesn't match canonical binary!");
    }
}
