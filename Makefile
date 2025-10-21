# SDP - Serial Data Protocol
# Unified test orchestration

.PHONY: help test test-go test-cpp test-rust test-swift test-wire benchmark clean

# Default target
help:
	@echo "SDP Test Orchestration"
	@echo "======================"
	@echo ""
	@echo "Available targets:"
	@echo "  make test          - Run all tests (Go, C++, Rust)"
	@echo "  make test-go       - Run Go tests only"
	@echo "  make test-cpp      - Run C++ tests only"
	@echo "  make test-rust     - Run Rust tests only"
	@echo "  make test-swift    - Run Swift tests only"
	@echo "  make test-wire     - Verify cross-language wire format compatibility"
	@echo "  make benchmark     - Run benchmark suite"
	@echo "  make clean         - Clean generated code and test artifacts"
	@echo ""
	@echo "Examples:"
	@echo "  make test          # Run full test suite"
	@echo "  make test-go       # Quick Go-only test"
	@echo "  make test-wire     # Verify wire format compatibility"

# Run all tests
test: test-go test-cpp test-rust
	@echo ""
	@echo "✅ All tests passed!"

# Run Go tests
test-go:
	@./tests/test_go.sh

# Run C++ tests
test-cpp:
	@./tests/test_cpp.sh

# Run Rust tests
test-rust:
	@./tests/test_rust.sh

# Run Swift tests
test-swift:
	@./tests/test_swift.sh

# Verify cross-language wire format compatibility
test-wire:
	@./tests/verify_wire_format.sh

# Run benchmarks
benchmark:
	@./tests/run_benchmarks.sh all

# Clean generated code and artifacts
clean:
	@echo "Cleaning generated code and artifacts..."
	@rm -rf testdata/go/*/
	@rm -rf testdata/cpp/*/
	@rm -rf testdata/rust/*/
	@rm -rf testdata/swift/*/
	@rm -f sdp-gen sdp-encode
	@cd benchmarks && $(MAKE) clean
	@echo "✓ Clean complete"
