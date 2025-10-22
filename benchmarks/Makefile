.PHONY: all bench bench-go bench-cpp clean help

all: bench

bench-go:
	@echo "=== Go SDP Benchmarks ==="
	@cd go && go test -bench=. -run=^\$\$ -benchtime=1s -benchmem

bench-go-byte:
	@echo "=== Go SDP Byte Mode Benchmarks ==="
	@cd go && go test -bench=BenchmarkGo_SDP_AudioUnit[^_] -run=^\$\$ -benchtime=1s -benchmem

bench-go-message:
	@echo "=== Go SDP Message Mode Benchmarks ==="
	@cd go && go test -bench=BenchmarkGo_SDP_AudioUnit_Message -run=^\$\$ -benchtime=1s -benchmem

bench: bench-go
	@echo "Benchmark suite complete"

clean:
	@echo "Cleaning benchmark binaries..."
	@rm -f cpp/bytemode/bench_* cpp/messagemode/bench_*

help:
	@echo "SDP Benchmark Suite"
	@echo ""
	@echo "Targets:"
	@echo "  bench            - Run all benchmarks"
	@echo "  bench-go         - Run Go benchmarks"
	@echo "  bench-go-byte    - Go byte mode only"
	@echo "  bench-go-message - Go message mode only"
	@echo "  clean            - Remove compiled binaries"
	@echo "  help             - Show this message"
