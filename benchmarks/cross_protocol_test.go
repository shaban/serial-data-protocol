package benchmarks

import (
	"os"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"google.golang.org/protobuf/proto"

	fb "github.com/shaban/serial-data-protocol/testdata/flatbuffers/go"
	audiounit "github.com/shaban/serial-data-protocol/testdata/go/audiounit"
	pb "github.com/shaban/serial-data-protocol/testdata/protobuf/go"
)

// ============================================================================
// Cross-Protocol Benchmark Suite
//
// Fair comparison of SDP vs Protocol Buffers vs FlatBuffers using identical
// real-world AudioUnit data (62 plugins, 1,759 parameters, ~110KB).
//
// Methodology:
// - All protocols use the same source data (testData loaded from audiounit.sdpb)
// - Each protocol converts to its own format once during init
// - Benchmarks measure encode/decode/roundtrip for each protocol
// - No cherry-picking, no micro-optimizations
//
// See: benchmarks/RESULTS.md for detailed analysis
// ============================================================================

var (
	// SDP data (from comparison_test.go)
	testDataSDP_Struct audiounit.PluginRegistry
	testDataSDP_Binary []byte

	// Protocol Buffers data
	testDataPB        *pb.PluginRegistry
	testDataPB_Binary []byte

	// FlatBuffers data
	testDataFB_Binary []byte
)

func init() {
	// Load SDP binary data
	sdpbData, err := loadBinaryFile("../testdata/binaries/audiounit.sdpb")
	if err != nil {
		panic("Failed to load audiounit.sdpb: " + err.Error())
	}
	testDataSDP_Binary = sdpbData

	// Decode SDP to get canonical struct data
	err = audiounit.DecodePluginRegistry(&testDataSDP_Struct, testDataSDP_Binary)
	if err != nil {
		panic("Failed to decode audiounit.sdpb: " + err.Error())
	}

	// Convert SDP struct to Protocol Buffers format
	testDataPB = convertSDPToProtobuf(&testDataSDP_Struct)

	// Pre-encode Protocol Buffers for decode benchmarks
	testDataPB_Binary, err = proto.Marshal(testDataPB)
	if err != nil {
		panic("Failed to encode Protocol Buffers: " + err.Error())
	}

	// Convert SDP struct to FlatBuffers format and encode
	testDataFB_Binary = convertSDPToFlatBuffers(&testDataSDP_Struct)
}

// ============================================================================
// Protocol Buffers Benchmarks
// ============================================================================

func BenchmarkProtobuf_Encode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		encoded, err := proto.Marshal(testDataPB)
		if err != nil {
			b.Fatal(err)
		}
		if len(encoded) == 0 {
			b.Fatal("empty encoding")
		}
	}
}

func BenchmarkProtobuf_Decode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoded := &pb.PluginRegistry{}
		err := proto.Unmarshal(testDataPB_Binary, decoded)
		if err != nil {
			b.Fatal(err)
		}
		if decoded.TotalPluginCount != testDataPB.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}

func BenchmarkProtobuf_Roundtrip(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Encode
		encoded, err := proto.Marshal(testDataPB)
		if err != nil {
			b.Fatal(err)
		}

		// Decode
		decoded := &pb.PluginRegistry{}
		err = proto.Unmarshal(encoded, decoded)
		if err != nil {
			b.Fatal(err)
		}

		// Verify
		if decoded.TotalPluginCount != testDataPB.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}

// ============================================================================
// FlatBuffers Benchmarks
// ============================================================================

func BenchmarkFlatBuffers_Encode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		encoded := convertSDPToFlatBuffers(&testDataSDP_Struct)
		if len(encoded) == 0 {
			b.Fatal("empty encoding")
		}
	}
}

func BenchmarkFlatBuffers_Decode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// FlatBuffers "decode" is zero-copy - just access the root
		registry := fb.GetRootAsPluginRegistry(testDataFB_Binary, 0)
		if registry.TotalPluginCount() != testDataSDP_Struct.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}

func BenchmarkFlatBuffers_Roundtrip(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Encode
		encoded := convertSDPToFlatBuffers(&testDataSDP_Struct)

		// Decode (zero-copy access)
		registry := fb.GetRootAsPluginRegistry(encoded, 0)

		// Verify
		if registry.TotalPluginCount() != testDataSDP_Struct.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func loadBinaryFile(path string) ([]byte, error) {
	// Imported from comparison_test.go
	return os.ReadFile(path)
}

func convertSDPToProtobuf(sdp *audiounit.PluginRegistry) *pb.PluginRegistry {
	pbRegistry := &pb.PluginRegistry{
		Plugins:             make([]*pb.Plugin, 0, len(sdp.Plugins)),
		TotalPluginCount:    sdp.TotalPluginCount,
		TotalParameterCount: sdp.TotalParameterCount,
	}

	for _, sdpPlugin := range sdp.Plugins {
		pbPlugin := &pb.Plugin{
			Name:             sdpPlugin.Name,
			ManufacturerId:   sdpPlugin.ManufacturerId,
			ComponentType:    sdpPlugin.ComponentType,
			ComponentSubtype: sdpPlugin.ComponentSubtype,
			Parameters:       make([]*pb.Parameter, 0, len(sdpPlugin.Parameters)),
		}

		for _, sdpParam := range sdpPlugin.Parameters {
			pbParam := &pb.Parameter{
				Address:      sdpParam.Address,
				DisplayName:  sdpParam.DisplayName,
				Identifier:   sdpParam.Identifier,
				Unit:         sdpParam.Unit,
				MinValue:     sdpParam.MinValue,
				MaxValue:     sdpParam.MaxValue,
				DefaultValue: sdpParam.DefaultValue,
				CurrentValue: sdpParam.CurrentValue,
				RawFlags:     sdpParam.RawFlags,
				IsWritable:   sdpParam.IsWritable,
				CanRamp:      sdpParam.CanRamp,
			}
			pbPlugin.Parameters = append(pbPlugin.Parameters, pbParam)
		}

		pbRegistry.Plugins = append(pbRegistry.Plugins, pbPlugin)
	}

	return pbRegistry
}

func convertSDPToFlatBuffers(sdp *audiounit.PluginRegistry) []byte {
	builder := flatbuffers.NewBuilder(1024 * 128) // 128KB initial capacity

	// Build parameters and plugins
	pluginOffsets := make([]flatbuffers.UOffsetT, 0, len(sdp.Plugins))

	for _, sdpPlugin := range sdp.Plugins {
		// Build parameters for this plugin
		paramOffsets := make([]flatbuffers.UOffsetT, 0, len(sdpPlugin.Parameters))

		for _, sdpParam := range sdpPlugin.Parameters {
			displayName := builder.CreateString(sdpParam.DisplayName)
			identifier := builder.CreateString(sdpParam.Identifier)
			unit := builder.CreateString(sdpParam.Unit)

			fb.ParameterStart(builder)
			fb.ParameterAddAddress(builder, sdpParam.Address)
			fb.ParameterAddDisplayName(builder, displayName)
			fb.ParameterAddIdentifier(builder, identifier)
			fb.ParameterAddUnit(builder, unit)
			fb.ParameterAddMinValue(builder, sdpParam.MinValue)
			fb.ParameterAddMaxValue(builder, sdpParam.MaxValue)
			fb.ParameterAddDefaultValue(builder, sdpParam.DefaultValue)
			fb.ParameterAddCurrentValue(builder, sdpParam.CurrentValue)
			fb.ParameterAddRawFlags(builder, sdpParam.RawFlags)
			fb.ParameterAddIsWritable(builder, sdpParam.IsWritable)
			fb.ParameterAddCanRamp(builder, sdpParam.CanRamp)
			paramOffsets = append(paramOffsets, fb.ParameterEnd(builder))
		}

		// Build plugin
		name := builder.CreateString(sdpPlugin.Name)
		manufacturerId := builder.CreateString(sdpPlugin.ManufacturerId)
		componentType := builder.CreateString(sdpPlugin.ComponentType)
		componentSubtype := builder.CreateString(sdpPlugin.ComponentSubtype)

		fb.PluginStartParametersVector(builder, len(paramOffsets))
		for i := len(paramOffsets) - 1; i >= 0; i-- {
			builder.PrependUOffsetT(paramOffsets[i])
		}
		parametersVector := builder.EndVector(len(paramOffsets))

		fb.PluginStart(builder)
		fb.PluginAddName(builder, name)
		fb.PluginAddManufacturerId(builder, manufacturerId)
		fb.PluginAddComponentType(builder, componentType)
		fb.PluginAddComponentSubtype(builder, componentSubtype)
		fb.PluginAddParameters(builder, parametersVector)
		pluginOffsets = append(pluginOffsets, fb.PluginEnd(builder))
	}

	// Build PluginRegistry
	fb.PluginRegistryStartPluginsVector(builder, len(pluginOffsets))
	for i := len(pluginOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(pluginOffsets[i])
	}
	pluginsVector := builder.EndVector(len(pluginOffsets))

	fb.PluginRegistryStart(builder)
	fb.PluginRegistryAddPlugins(builder, pluginsVector)
	fb.PluginRegistryAddTotalPluginCount(builder, sdp.TotalPluginCount)
	fb.PluginRegistryAddTotalParameterCount(builder, sdp.TotalParameterCount)
	registry := fb.PluginRegistryEnd(builder)

	builder.Finish(registry)
	return builder.FinishedBytes()
}
