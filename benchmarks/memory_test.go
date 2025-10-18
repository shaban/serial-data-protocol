package benchmarks

import (
	"runtime"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"google.golang.org/protobuf/proto"

	"github.com/shaban/serial-data-protocol/benchmarks/fb"
	pb "github.com/shaban/serial-data-protocol/benchmarks/pb"
	audiounit "github.com/shaban/serial-data-protocol/testdata/audiounit/go"
)

// Memory profiling benchmarks to measure actual heap usage and GC pressure

func BenchmarkMemory_SDP_Encode(b *testing.B) {
	b.ReportAllocs()

	// Force GC and get baseline
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoded, err := audiounit.EncodePluginRegistry(&testData)
		if err != nil {
			b.Fatal(err)
		}
		_ = encoded
	}
	b.StopTimer()

	// Measure after benchmark
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/encode")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs/encode")
}

func BenchmarkMemory_SDP_Decode(b *testing.B) {
	encoded, err := audiounit.EncodePluginRegistry(&testData)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	// Force GC and get baseline
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded audiounit.PluginRegistry
		err := audiounit.DecodePluginRegistry(&decoded, encoded)
		if err != nil {
			b.Fatal(err)
		}
		_ = decoded
	}
	b.StopTimer()

	// Measure after benchmark
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/decode")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs/decode")
}

func BenchmarkMemory_Protobuf_Encode(b *testing.B) {
	b.ReportAllocs()

	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoded, err := proto.Marshal(testDataPB)
		if err != nil {
			b.Fatal(err)
		}
		_ = encoded
	}
	b.StopTimer()

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/encode")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs/encode")
}

func BenchmarkMemory_Protobuf_Decode(b *testing.B) {
	encoded, err := proto.Marshal(testDataPB)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoded := &pb.PluginRegistry{}
		err := proto.Unmarshal(encoded, decoded)
		if err != nil {
			b.Fatal(err)
		}
		_ = decoded
	}
	b.StopTimer()

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/decode")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs/decode")
}

func BenchmarkMemory_FlatBuffers_Encode(b *testing.B) {
	b.ReportAllocs()

	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
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
		_ = builder.FinishedBytes()
	}
	b.StopTimer()

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/encode")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs/encode")
}

func BenchmarkMemory_FlatBuffers_Decode(b *testing.B) {
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

	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoded := fb.GetRootAsPluginRegistry(encoded, 0)
		if decoded.TotalPluginCount() != testData.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
		_ = decoded
	}
	b.StopTimer()

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/decode")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs/decode")
}

// Peak heap usage tests
func TestPeakHeapUsage_SDP(t *testing.T) {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Encode
	encoded, err := audiounit.EncodePluginRegistry(&testData)
	if err != nil {
		t.Fatal(err)
	}

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	encodePeak := m2.HeapAlloc - m1.HeapAlloc

	// Decode
	var decoded audiounit.PluginRegistry
	err = audiounit.DecodePluginRegistry(&decoded, encoded)
	if err != nil {
		t.Fatal(err)
	}

	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)
	decodePeak := m3.HeapAlloc - m2.HeapAlloc

	t.Logf("SDP Peak Heap Usage:")
	t.Logf("  Encode: %d bytes (%.2f KB)", encodePeak, float64(encodePeak)/1024)
	t.Logf("  Decode: %d bytes (%.2f KB)", decodePeak, float64(decodePeak)/1024)
	t.Logf("  Total:  %d bytes (%.2f KB)", encodePeak+decodePeak, float64(encodePeak+decodePeak)/1024)
}

func TestPeakHeapUsage_Protobuf(t *testing.T) {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Encode
	encoded, err := proto.Marshal(testDataPB)
	if err != nil {
		t.Fatal(err)
	}

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	encodePeak := m2.HeapAlloc - m1.HeapAlloc

	// Decode
	decoded := &pb.PluginRegistry{}
	err = proto.Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatal(err)
	}

	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)
	decodePeak := m3.HeapAlloc - m2.HeapAlloc

	t.Logf("Protocol Buffers Peak Heap Usage:")
	t.Logf("  Encode: %d bytes (%.2f KB)", encodePeak, float64(encodePeak)/1024)
	t.Logf("  Decode: %d bytes (%.2f KB)", decodePeak, float64(decodePeak)/1024)
	t.Logf("  Total:  %d bytes (%.2f KB)", encodePeak+decodePeak, float64(encodePeak+decodePeak)/1024)
}

func TestPeakHeapUsage_FlatBuffers(t *testing.T) {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

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

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	encodePeak := m2.HeapAlloc - m1.HeapAlloc

	// Decode (zero-copy, should be minimal)
	decoded := fb.GetRootAsPluginRegistry(encoded, 0)
	_ = decoded

	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)
	decodePeak := m3.HeapAlloc - m2.HeapAlloc

	t.Logf("FlatBuffers Peak Heap Usage:")
	t.Logf("  Encode: %d bytes (%.2f KB)", encodePeak, float64(encodePeak)/1024)
	t.Logf("  Decode: %d bytes (%.2f KB)", decodePeak, float64(decodePeak)/1024)
	t.Logf("  Total:  %d bytes (%.2f KB)", encodePeak+decodePeak, float64(encodePeak+decodePeak)/1024)
}
