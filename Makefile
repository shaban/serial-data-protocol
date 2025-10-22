# SDP - Serial Data Protocol
# Unified build, test, and benchmark orchestration

# Include shared variables
include Makefile.vars

.PHONY: help build generate verify-generated test test-go test-cpp test-rust test-swift test-wire benchmark clean

# Default target
help:
	@echo "SDP - Serial Data Protocol"
	@echo "=========================="
	@echo ""
	@echo "Build targets:"
	@echo "  make build           - Build sdp-gen and sdp-encode"
	@echo "  make generate        - Generate all code from schemas (clean slate)"
	@echo "  make verify-generated - Verify generated code hasn't been tampered with"
	@echo ""
	@echo "Test targets:"
	@echo "  make test            - Run all tests (Go, C++, Rust)"
	@echo "  make test-go         - Run Go tests only"
	@echo "  make test-cpp        - Run C++ tests only"
	@echo "  make test-rust       - Run Rust tests only"
	@echo "  make test-swift      - Run Swift tests only"
	@echo "  make test-wire       - Verify cross-language wire format compatibility"
	@echo ""
	@echo "Benchmark targets:"
	@echo "  make benchmark       - Run all benchmarks (Go, C++, Rust)"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean           - Clean generated code and build artifacts"
	@echo ""
	@echo "Examples:"
	@echo "  make generate        # Regenerate all code from schemas"
	@echo "  make test            # Run full test suite"
	@echo "  make benchmark       # Run all benchmarks"

# Build tools
build: $(SDP_GEN) $(SDP_ENCODE)

$(SDP_GEN):
	@echo "Building sdp-gen..."
	@go build -o $(SDP_GEN) ./cmd/sdp-gen

$(SDP_ENCODE):
	@echo "Building sdp-encode..."
	@go build -o $(SDP_ENCODE) ./cmd/sdp-encode

# Generate all code from schemas (clean slate)
generate: build
	@echo "Generating code from schemas..."
	@echo "Cleaning previous generated code..."
	@rm -rf $(GENERATED_DIR)/*
	@mkdir -p $(GENERATED_GO) $(GENERATED_CPP) $(GENERATED_RUST) $(GENERATED_SWIFT)
	@echo ""
	@echo "Generating from official schemas..."
	@for schema in $(SCHEMAS_DIR)/*.sdp; do \
		name=$$(basename $$schema .sdp); \
		echo "  $$name.sdp -> Go/C++/Rust/Swift"; \
		$(SDP_GEN) -schema $$schema -output $(GENERATED_GO)/$$name -lang go || exit 1; \
		$(SDP_GEN) -schema $$schema -output $(GENERATED_CPP)/$$name -lang cpp || exit 1; \
		$(SDP_GEN) -schema $$schema -output $(GENERATED_RUST)/$$name -lang rust || exit 1; \
		$(SDP_GEN) -schema $$schema -output $(GENERATED_SWIFT)/$$name -lang swift || exit 1; \
	done
	@echo ""
	@echo "✓ Code generation complete"
	@echo "  Generated: $(GENERATED_DIR)/{go,cpp,rust,swift}/*"

# Verify generated code hasn't been manually edited
verify-generated:
	@echo "Verifying generated code integrity..."
	@if git diff --quiet $(GENERATED_DIR); then \
		echo "✓ Generated code is clean (no manual edits)"; \
	else \
		echo "❌ ERROR: Generated code has been manually edited!"; \
		echo ""; \
		echo "Modified files:"; \
		git diff --name-only $(GENERATED_DIR); \
		echo ""; \
		echo "Run 'make generate' to restore clean generated code."; \
		exit 1; \
	fi

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
	@echo "Running benchmark suite..."
	@cd benchmarks && $(MAKE) bench

# Clean generated code and artifacts
clean:
	@echo "Cleaning generated code and artifacts..."
	@rm -rf $(GENERATED_DIR)/*
	@rm -f $(SDP_GEN) $(SDP_ENCODE)
	@cd benchmarks && $(MAKE) clean
	@echo "✓ Clean complete"
