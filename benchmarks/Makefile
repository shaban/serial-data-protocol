# Makefile directory (where this Makefile is located)
MKFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
# Project root (parent of benchmarks/)
PROJECT_ROOT := $(abspath $(MKFILE_DIR)/..)

.PHONY: all bench bench-go bench-cpp bench-cpp-message clean help

all: bench

# Go benchmarks
bench-go:
	@echo "=== Go SDP Benchmarks ==="
	@cd go && go test -bench=. -run=^\$\$ -benchtime=1s -benchmem

bench-go-byte:
	@echo "=== Go SDP Byte Mode Benchmarks ==="
	@cd go && go test -bench=BenchmarkGo_SDP_AudioUnit[^_] -run=^\$\$ -benchtime=1s -benchmem

bench-go-message:
	@echo "=== Go SDP Message Mode Benchmarks ==="
	@cd go && go test -bench=BenchmarkGo_SDP_AudioUnit_Message -run=^\$\$ -benchtime=1s -benchmem

# C++ benchmarks
bench-cpp-message: cpp/messagemode/bench_audiounit
	@echo "=== C++ SDP Message Mode Benchmarks ==="
	@cd $(PROJECT_ROOT) && $(MKFILE_DIR)/cpp/messagemode/bench_audiounit 10000

cpp/messagemode/bench_audiounit: cpp/messagemode/bench_audiounit.cpp
	@echo "Building C++ AudioUnit message mode benchmark..."
	@g++ -std=c++17 -O3 -march=native -flto \
		-I$(PROJECT_ROOT)/testdata/cpp/audiounit \
		-o $(MKFILE_DIR)/cpp/messagemode/bench_audiounit \
		$(MKFILE_DIR)/cpp/messagemode/bench_audiounit.cpp \
		$(PROJECT_ROOT)/testdata/cpp/audiounit/encode.cpp \
		$(PROJECT_ROOT)/testdata/cpp/audiounit/decode.cpp \
		$(PROJECT_ROOT)/testdata/cpp/audiounit/message_encode.cpp \
		$(PROJECT_ROOT)/testdata/cpp/audiounit/message_decode.cpp

bench-cpp: bench-cpp-message

bench: bench-go bench-cpp
	@echo "Benchmark suite complete"

clean:
	@echo "Cleaning benchmark binaries..."
	@rm -f cpp/bytemode/bench_* cpp/messagemode/bench_*

help:
	@echo "SDP Benchmark Suite"
	@echo ""
	@echo "Targets:"
	@echo "  bench              - Run all benchmarks (Go + C++)"
	@echo "  bench-go           - Run Go benchmarks"
	@echo "  bench-go-byte      - Go byte mode only"
	@echo "  bench-go-message   - Go message mode only"
	@echo "  bench-cpp          - Run C++ benchmarks"
	@echo "  bench-cpp-message  - C++ message mode (AudioUnit)"
	@echo "  clean              - Remove compiled binaries"
	@echo "  help               - Show this message"
