package benchmarks

import (
	"encoding/json"
	"os"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"google.golang.org/protobuf/proto"

	"github.com/shaban/serial-data-protocol/benchmarks/fb"
	pb "github.com/shaban/serial-data-protocol/benchmarks/pb"
	"github.com/shaban/serial-data-protocol/testdata/audiounit"
)

// loadTestData loads the real-world AudioUnit plugin data from JSON
// This is done once in init() and reused across all benchmarks for fairness
var testData audiounit.PluginRegistry
var testDataPB *pb.PluginRegistry

func init() {
	jsonData, err := os.ReadFile("../testdata/plugins.json")
	if err != nil {
		panic("plugins.json not found: " + err.Error())
	}

	// Parse JSON into intermediate structure
	var pluginsJSON []struct {
		Name           string `json:"name"`
		ManufacturerID string `json:"manufacturerID"`
		Type           string `json:"type"`
		Subtype        string `json:"subtype"`
		Parameters     []struct {
			Address      uint64  `json:"address"`
			DisplayName  string  `json:"displayName"`
			Identifier   string  `json:"identifier"`
			Unit         string  `json:"unit"`
			MinValue     float32 `json:"minValue"`
			MaxValue     float32 `json:"maxValue"`
			DefaultValue float32 `json:"defaultValue"`
			CurrentValue float32 `json:"currentValue"`
			RawFlags     uint32  `json:"rawFlags"`
			IsWritable   bool    `json:"isWritable"`
			CanRamp      bool    `json:"canRamp"`
		} `json:"parameters"`
	}

	if err := json.Unmarshal(jsonData, &pluginsJSON); err != nil {
		panic("Failed to parse plugins.json: " + err.Error())
	}

	// Convert to SDP structs
	testData = audiounit.PluginRegistry{
		Plugins:             make([]audiounit.Plugin, 0, len(pluginsJSON)),
		TotalPluginCount:    uint32(len(pluginsJSON)),
		TotalParameterCount: 0,
	}

	// Also prepare Protocol Buffers version
	testDataPB = &pb.PluginRegistry{
		Plugins:             make([]*pb.Plugin, 0, len(pluginsJSON)),
		TotalPluginCount:    uint32(len(pluginsJSON)),
		TotalParameterCount: 0,
	}

	totalParams := uint32(0)
	for _, pJSON := range pluginsJSON {
		// SDP version
		plugin := audiounit.Plugin{
			Name:             pJSON.Name,
			ManufacturerId:   pJSON.ManufacturerID,
			ComponentType:    pJSON.Type,
			ComponentSubtype: pJSON.Subtype,
			Parameters:       make([]audiounit.Parameter, 0, len(pJSON.Parameters)),
		}

		// Protocol Buffers version
		pluginPB := &pb.Plugin{
			Name:             pJSON.Name,
			ManufacturerId:   pJSON.ManufacturerID,
			ComponentType:    pJSON.Type,
			ComponentSubtype: pJSON.Subtype,
			Parameters:       make([]*pb.Parameter, 0, len(pJSON.Parameters)),
		}

		for _, paramJSON := range pJSON.Parameters {
			// SDP parameter
			param := audiounit.Parameter{
				Address:      paramJSON.Address,
				DisplayName:  paramJSON.DisplayName,
				Identifier:   paramJSON.Identifier,
				Unit:         paramJSON.Unit,
				MinValue:     paramJSON.MinValue,
				MaxValue:     paramJSON.MaxValue,
				DefaultValue: paramJSON.DefaultValue,
				CurrentValue: paramJSON.CurrentValue,
				RawFlags:     paramJSON.RawFlags,
				IsWritable:   paramJSON.IsWritable,
				CanRamp:      paramJSON.CanRamp,
			}
			plugin.Parameters = append(plugin.Parameters, param)

			// Protocol Buffers parameter
			paramPB := &pb.Parameter{
				Address:      paramJSON.Address,
				DisplayName:  paramJSON.DisplayName,
				Identifier:   paramJSON.Identifier,
				Unit:         paramJSON.Unit,
				MinValue:     paramJSON.MinValue,
				MaxValue:     paramJSON.MaxValue,
				DefaultValue: paramJSON.DefaultValue,
				CurrentValue: paramJSON.CurrentValue,
				RawFlags:     paramJSON.RawFlags,
				IsWritable:   paramJSON.IsWritable,
				CanRamp:      paramJSON.CanRamp,
			}
			pluginPB.Parameters = append(pluginPB.Parameters, paramPB)

			totalParams++
		}

		testData.Plugins = append(testData.Plugins, plugin)
		testDataPB.Plugins = append(testDataPB.Plugins, pluginPB)
	}

	testData.TotalParameterCount = totalParams
	testDataPB.TotalParameterCount = totalParams
}

// ============================================================================
// SDP Benchmarks
// ============================================================================

func BenchmarkSDP_Encode(b *testing.B) {
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

func BenchmarkSDP_Decode(b *testing.B) {
	encoded, err := audiounit.EncodePluginRegistry(&testData)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded audiounit.PluginRegistry
		err := audiounit.DecodePluginRegistry(&decoded, encoded)
		if err != nil {
			b.Fatal(err)
		}
		if decoded.TotalPluginCount != testData.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}

func BenchmarkSDP_Roundtrip(b *testing.B) {
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
	encoded, err := proto.Marshal(testDataPB)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoded := &pb.PluginRegistry{}
		err := proto.Unmarshal(encoded, decoded)
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
		builder := flatbuffers.NewBuilder(0)

		// Build parameters for each plugin
		pluginOffsets := make([]flatbuffers.UOffsetT, 0, len(testData.Plugins))

		for _, plugin := range testData.Plugins {
			// Build parameters array
			paramOffsets := make([]flatbuffers.UOffsetT, 0, len(plugin.Parameters))
			for _, param := range plugin.Parameters {
				displayName := builder.CreateString(param.DisplayName)
				identifier := builder.CreateString(param.Identifier)
				unit := builder.CreateString(param.Unit)

				fb.ParameterStart(builder)
				fb.ParameterAddAddress(builder, param.Address)
				fb.ParameterAddDisplayName(builder, displayName)
				fb.ParameterAddIdentifier(builder, identifier)
				fb.ParameterAddUnit(builder, unit)
				fb.ParameterAddMinValue(builder, param.MinValue)
				fb.ParameterAddMaxValue(builder, param.MaxValue)
				fb.ParameterAddDefaultValue(builder, param.DefaultValue)
				fb.ParameterAddCurrentValue(builder, param.CurrentValue)
				fb.ParameterAddRawFlags(builder, param.RawFlags)
				fb.ParameterAddIsWritable(builder, param.IsWritable)
				fb.ParameterAddCanRamp(builder, param.CanRamp)
				paramOffsets = append(paramOffsets, fb.ParameterEnd(builder))
			}

			// Build plugin
			name := builder.CreateString(plugin.Name)
			manufacturerId := builder.CreateString(plugin.ManufacturerId)
			componentType := builder.CreateString(plugin.ComponentType)
			componentSubtype := builder.CreateString(plugin.ComponentSubtype)

			fb.PluginStartParametersVector(builder, len(paramOffsets))
			for j := len(paramOffsets) - 1; j >= 0; j-- {
				builder.PrependUOffsetT(paramOffsets[j])
			}
			parametersVec := builder.EndVector(len(paramOffsets))

			fb.PluginStart(builder)
			fb.PluginAddName(builder, name)
			fb.PluginAddManufacturerId(builder, manufacturerId)
			fb.PluginAddComponentType(builder, componentType)
			fb.PluginAddComponentSubtype(builder, componentSubtype)
			fb.PluginAddParameters(builder, parametersVec)
			pluginOffsets = append(pluginOffsets, fb.PluginEnd(builder))
		}

		// Build plugin registry
		fb.PluginRegistryStartPluginsVector(builder, len(pluginOffsets))
		for j := len(pluginOffsets) - 1; j >= 0; j-- {
			builder.PrependUOffsetT(pluginOffsets[j])
		}
		pluginsVec := builder.EndVector(len(pluginOffsets))

		fb.PluginRegistryStart(builder)
		fb.PluginRegistryAddPlugins(builder, pluginsVec)
		fb.PluginRegistryAddTotalPluginCount(builder, testData.TotalPluginCount)
		fb.PluginRegistryAddTotalParameterCount(builder, testData.TotalParameterCount)
		registry := fb.PluginRegistryEnd(builder)

		builder.Finish(registry)
		encoded := builder.FinishedBytes()

		if len(encoded) == 0 {
			b.Fatal("empty encoding")
		}
	}
}

func BenchmarkFlatBuffers_Decode(b *testing.B) {
	// Encode once
	builder := flatbuffers.NewBuilder(0)

	pluginOffsets := make([]flatbuffers.UOffsetT, 0, len(testData.Plugins))
	for _, plugin := range testData.Plugins {
		paramOffsets := make([]flatbuffers.UOffsetT, 0, len(plugin.Parameters))
		for _, param := range plugin.Parameters {
			displayName := builder.CreateString(param.DisplayName)
			identifier := builder.CreateString(param.Identifier)
			unit := builder.CreateString(param.Unit)

			fb.ParameterStart(builder)
			fb.ParameterAddAddress(builder, param.Address)
			fb.ParameterAddDisplayName(builder, displayName)
			fb.ParameterAddIdentifier(builder, identifier)
			fb.ParameterAddUnit(builder, unit)
			fb.ParameterAddMinValue(builder, param.MinValue)
			fb.ParameterAddMaxValue(builder, param.MaxValue)
			fb.ParameterAddDefaultValue(builder, param.DefaultValue)
			fb.ParameterAddCurrentValue(builder, param.CurrentValue)
			fb.ParameterAddRawFlags(builder, param.RawFlags)
			fb.ParameterAddIsWritable(builder, param.IsWritable)
			fb.ParameterAddCanRamp(builder, param.CanRamp)
			paramOffsets = append(paramOffsets, fb.ParameterEnd(builder))
		}

		name := builder.CreateString(plugin.Name)
		manufacturerId := builder.CreateString(plugin.ManufacturerId)
		componentType := builder.CreateString(plugin.ComponentType)
		componentSubtype := builder.CreateString(plugin.ComponentSubtype)

		fb.PluginStartParametersVector(builder, len(paramOffsets))
		for j := len(paramOffsets) - 1; j >= 0; j-- {
			builder.PrependUOffsetT(paramOffsets[j])
		}
		parametersVec := builder.EndVector(len(paramOffsets))

		fb.PluginStart(builder)
		fb.PluginAddName(builder, name)
		fb.PluginAddManufacturerId(builder, manufacturerId)
		fb.PluginAddComponentType(builder, componentType)
		fb.PluginAddComponentSubtype(builder, componentSubtype)
		fb.PluginAddParameters(builder, parametersVec)
		pluginOffsets = append(pluginOffsets, fb.PluginEnd(builder))
	}

	fb.PluginRegistryStartPluginsVector(builder, len(pluginOffsets))
	for j := len(pluginOffsets) - 1; j >= 0; j-- {
		builder.PrependUOffsetT(pluginOffsets[j])
	}
	pluginsVec := builder.EndVector(len(pluginOffsets))

	fb.PluginRegistryStart(builder)
	fb.PluginRegistryAddPlugins(builder, pluginsVec)
	fb.PluginRegistryAddTotalPluginCount(builder, testData.TotalPluginCount)
	fb.PluginRegistryAddTotalParameterCount(builder, testData.TotalParameterCount)
	registry := fb.PluginRegistryEnd(builder)

	builder.Finish(registry)
	encoded := builder.FinishedBytes()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoded := fb.GetRootAsPluginRegistry(encoded, 0)
		if decoded.TotalPluginCount() != testData.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
		// Access some data to prevent optimization
		if decoded.PluginsLength() == 0 {
			b.Fatal("no plugins")
		}
	}
}

func BenchmarkFlatBuffers_Roundtrip(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Encode
		builder := flatbuffers.NewBuilder(0)

		pluginOffsets := make([]flatbuffers.UOffsetT, 0, len(testData.Plugins))
		for _, plugin := range testData.Plugins {
			paramOffsets := make([]flatbuffers.UOffsetT, 0, len(plugin.Parameters))
			for _, param := range plugin.Parameters {
				displayName := builder.CreateString(param.DisplayName)
				identifier := builder.CreateString(param.Identifier)
				unit := builder.CreateString(param.Unit)

				fb.ParameterStart(builder)
				fb.ParameterAddAddress(builder, param.Address)
				fb.ParameterAddDisplayName(builder, displayName)
				fb.ParameterAddIdentifier(builder, identifier)
				fb.ParameterAddUnit(builder, unit)
				fb.ParameterAddMinValue(builder, param.MinValue)
				fb.ParameterAddMaxValue(builder, param.MaxValue)
				fb.ParameterAddDefaultValue(builder, param.DefaultValue)
				fb.ParameterAddCurrentValue(builder, param.CurrentValue)
				fb.ParameterAddRawFlags(builder, param.RawFlags)
				fb.ParameterAddIsWritable(builder, param.IsWritable)
				fb.ParameterAddCanRamp(builder, param.CanRamp)
				paramOffsets = append(paramOffsets, fb.ParameterEnd(builder))
			}

			name := builder.CreateString(plugin.Name)
			manufacturerId := builder.CreateString(plugin.ManufacturerId)
			componentType := builder.CreateString(plugin.ComponentType)
			componentSubtype := builder.CreateString(plugin.ComponentSubtype)

			fb.PluginStartParametersVector(builder, len(paramOffsets))
			for j := len(paramOffsets) - 1; j >= 0; j-- {
				builder.PrependUOffsetT(paramOffsets[j])
			}
			parametersVec := builder.EndVector(len(paramOffsets))

			fb.PluginStart(builder)
			fb.PluginAddName(builder, name)
			fb.PluginAddManufacturerId(builder, manufacturerId)
			fb.PluginAddComponentType(builder, componentType)
			fb.PluginAddComponentSubtype(builder, componentSubtype)
			fb.PluginAddParameters(builder, parametersVec)
			pluginOffsets = append(pluginOffsets, fb.PluginEnd(builder))
		}

		fb.PluginRegistryStartPluginsVector(builder, len(pluginOffsets))
		for j := len(pluginOffsets) - 1; j >= 0; j-- {
			builder.PrependUOffsetT(pluginOffsets[j])
		}
		pluginsVec := builder.EndVector(len(pluginOffsets))

		fb.PluginRegistryStart(builder)
		fb.PluginRegistryAddPlugins(builder, pluginsVec)
		fb.PluginRegistryAddTotalPluginCount(builder, testData.TotalPluginCount)
		fb.PluginRegistryAddTotalParameterCount(builder, testData.TotalParameterCount)
		registry := fb.PluginRegistryEnd(builder)

		builder.Finish(registry)
		encoded := builder.FinishedBytes()

		// Decode
		decoded := fb.GetRootAsPluginRegistry(encoded, 0)

		// Verify
		if decoded.TotalPluginCount() != testData.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}
