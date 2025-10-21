package benchmarks

import (
	"os"
	"testing"

	audiounit "github.com/shaban/serial-data-protocol/testdata/audiounit/go"
)

// Test data loaded once in init() and reused across all benchmarks
// testData: in-memory structs for encode benchmarks (loaded from JSON)
// testDataSDP: pre-encoded SDP bytes for decode benchmarks (loaded from .sdpb)
var testData audiounit.PluginRegistry
var testDataSDP []byte

func init() {
	// Load .sdpb binary for decode benchmarks (measures decode performance only)
	sdpbData, err := os.ReadFile("../testdata/audiounit.sdpb")
	if err != nil {
		panic("audiounit.sdpb not found: " + err.Error())
	}
	testDataSDP = sdpbData

	// Decode the .sdpb to get struct for encode benchmarks
	// This way we avoid JSON parsing entirely - we have a canonical binary format
	err = audiounit.DecodePluginRegistry(&testData, testDataSDP)
	if err != nil {
		panic("Failed to decode audiounit.sdpb: " + err.Error())
	}
}

// ============================================================================
// SDP Go Implementation Benchmarks - AudioUnit Schema
// Schema: audiounit.sdp (PluginRegistry with 62 plugins, 1,759 parameters)
// Data: audiounit.sdpb (110KB binary)
// ============================================================================

func BenchmarkGo_SDP_AudioUnit_Encode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		encoded, err := audiounit.EncodePluginRegistry(&testData)
		if err != nil {
			b.Fatal(err)
		}
		if len(encoded) == 0 {
			b.Fatal("empty encoding")
		}
	}
}

func BenchmarkGo_SDP_AudioUnit_Decode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded audiounit.PluginRegistry
		err := audiounit.DecodePluginRegistry(&decoded, testDataSDP)
		if err != nil {
			b.Fatal(err)
		}
		if decoded.TotalPluginCount != testData.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}

func BenchmarkGo_SDP_AudioUnit_Roundtrip(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Encode
		encoded, err := audiounit.EncodePluginRegistry(&testData)
		if err != nil {
			b.Fatal(err)
		}

		// Decode
		var decoded audiounit.PluginRegistry
		err = audiounit.DecodePluginRegistry(&decoded, encoded)
		if err != nil {
			b.Fatal(err)
		}

		// Verify
		if decoded.TotalPluginCount != testData.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}
