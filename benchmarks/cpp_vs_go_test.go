package benchmarks
package benchmarks

// C++ vs Go benchmark comparison

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"

	arrayspkg "github.com/shaban/serial-data-protocol/testdata/arrays/go"
	audiounit "github.com/shaban/serial-data-protocol/testdata/audiounit/go"
	optionalpkg "github.com/shaban/serial-data-protocol/testdata/optional/go"
	primitives "github.com/shaban/serial-data-protocol/testdata/primitives/go"
)

// Helper to run C++ benchmark and get ns/op
func runCppBench(schema, operation string, iterations int) (int64, error) {
	cmd := exec.Command("./cpp_bench", schema, operation, strconv.Itoa(iterations))
	cmd.Dir = "."
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	nsStr := strings.TrimSpace(string(output))
	ns, err := strconv.ParseInt(nsStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return ns, nil
}

// ============= Primitives Benchmarks =============

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

func BenchmarkCpp_Primitives_Encode(b *testing.B) {
	ns, err := runCppBench("primitives", "encode", 10000000)
	if err != nil {
		b.Fatalf("C++ benchmark failed: %v", err)
	}
	b.ReportMetric(float64(ns), "ns/op")
}

func BenchmarkCpp_Primitives_Decode(b *testing.B) {
	ns, err := runCppBench("primitives", "decode", 10000000)
	if err != nil {
		b.Fatalf("C++ benchmark failed: %v", err)
	}
	b.ReportMetric(float64(ns), "ns/op")
}

// ============= Arrays Benchmarks =============

func BenchmarkGo_Arrays_Encode(b *testing.B) {
	data := arrayspkg.ArraysOfPrimitives{
		U8Array:   []uint8{1, 2, 3, 4, 5},
		U32Array:  []uint32{1000, 2000, 3000, 4000},
		F64Array:  []float64{10.5, 20.5, 30.5},
		BoolArray: []bool{true, false, true, true, false},
		StrArray:  []string{"Hello", "World", "Go", "Arrays"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := arrayspkg.EncodeArraysOfPrimitives(&data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGo_Arrays_Decode(b *testing.B) {
	data := arrayspkg.ArraysOfPrimitives{
		U8Array:   []uint8{1, 2, 3, 4, 5},
		U32Array:  []uint32{1000, 2000, 3000, 4000},
		F64Array:  []float64{10.5, 20.5, 30.5},
		BoolArray: []bool{true, false, true, true, false},
		StrArray:  []string{"Hello", "World", "Go", "Arrays"},
	}

	encoded, _ := arrayspkg.EncodeArraysOfPrimitives(&data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded arrayspkg.ArraysOfPrimitives
		if err := arrayspkg.DecodeArraysOfPrimitives(&decoded, encoded); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCpp_Arrays_Encode(b *testing.B) {
	ns, err := runCppBench("arrays", "encode", 10000000)
	if err != nil {
		b.Fatalf("C++ benchmark failed: %v", err)
	}
	b.ReportMetric(float64(ns), "ns/op")
}

func BenchmarkCpp_Arrays_Decode(b *testing.B) {
	ns, err := runCppBench("arrays", "decode", 10000000)
	if err != nil {
		b.Fatalf("C++ benchmark failed: %v", err)
	}
	b.ReportMetric(float64(ns), "ns/op")
}

// ============= AudioUnit (Nested) Benchmarks =============

func BenchmarkGo_AudioUnit_Encode(b *testing.B) {
	data := audiounit.PluginRegistry{
		TotalPluginCount:    2,
		TotalParameterCount: 5,
		Plugins: []audiounit.Plugin{
			{
				Name:           "Reverb FX",
				ManufacturerID: "ACME",
				ComponentType:  "aufx",
				ComponentSubtype: "rvb1",
				Parameters: []audiounit.Parameter{
					{
						Address:      0,
						DisplayName:  "Room Size",
						Identifier:   "roomSize",
						Unit:         "percent",
						MinValue:     0.0,
						MaxValue:     100.0,
						DefaultValue: 50.0,
						CurrentValue: 75.0,
						RawFlags:     0,
						IsWritable:   true,
						CanRamp:      false,
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := audiounit.EncodePluginRegistry(&data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGo_AudioUnit_Decode(b *testing.B) {
	data := audiounit.PluginRegistry{
		TotalPluginCount:    2,
		TotalParameterCount: 5,
		Plugins: []audiounit.Plugin{
			{
				Name:           "Reverb FX",
				ManufacturerID: "ACME",
				ComponentType:  "aufx",
				ComponentSubtype: "rvb1",
				Parameters: []audiounit.Parameter{
					{
						Address:      0,
						DisplayName:  "Room Size",
						Identifier:   "roomSize",
						Unit:         "percent",
						MinValue:     0.0,
						MaxValue:     100.0,
						DefaultValue: 50.0,
						CurrentValue: 75.0,
						RawFlags:     0,
						IsWritable:   true,
						CanRamp:      false,
					},
				},
			},
		},
	}

	encoded, _ := audiounit.EncodePluginRegistry(&data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded audiounit.PluginRegistry
		if err := audiounit.DecodePluginRegistry(&decoded, encoded); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCpp_AudioUnit_Encode(b *testing.B) {
	ns, err := runCppBench("audiounit", "encode", 1000000)
	if err != nil {
		b.Fatalf("C++ benchmark failed: %v", err)
	}
	b.ReportMetric(float64(ns), "ns/op")
}

func BenchmarkCpp_AudioUnit_Decode(b *testing.B) {
	ns, err := runCppBench("audiounit", "decode", 1000000)
	if err != nil {
		b.Fatalf("C++ benchmark failed: %v", err)
	}
	b.ReportMetric(float64(ns), "ns/op")
}

// ============= Optional Benchmarks =============

func BenchmarkGo_Optional_Encode(b *testing.B) {
	db := optionalpkg.DatabaseConfig{
		Host: "db.example.com",
		Port: 5432,
	}
	data := optionalpkg.Config{
		Name:     "production",
		Database: &db,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := optionalpkg.EncodeConfig(&data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGo_Optional_Decode(b *testing.B) {
	db := optionalpkg.DatabaseConfig{
		Host: "db.example.com",
		Port: 5432,
	}
	data := optionalpkg.Config{
		Name:     "production",
		Database: &db,
	}

	encoded, _ := optionalpkg.EncodeConfig(&data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded optionalpkg.Config
		if err := optionalpkg.DecodeConfig(&decoded, encoded); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCpp_Optional_Encode(b *testing.B) {
	ns, err := runCppBench("optional", "encode", 10000000)
	if err != nil {
		b.Fatalf("C++ benchmark failed: %v", err)
	}
	b.ReportMetric(float64(ns), "ns/op")
}

func BenchmarkCpp_Optional_Decode(b *testing.B) {
	ns, err := runCppBench("optional", "decode", 10000000)
	if err != nil {
		b.Fatalf("C++ benchmark failed: %v", err)
	}
	b.ReportMetric(float64(ns), "ns/op")
}
