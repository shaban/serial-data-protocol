package benchmarks

import (
	"os"
	"testing"

	arrays "github.com/shaban/serial-data-protocol/testdata/generated/go/arrays"
)

// Arrays benchmark data - loaded once and reused
var arraysTestData arrays.ArraysOfPrimitives
var arraysTestDataSDP []byte

func init() {
	// Load canonical .sdpb binary for decode benchmarks
	sdpbData, err := os.ReadFile("../../testdata/binaries/arrays_primitives.sdpb")
	if err != nil {
		panic("arrays_primitives.sdpb not found: " + err.Error())
	}
	arraysTestDataSDP = sdpbData

	// Decode to get struct for encode benchmarks
	err = arrays.DecodeArraysOfPrimitives(&arraysTestData, arraysTestDataSDP)
	if err != nil {
		panic("Failed to decode arrays_primitives.sdpb: " + err.Error())
	}
}

// ============================================================================
// SDP Go Arrays Benchmarks - Bulk Array Optimization
// Schema: arrays.sdp (ArraysOfPrimitives with u8, u32, f64, str, bool arrays)
// Data: arrays_primitives.sdpb (canonical binary)
// ============================================================================

func BenchmarkGo_SDP_Arrays_Encode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		encoded, err := arrays.EncodeArraysOfPrimitives(&arraysTestData)
		if err != nil {
			b.Fatal(err)
		}
		if len(encoded) == 0 {
			b.Fatal("empty encoding")
		}
	}
}

func BenchmarkGo_SDP_Arrays_Decode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded arrays.ArraysOfPrimitives
		err := arrays.DecodeArraysOfPrimitives(&decoded, arraysTestDataSDP)
		if err != nil {
			b.Fatal(err)
		}
		if len(decoded.U8Array) != len(arraysTestData.U8Array) {
			b.Fatal("decode mismatch")
		}
	}
}

func BenchmarkGo_SDP_Arrays_Roundtrip(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Encode
		encoded, err := arrays.EncodeArraysOfPrimitives(&arraysTestData)
		if err != nil {
			b.Fatal(err)
		}

		// Decode
		var decoded arrays.ArraysOfPrimitives
		err = arrays.DecodeArraysOfPrimitives(&decoded, encoded)
		if err != nil {
			b.Fatal(err)
		}

		// Verify
		if len(decoded.U8Array) != len(arraysTestData.U8Array) {
			b.Fatal("decode mismatch")
		}
	}
}
