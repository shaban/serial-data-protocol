package benchmarks

// Cross-language benchmark comparison
// These benchmarks measure Rust performance from Go, enabling direct comparison

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	audiounit "github.com/shaban/serial-data-protocol/testdata/audiounit/go"
	primitives "github.com/shaban/serial-data-protocol/testdata/primitives/go"
)

// CriterionEstimates represents Criterion's estimates.json structure
type CriterionEstimates struct {
	Mean struct {
		PointEstimate float64 `json:"point_estimate"`
	} `json:"mean"`
}

// runCriterionBench runs Criterion benchmarks for a package and returns timing in nanoseconds
// Returns map of benchmark_name -> nanoseconds (mean)
func runCriterionBench(packagePath string) (map[string]float64, error) {
	// Run cargo bench to generate fresh results
	cmd := exec.Command("cargo", "bench", "--quiet")
	cmd.Dir = packagePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("cargo bench failed: %w\nOutput: %s", err, output)
	}

	// Parse estimates.json files from target/criterion/*/new/estimates.json
	results := make(map[string]float64)
	criterionDir := filepath.Join(packagePath, "target", "criterion")

	entries, err := os.ReadDir(criterionDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read criterion directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		benchName := entry.Name()
		estimatesPath := filepath.Join(criterionDir, benchName, "new", "estimates.json")

		data, err := os.ReadFile(estimatesPath)
		if err != nil {
			continue // Skip if estimates.json doesn't exist
		}

		var estimates CriterionEstimates
		if err := json.Unmarshal(data, &estimates); err != nil {
			continue // Skip malformed JSON
		}

		// Clean up benchmark name (remove package prefix if present)
		cleanName := benchName
		if idx := strings.LastIndex(benchName, "_"); idx != -1 {
			cleanName = benchName[idx+1:] // e.g., "all_primitives_encode" -> "encode"
		}

		results[cleanName] = estimates.Mean.PointEstimate
	}

	return results, nil
}

// Helper to run Swift benchmark binary and parse timing
func runSwiftBench(command string, args ...string) int64 {
	// Compile swift_bench if needed (release mode with ALL optimizations)
	if _, err := os.Stat("./swift_bench"); os.IsNotExist(err) {
		cmd := exec.Command("swiftc",
			"-O",
			"-whole-module-optimization",
			"-swift-version", "5",
			"-cross-module-optimization",
			"-remove-runtime-asserts",
			"-enforce-exclusivity=unchecked",
			"swift_bench.swift",
			"-o", "swift_bench")
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

// Benchmark: Rust encode primitives (using Criterion)
func BenchmarkRust_Primitives_Encode(b *testing.B) {
	packagePath := "../testdata/primitives/rust"
	
	results, err := runCriterionBench(packagePath)
	if err != nil {
		b.Fatalf("Failed to run Criterion benchmark: %v", err)
	}

	encodeNs, ok := results["encode"]
	if !ok {
		b.Fatalf("encode benchmark not found in Criterion results")
	}

	// Report the Criterion measurement (Criterion already measured many iterations)
	b.ReportMetric(encodeNs, "ns/op")
}

// Benchmark: Rust decode primitives (using Criterion)
func BenchmarkRust_Primitives_Decode(b *testing.B) {
	packagePath := "../testdata/primitives/rust"
	
	results, err := runCriterionBench(packagePath)
	if err != nil {
		b.Fatalf("Failed to run Criterion benchmark: %v", err)
	}

	decodeNs, ok := results["decode"]
	if !ok {
		b.Fatalf("decode benchmark not found in Criterion results")
	}

	// Report the Criterion measurement
	b.ReportMetric(decodeNs, "ns/op")
}

// Benchmark: Rust roundtrip primitives (using Criterion)
func BenchmarkRust_Primitives_Roundtrip(b *testing.B) {
	packagePath := "../testdata/primitives/rust"
	
	results, err := runCriterionBench(packagePath)
	if err != nil {
		b.Fatalf("Failed to run Criterion benchmark: %v", err)
	}

	roundtripNs, ok := results["roundtrip"]
	if !ok {
		b.Fatalf("roundtrip benchmark not found in Criterion results")
	}

	// Report the Criterion measurement
	b.ReportMetric(roundtripNs, "ns/op")
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

// Benchmark: Rust encode audiounit (using Criterion)
func BenchmarkRust_AudioUnit_Encode(b *testing.B) {
	packagePath := "../testdata/audiounit/rust"
	
	results, err := runCriterionBench(packagePath)
	if err != nil {
		b.Fatalf("Failed to run Criterion benchmark: %v", err)
	}

	encodeNs, ok := results["encode"]
	if !ok {
		b.Fatalf("encode benchmark not found in Criterion results")
	}

	// Report the Criterion measurement
	b.ReportMetric(encodeNs, "ns/op")
}

// Benchmark: Rust decode audiounit (using Criterion)
func BenchmarkRust_AudioUnit_Decode(b *testing.B) {
	packagePath := "../testdata/audiounit/rust"
	
	results, err := runCriterionBench(packagePath)
	if err != nil {
		b.Fatalf("Failed to run Criterion benchmark: %v", err)
	}

	decodeNs, ok := results["decode"]
	if !ok {
		b.Fatalf("decode benchmark not found in Criterion results")
	}

	// Report the Criterion measurement
	b.ReportMetric(decodeNs, "ns/op")
}

// Benchmark: Rust roundtrip audiounit (using Criterion)
func BenchmarkRust_AudioUnit_Roundtrip(b *testing.B) {
	packagePath := "../testdata/audiounit/rust"
	
	results, err := runCriterionBench(packagePath)
	if err != nil {
		b.Fatalf("Failed to run Criterion benchmark: %v", err)
	}

	roundtripNs, ok := results["roundtrip"]
	if !ok {
		b.Fatalf("roundtrip benchmark not found in Criterion results")
	}

	// Report the Criterion measurement
	b.ReportMetric(roundtripNs, "ns/op")
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
