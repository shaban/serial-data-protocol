package benchmarks

// Cross-language benchmark comparison
// These benchmarks measure Rust performance from Go, enabling direct comparison

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	audiounit "github.com/shaban/serial-data-protocol/testdata/audiounit/go"
	primitives "github.com/shaban/serial-data-protocol/testdata/primitives/go"
)

// Helper to run Rust benchmark binary and parse timing
func runRustBench(command string, args ...string) int64 {
	// Build rust-bench if needed
	cmd := exec.Command("cargo", "build", "--release", "--bin", "rust-bench")
	cmd.Dir = "../rust/sdp"
	if err := cmd.Run(); err != nil {
		panic("Failed to build rust-bench: " + err.Error())
	}

	// Run benchmark
	allArgs := append([]string{command}, args...)
	cmd = exec.Command("../rust/target/release/rust-bench", allArgs...)
	output, err := cmd.Output()
	if err != nil {
		panic("Failed to run rust-bench: " + err.Error())
	}

	// Parse nanoseconds per operation
	nsStr := strings.TrimSpace(string(output))
	ns, err := strconv.ParseInt(nsStr, 10, 64)
	if err != nil {
		panic("Failed to parse rust-bench output: " + err.Error())
	}

	return ns
}

// Helper to run Swift benchmark binary and parse timing
func runSwiftBench(command string, args ...string) int64 {
	// Compile swift_bench if needed (release mode with optimizations)
	if _, err := os.Stat("./swift_bench"); os.IsNotExist(err) {
		cmd := exec.Command("swiftc", "-O", "-whole-module-optimization", "swift_bench.swift", "-o", "swift_bench")
		if err := cmd.Run(); err != nil {
			panic("Failed to compile swift_bench: " + err.Error())
		}
	}

	// Run benchmark
	allArgs := append([]string{command}, args...)
	cmd := exec.Command("./swift_bench", allArgs...)
	output, err := cmd.Output()
	if err != nil {
		panic("Failed to run swift_bench: " + err.Error())
	}

	// Parse nanoseconds per operation
	nsStr := strings.TrimSpace(string(output))
	ns, err := strconv.ParseInt(nsStr, 10, 64)
	if err != nil {
		panic("Failed to parse swift_bench output: " + err.Error())
	}

	return ns
}

// Benchmark: Go encode primitives
func BenchmarkGo_Primitives_Encode(b *testing.B) {
	data := primitives.AllPrimitives{
		U8Field:   255,
		U16Field:  65535,
		U32Field:  4294967295,
		U64Field:  18446744073709551615,
		I8Field:   -128,
		I16Field:  -32768,
		I32Field:  -2147483648,
		I64Field:  -9223372036854775808,
		F32Field:  3.14159,
		F64Field:  2.718281828459045,
		BoolField: true,
		StrField:  "Hello from Go!",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := primitives.EncodeAllPrimitives(&data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Go decode primitives
func BenchmarkGo_Primitives_Decode(b *testing.B) {
	data := primitives.AllPrimitives{
		U8Field:   255,
		U16Field:  65535,
		U32Field:  4294967295,
		U64Field:  18446744073709551615,
		I8Field:   -128,
		I16Field:  -32768,
		I32Field:  -2147483648,
		I64Field:  -9223372036854775808,
		F32Field:  3.14159,
		F64Field:  2.718281828459045,
		BoolField: true,
		StrField:  "Hello from Go!",
	}

	encoded, _ := primitives.EncodeAllPrimitives(&data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded primitives.AllPrimitives
		if err := primitives.DecodeAllPrimitives(&decoded, encoded); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Rust encode primitives (called from Go)
func BenchmarkRust_Primitives_Encode(b *testing.B) {
	// Run Rust benchmark once to get timing
	ns := runRustBench("encode-primitives", strconv.Itoa(b.N))

	b.ReportMetric(float64(ns), "ns/op")
	b.ReportMetric(float64(b.N)*1e9/float64(ns*int64(b.N)), "ops/sec")
}

// Benchmark: Rust decode primitives (called from Go)
func BenchmarkRust_Primitives_Decode(b *testing.B) {
	// Create test file
	data := primitives.AllPrimitives{
		U8Field:   255,
		U16Field:  65535,
		U32Field:  4294967295,
		U64Field:  18446744073709551615,
		I8Field:   -128,
		I16Field:  -32768,
		I32Field:  -2147483648,
		I64Field:  -9223372036854775808,
		F32Field:  3.14159,
		F64Field:  2.718281828459045,
		BoolField: true,
		StrField:  "Hello from Rust!",
	}

	encoded, _ := primitives.EncodeAllPrimitives(&data)
	tmpFile := "/tmp/bench_primitives.bin"
	os.WriteFile(tmpFile, encoded, 0644)

	b.ResetTimer()

	// Run Rust benchmark
	ns := runRustBench("decode-primitives", tmpFile, strconv.Itoa(b.N))

	b.ReportMetric(float64(ns), "ns/op")
	b.ReportMetric(float64(b.N)*1e9/float64(ns*int64(b.N)), "ops/sec")
}

// Benchmark: Go encode audiounit
func BenchmarkGo_AudioUnit_Encode(b *testing.B) {
	param1 := audiounit.Parameter{
		Address:      0,
		DisplayName:  "Volume",
		Identifier:   "volume",
		Unit:         "dB",
		MinValue:     -96.0,
		MaxValue:     6.0,
		DefaultValue: 0.0,
		CurrentValue: -12.0,
		RawFlags:     0x01,
		IsWritable:   true,
		CanRamp:      true,
	}

	param2 := audiounit.Parameter{
		Address:      1,
		DisplayName:  "Pan",
		Identifier:   "pan",
		Unit:         "%",
		MinValue:     -100.0,
		MaxValue:     100.0,
		DefaultValue: 0.0,
		CurrentValue: 0.0,
		RawFlags:     0x01,
		IsWritable:   true,
		CanRamp:      true,
	}

	plugin1 := audiounit.Plugin{
		Name:             "Test Synth",
		ManufacturerId:   "TEST",
		ComponentType:    "aumu",
		ComponentSubtype: "test",
		Parameters:       []audiounit.Parameter{param1, param2},
	}

	plugin2 := audiounit.Plugin{
		Name:             "Test Effect",
		ManufacturerId:   "TEST",
		ComponentType:    "aumf",
		ComponentSubtype: "tsfx",
		Parameters:       []audiounit.Parameter{param1},
	}

	registry := audiounit.PluginRegistry{
		Plugins:             []audiounit.Plugin{plugin1, plugin2},
		TotalPluginCount:    2,
		TotalParameterCount: 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := audiounit.EncodePluginRegistry(&registry)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Go decode audiounit
func BenchmarkGo_AudioUnit_Decode(b *testing.B) {
	param1 := audiounit.Parameter{
		Address:      0,
		DisplayName:  "Volume",
		Identifier:   "volume",
		Unit:         "dB",
		MinValue:     -96.0,
		MaxValue:     6.0,
		DefaultValue: 0.0,
		CurrentValue: -12.0,
		RawFlags:     0x01,
		IsWritable:   true,
		CanRamp:      true,
	}

	param2 := audiounit.Parameter{
		Address:      1,
		DisplayName:  "Pan",
		Identifier:   "pan",
		Unit:         "%",
		MinValue:     -100.0,
		MaxValue:     100.0,
		DefaultValue: 0.0,
		CurrentValue: 0.0,
		RawFlags:     0x01,
		IsWritable:   true,
		CanRamp:      true,
	}

	plugin1 := audiounit.Plugin{
		Name:             "Test Synth",
		ManufacturerId:   "TEST",
		ComponentType:    "aumu",
		ComponentSubtype: "test",
		Parameters:       []audiounit.Parameter{param1, param2},
	}

	plugin2 := audiounit.Plugin{
		Name:             "Test Effect",
		ManufacturerId:   "TEST",
		ComponentType:    "aumf",
		ComponentSubtype: "tsfx",
		Parameters:       []audiounit.Parameter{param1},
	}

	registry := audiounit.PluginRegistry{
		Plugins:             []audiounit.Plugin{plugin1, plugin2},
		TotalPluginCount:    2,
		TotalParameterCount: 3,
	}

	encoded, _ := audiounit.EncodePluginRegistry(&registry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded audiounit.PluginRegistry
		if err := audiounit.DecodePluginRegistry(&decoded, encoded); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Rust encode audiounit (called from Go)
func BenchmarkRust_AudioUnit_Encode(b *testing.B) {
	ns := runRustBench("encode-audiounit", strconv.Itoa(b.N))

	b.ReportMetric(float64(ns), "ns/op")
	b.ReportMetric(float64(b.N)*1e9/float64(ns*int64(b.N)), "ops/sec")
}

// Benchmark: Rust decode audiounit (called from Go)
func BenchmarkRust_AudioUnit_Decode(b *testing.B) {
	// Create test file using Go-encoded data
	param1 := audiounit.Parameter{
		Address:      0,
		DisplayName:  "Volume",
		Identifier:   "volume",
		Unit:         "dB",
		MinValue:     -96.0,
		MaxValue:     6.0,
		DefaultValue: 0.0,
		CurrentValue: -12.0,
		RawFlags:     0x01,
		IsWritable:   true,
		CanRamp:      true,
	}

	param2 := audiounit.Parameter{
		Address:      1,
		DisplayName:  "Pan",
		Identifier:   "pan",
		Unit:         "%",
		MinValue:     -100.0,
		MaxValue:     100.0,
		DefaultValue: 0.0,
		CurrentValue: 0.0,
		RawFlags:     0x01,
		IsWritable:   true,
		CanRamp:      true,
	}

	plugin1 := audiounit.Plugin{
		Name:             "Test Synth",
		ManufacturerId:   "TEST",
		ComponentType:    "aumu",
		ComponentSubtype: "test",
		Parameters:       []audiounit.Parameter{param1, param2},
	}

	plugin2 := audiounit.Plugin{
		Name:             "Test Effect",
		ManufacturerId:   "TEST",
		ComponentType:    "aumf",
		ComponentSubtype: "tsfx",
		Parameters:       []audiounit.Parameter{param1},
	}

	registry := audiounit.PluginRegistry{
		Plugins:             []audiounit.Plugin{plugin1, plugin2},
		TotalPluginCount:    2,
		TotalParameterCount: 3,
	}

	encoded, _ := audiounit.EncodePluginRegistry(&registry)
	tmpFile := "/tmp/bench_audiounit.bin"
	os.WriteFile(tmpFile, encoded, 0644)

	b.ResetTimer()

	ns := runRustBench("decode-audiounit", tmpFile, strconv.Itoa(b.N))

	b.ReportMetric(float64(ns), "ns/op")
	b.ReportMetric(float64(b.N)*1e9/float64(ns*int64(b.N)), "ops/sec")
}

// ============================================================================
// Swift Benchmarks
// ============================================================================

// Benchmark: Swift encode primitives (called from Go)
func BenchmarkSwift_Primitives_Encode(b *testing.B) {
	// Run Swift benchmark
	ns := runSwiftBench("encode-primitives", strconv.Itoa(b.N))

	b.ReportMetric(float64(ns), "ns/op")
	b.ReportMetric(float64(b.N)*1e9/float64(ns*int64(b.N)), "ops/sec")
}

// Benchmark: Swift decode primitives (called from Go)
func BenchmarkSwift_Primitives_Decode(b *testing.B) {
	// Create test file
	data := primitives.AllPrimitives{
		U8Field:   255,
		U16Field:  65535,
		U32Field:  4294967295,
		U64Field:  18446744073709551615,
		I8Field:   -128,
		I16Field:  -32768,
		I32Field:  -2147483648,
		I64Field:  -9223372036854775808,
		F32Field:  3.14159,
		F64Field:  2.718281828459045,
		BoolField: true,
		StrField:  "Hello from Swift!",
	}

	encoded, _ := primitives.EncodeAllPrimitives(&data)
	tmpFile := "/tmp/bench_primitives.bin"
	os.WriteFile(tmpFile, encoded, 0644)

	b.ResetTimer()

	// Run Swift benchmark
	ns := runSwiftBench("decode-primitives", tmpFile, strconv.Itoa(b.N))

	b.ReportMetric(float64(ns), "ns/op")
	b.ReportMetric(float64(b.N)*1e9/float64(ns*int64(b.N)), "ops/sec")
}
