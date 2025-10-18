.PHONY: all install-tools generate-pb generate-fb generate test bench clean

all: generate

# Install required code generation tools
install-tools:
	@echo "Installing Protocol Buffers compiler..."
	@which protoc || (echo "Please install protoc: brew install protobuf" && exit 1)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "Installing FlatBuffers compiler..."
	@which flatc || (echo "Please install flatc: brew install flatbuffers" && exit 1)

# Generate Protocol Buffers code
generate-pb: install-tools
	@echo "Generating Protocol Buffers code..."
	@mkdir -p pb
	protoc --go_out=pb --go_opt=paths=source_relative audiounit.proto

# Generate FlatBuffers code
generate-fb: install-tools
	@echo "Generating FlatBuffers code..."
	@mkdir -p fb
	flatc --go --gen-onefile --go-namespace fb -o fb audiounit.fbs

# Generate all schemas
generate: generate-pb generate-fb
	@echo "All schemas generated successfully"

# Run tests
test:
	go test -v

# Run benchmarks
bench:
	go test -bench=. -benchmem -benchtime=5s

# Run comparative benchmarks
bench-compare:
	@echo "Running comparative benchmarks (10 iterations)..."
	go test -bench=. -benchmem -count=10 | tee results.txt
	@echo "\nResults saved to results.txt"
	@echo "Install benchstat for statistical comparison: go install golang.org/x/perf/cmd/benchstat@latest"

# Clean generated files
clean:
	rm -rf pb/ fb/ results.txt

# Help
help:
	@echo "Available targets:"
	@echo "  install-tools    - Install protoc and flatc code generators"
	@echo "  generate-pb      - Generate Protocol Buffers code"
	@echo "  generate-fb      - Generate FlatBuffers code"
	@echo "  generate         - Generate all schemas"
	@echo "  test             - Run tests"
	@echo "  bench            - Run benchmarks (5s each)"
	@echo "  bench-compare    - Run comparative benchmarks (10 iterations)"
	@echo "  clean            - Remove generated files"
