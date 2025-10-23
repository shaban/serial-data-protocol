# Include shared variables from project root
MKFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
PROJECT_ROOT := $(abspath $(MKFILE_DIR)/..)
include $(PROJECT_ROOT)/Makefile.vars

.PHONY: all bench bench-go bench-cpp bench-rust bench-cpp-message bench-rust-message clean help

all: bench

# Go benchmarks
bench-go:
	@echo "=== Go SDP Benchmarks ==="
	@cd go && go test -bench=. -run=^\$\$ -benchtime=1s -benchmem

bench-go-arrays:
	@echo "=== Go SDP Arrays Benchmarks (Bulk Optimization) ==="
	@cd go && go test -bench=BenchmarkGo_SDP_Arrays -run=^\$\$ -benchtime=1s -benchmem

bench-go-byte:
	@echo "=== Go SDP Byte Mode Benchmarks ==="
	@cd go && go test -bench=BenchmarkGo_SDP_AudioUnit[^_] -run=^\$\$ -benchtime=1s -benchmem

bench-go-message:
	@echo "=== Go SDP Message Mode Benchmarks ==="
	@cd go && go test -bench=BenchmarkGo_SDP_AudioUnit_Message -run=^\$\$ -benchtime=1s -benchmem

# C++ benchmarks
bench-cpp-arrays: cpp/bytemode/bench_arrays
	@echo "=== C++ SDP Arrays Benchmarks (Bulk Optimization) ==="
	@cd $(PROJECT_ROOT) && $(MKFILE_DIR)/cpp/bytemode/bench_arrays 10000

cpp/bytemode/bench_arrays: cpp/bytemode/bench_arrays.cpp
	@echo "Building C++ Arrays byte mode benchmark..."
	@g++ $(CPP_BUILD_FLAGS) \
		-I$(GENERATED_CPP)/arrays \
		-o $(MKFILE_DIR)/cpp/bytemode/bench_arrays \
		$(MKFILE_DIR)/cpp/bytemode/bench_arrays.cpp \
		$(GENERATED_CPP)/arrays/encode.cpp \
		$(GENERATED_CPP)/arrays/decode.cpp

bench-cpp-message: cpp/messagemode/bench_audiounit
	@echo "=== C++ SDP Message Mode Benchmarks ==="
	@cd $(PROJECT_ROOT) && $(MKFILE_DIR)/cpp/messagemode/bench_audiounit 10000

cpp/messagemode/bench_audiounit: cpp/messagemode/bench_audiounit.cpp
	@echo "Building C++ AudioUnit message mode benchmark..."
	@g++ $(CPP_BUILD_FLAGS) \
		-I$(GENERATED_CPP)/audiounit \
		-o $(MKFILE_DIR)/cpp/messagemode/bench_audiounit \
		$(MKFILE_DIR)/cpp/messagemode/bench_audiounit.cpp \
		$(GENERATED_CPP)/audiounit/encode.cpp \
		$(GENERATED_CPP)/audiounit/decode.cpp \
		$(GENERATED_CPP)/audiounit/message_encode.cpp \
		$(GENERATED_CPP)/audiounit/message_decode.cpp

bench-cpp: bench-cpp-arrays bench-cpp-message

# Rust benchmarks
bench-rust-byte:
	@echo "=== Rust SDP Byte Mode Benchmarks (Criterion) ==="
	@echo "Building and running Rust Arrays benchmark with statistical analysis..."
	@cd $(GENERATED_RUST)/arrays && cargo build $(RUST_BUILD_FLAGS)
	@cd $(MKFILE_DIR)/rust/bytemode && cargo bench --bench arrays_criterion -- --output-format bencher

bench-rust-byte-simple:
	@echo "=== Rust SDP Byte Mode Benchmarks (Simple) ==="
	@echo "Building and running Rust Arrays benchmark..."
	@cd $(GENERATED_RUST)/arrays && cargo build $(RUST_BUILD_FLAGS)
	@cd $(MKFILE_DIR)/rust/bytemode && cargo run --release --bin arrays_benchmark

bench-rust-message:
	@echo "=== Rust SDP Message Mode Benchmarks ==="
	@echo "Building and running Rust AudioUnit benchmark..."
	@cd $(GENERATED_RUST)/audiounit && cargo build $(RUST_BUILD_FLAGS)
	@cd $(MKFILE_DIR)/rust/messagemode && \
		AUDIOUNIT_DATA=$(BINARIES_DIR)/audiounit.sdpb \
		AUDIOUNIT_LIB=$(GENERATED_RUST)/audiounit/target/release \
		cargo run --release --bin audiounit_benchmark

bench-rust: bench-rust-byte bench-rust-message

bench: bench-go bench-cpp bench-rust
	@echo "✓ Benchmark suite complete"

clean:
	@echo "Cleaning benchmark binaries..."
	@rm -f cpp/bytemode/bench_* cpp/messagemode/bench_*

help:
	@echo "SDP Benchmark Suite"
	@echo ""
	@echo "Targets:"
	@echo "  bench              - Run all benchmarks (Go + C++ + Rust)"
	@echo "  bench-go           - Run Go benchmarks"
	@echo "  bench-go-byte      - Go byte mode only"
	@echo "  bench-go-message   - Go message mode only"
	@echo "  bench-cpp          - Run C++ benchmarks"
	@echo "  bench-cpp-message  - C++ message mode (AudioUnit)"
	@echo "  bench-rust         - Run Rust benchmarks"
	@echo "  bench-rust-message - Rust message mode (AudioUnit)"
	@echo "  clean              - Remove compiled binaries"
	@echo "  help               - Show this message"
