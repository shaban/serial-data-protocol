package benchmarks

import (
	"testing"

	audiounit "github.com/shaban/serial-data-protocol/testdata/go/audiounit"
)

// ============================================================================
// SDP Go Implementation - MESSAGE MODE Benchmarks
// Schema: audiounit.sdp (PluginRegistry with 62 plugins, 1,759 parameters)
// Data: 110KB AudioUnit data
//
// These benchmarks measure the overhead of message mode on REAL data,
// not just primitives. Critical for verifying claims before implementing
// message mode in C++/Rust.
// ============================================================================

func BenchmarkGo_SDP_AudioUnit_Message_Encode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		encoded, err := audiounit.EncodePluginRegistryMessage(&testData)
		if err != nil {
			b.Fatal(err)
		}
		if len(encoded) == 0 {
			b.Fatal("empty encoding")
		}
	}
}

func BenchmarkGo_SDP_AudioUnit_Message_Decode(b *testing.B) {
	// Encode once in message mode to get test data
	testDataMessage, err := audiounit.EncodePluginRegistryMessage(&testData)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		decoded, err := audiounit.DecodePluginRegistryMessage(testDataMessage)
		if err != nil {
			b.Fatal(err)
		}
		if decoded.TotalPluginCount != testData.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}

func BenchmarkGo_SDP_AudioUnit_Message_Roundtrip(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Encode in message mode
		encoded, err := audiounit.EncodePluginRegistryMessage(&testData)
		if err != nil {
			b.Fatal(err)
		}

		// Decode from message mode
		decoded, err := audiounit.DecodePluginRegistryMessage(encoded)
		if err != nil {
			b.Fatal(err)
		}

		// Verify
		if decoded.TotalPluginCount != testData.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}

// BenchmarkGo_SDP_AudioUnit_Message_Dispatcher measures the overhead
// of using the type-safe dispatcher (DecodeMessage) vs direct decoding.
// This is what users would actually use in IPC scenarios.
func BenchmarkGo_SDP_AudioUnit_Message_Dispatcher(b *testing.B) {
	testDataMessage, err := audiounit.EncodePluginRegistryMessage(&testData)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		decoded, err := audiounit.DecodeMessage(testDataMessage)
		if err != nil {
			b.Fatal(err)
		}
		registry, ok := decoded.(*audiounit.PluginRegistry)
		if !ok {
			b.Fatal("wrong type")
		}
		if registry.TotalPluginCount != testData.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}

// BenchmarkGo_SDP_AudioUnit_Message_HeaderOverhead measures JUST the
// overhead of the 10-byte message header on encoding.
// Compare: EncodePluginRegistry vs EncodePluginRegistryMessage
func BenchmarkGo_SDP_AudioUnit_Message_HeaderOverhead(b *testing.B) {
	// First measure byte mode
	b.Run("ByteMode", func(b *testing.B) {
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
	})

	// Then measure message mode
	b.Run("MessageMode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			encoded, err := audiounit.EncodePluginRegistryMessage(&testData)
			if err != nil {
				b.Fatal(err)
			}
			if len(encoded) == 0 {
				b.Fatal("empty encoding")
			}
		}
	})
}
