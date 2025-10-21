.PHONY: all bench bench-go bench-c bench-cpp build-c build-cpp clean help

all: bench

# Build C benchmark
build-c:
	@echo "Building C benchmark..."
	gcc -std=c99 -O3 -march=native -flto -Wall -Wextra \
		standalone/bench_c.c \
		../testdata/audiounit_c/encode.c \
		../testdata/audiounit_c/decode.c \
		-I../testdata/audiounit_c \
		-o standalone/bench_c

# Build C++ benchmark
build-cpp:
	@echo "Building C++ benchmark..."
	g++ -std=c++17 -O3 -march=native -flto -Wall -Wextra \
		standalone/bench_cpp.cpp \
		../testdata/audiounit_cpp/encode.cpp \
		../testdata/audiounit_cpp/decode.cpp \
		-I../testdata/audiounit_cpp \
		-o standalone/bench_cpp

# Run Go benchmarks
bench-go:
	@echo "=== Go SDP AudioUnit Benchmarks ==="
	@go test -bench=BenchmarkGo -run=^$$ -benchtime=1s

# Run C benchmarks
bench-c: build-c
	@echo "\n=== C SDP AudioUnit Benchmarks ==="
	@./standalone/bench_c 10000

# Run C++ benchmarks
bench-cpp: build-cpp
	@echo "\n=== C++ SDP AudioUnit Benchmarks ==="
	@./standalone/bench_cpp 10000

# Run all benchmarks
bench: bench-go bench-c bench-cpp
	@echo "\n=== Benchmark Summary ==="
	@echo "All benchmarks completed successfully"

# Clean generated files
clean:
	rm -f standalone/bench_c standalone/bench_cpp

# Help
help:
	@echo "Available targets:"
	@echo "  bench         - Run all language benchmarks (Go, C, C++)"
	@echo "  bench-go      - Run Go benchmarks only"
	@echo "  bench-c       - Build and run C benchmarks"
	@echo "  bench-cpp     - Build and run C++ benchmarks"
	@echo "  build-c       - Build C benchmark executable"
	@echo "  build-cpp     - Build C++ benchmark executable"
	@echo "  clean         - Remove compiled binaries"
	@echo ""
	@echo "All benchmarks use ../testdata/audiounit.sdpb (110KB, 62 plugins, 1,759 params)"

