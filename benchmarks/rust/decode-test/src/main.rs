// Compare different decode strategies for misaligned u32 arrays
use std::time::Instant;
use std::convert::TryInto;

fn decode_chunks_exact(bytes: &[u8]) -> Vec<u32> {
    bytes.chunks_exact(4)
        .map(|chunk| u32::from_le_bytes(chunk.try_into().unwrap()))
        .collect()
}

fn decode_pod_read(bytes: &[u8]) -> Vec<u32> {
    let mut result = Vec::with_capacity(bytes.len() / 4);
    let mut offset = 0;
    while offset + 4 <= bytes.len() {
        result.push(bytemuck::pod_read_unaligned(&bytes[offset..offset+4]));
        offset += 4;
    }
    result
}

fn decode_loop_manual(bytes: &[u8]) -> Vec<u32> {
    let count = bytes.len() / 4;
    let mut result = Vec::with_capacity(count);
    let mut offset = 0;
    for _ in 0..count {
        let val = u32::from_le_bytes(bytes[offset..offset+4].try_into().unwrap());
        result.push(val);
        offset += 4;
    }
    result
}

fn main() {
    // Create test data (20 bytes = 5 u32s, misaligned)
    let data: Vec<u8> = (0..20).collect();
    
    let iterations = 1_000_000;
    
    // Warmup
    for _ in 0..1000 {
        let _ = decode_chunks_exact(&data);
        let _ = decode_pod_read(&data);
        let _ = decode_loop_manual(&data);
    }
    
    // Benchmark chunks_exact
    let start = Instant::now();
    for _ in 0..iterations {
        let result = decode_chunks_exact(&data);
        std::hint::black_box(result);
    }
    let chunks_ns = start.elapsed().as_nanos() / iterations;
    
    // Benchmark pod_read_unaligned
    let start = Instant::now();
    for _ in 0..iterations {
        let result = decode_pod_read(&data);
        std::hint::black_box(result);
    }
    let pod_ns = start.elapsed().as_nanos() / iterations;
    
    // Benchmark manual loop
    let start = Instant::now();
    for _ in 0..iterations {
        let result = decode_loop_manual(&data);
        std::hint::black_box(result);
    }
    let loop_ns = start.elapsed().as_nanos() / iterations;
    
    println!("=== Decode Strategy Comparison (5 u32s, {} iterations) ===", iterations);
    println!("chunks_exact:      {} ns/op", chunks_ns);
    println!("pod_read_unaligned: {} ns/op", pod_ns);
    println!("manual loop:       {} ns/op", loop_ns);
    
    // Verify all produce same result
    let r1 = decode_chunks_exact(&data);
    let r2 = decode_pod_read(&data);
    let r3 = decode_loop_manual(&data);
    assert_eq!(r1, r2);
    assert_eq!(r2, r3);
    println!("\n✓ All methods produce identical results");
}
