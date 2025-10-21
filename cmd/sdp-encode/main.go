package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	arrays "github.com/shaban/serial-data-protocol/testdata/go/arrays"
	audiounit "github.com/shaban/serial-data-protocol/testdata/go/audiounit"
	nested "github.com/shaban/serial-data-protocol/testdata/go/nested"
	optional "github.com/shaban/serial-data-protocol/testdata/go/optional"
	primitives "github.com/shaban/serial-data-protocol/testdata/go/primitives"
)

func main() {
	schema := flag.String("schema", "", "Schema name: primitives, arrays, nested, optional, audiounit")
	jsonFile := flag.String("json", "", "Path to input .json file")
	outFile := flag.String("out", "", "Path to output .sdpb file")
	typeName := flag.String("type", "", "Type name to encode")
	flag.Parse()

	if *schema == "" || *jsonFile == "" || *outFile == "" || *typeName == "" {
		fmt.Fprintf(os.Stderr, "Usage: sdp-encode -schema <primitives|arrays|nested|optional|audiounit> -json <file.json> -out <file.sdpb> -type <TypeName>\n")
		os.Exit(1)
	}

	// Read JSON
	jsonData, err := os.ReadFile(*jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read JSON: %v\n", err)
		os.Exit(1)
	}

	// Encode based on schema and type
	var encoded []byte
	switch *schema {
	case "primitives":
		if *typeName != "AllPrimitives" {
			fmt.Fprintf(os.Stderr, "Type '%s' not found in schema 'primitives'\n", *typeName)
			os.Exit(1)
		}
		var data []primitives.AllPrimitives
		if err := json.Unmarshal(jsonData, &data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse JSON: %v\n", err)
			os.Exit(1)
		}
		// Encode all instances concatenated
		for _, item := range data {
			buf, err := primitives.EncodeAllPrimitives(&item)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to encode: %v\n", err)
				os.Exit(1)
			}
			encoded = append(encoded, buf...)
		}

	case "arrays":
		// arrays.json has "primitives" and "structs" keys
		var wrapper struct {
			Primitives []arrays.ArraysOfPrimitives `json:"primitives"`
			Structs    []arrays.ArraysOfStructs    `json:"structs"`
		}
		if err := json.Unmarshal(jsonData, &wrapper); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse JSON: %v\n", err)
			os.Exit(1)
		}

		switch *typeName {
		case "ArraysOfPrimitives":
			for _, item := range wrapper.Primitives {
				buf, err := arrays.EncodeArraysOfPrimitives(&item)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to encode: %v\n", err)
					os.Exit(1)
				}
				encoded = append(encoded, buf...)
			}
		case "ArraysOfStructs":
			for _, item := range wrapper.Structs {
				buf, err := arrays.EncodeArraysOfStructs(&item)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to encode: %v\n", err)
					os.Exit(1)
				}
				encoded = append(encoded, buf...)
			}
		default:
			fmt.Fprintf(os.Stderr, "Type '%s' not found in schema 'arrays'\n", *typeName)
			os.Exit(1)
		}

	case "nested":
		if *typeName != "Scene" {
			fmt.Fprintf(os.Stderr, "Type '%s' not found in schema 'nested'\n", *typeName)
			os.Exit(1)
		}
		var data []nested.Scene
		if err := json.Unmarshal(jsonData, &data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse JSON: %v\n", err)
			os.Exit(1)
		}
		for _, item := range data {
			buf, err := nested.EncodeScene(&item)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to encode: %v\n", err)
				os.Exit(1)
			}
			encoded = append(encoded, buf...)
		}

	case "optional":
		if *typeName != "Config" {
			fmt.Fprintf(os.Stderr, "Type '%s' not found in schema 'optional'\n", *typeName)
			os.Exit(1)
		}
		var data []optional.Config
		if err := json.Unmarshal(jsonData, &data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse JSON: %v\n", err)
			os.Exit(1)
		}
		for _, item := range data {
			buf, err := optional.EncodeConfig(&item)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to encode: %v\n", err)
				os.Exit(1)
			}
			encoded = append(encoded, buf...)
		}

	case "audiounit":
		if *typeName != "PluginRegistry" {
			fmt.Fprintf(os.Stderr, "Type '%s' not found in schema 'audiounit'\n", *typeName)
			os.Exit(1)
		}
		// plugins.json is array of plugins, need to wrap in PluginRegistry
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
			fmt.Fprintf(os.Stderr, "Failed to parse JSON: %v\n", err)
			os.Exit(1)
		}

		// Convert to PluginRegistry
		registry := audiounit.PluginRegistry{
			Plugins:             make([]audiounit.Plugin, 0, len(pluginsJSON)),
			TotalPluginCount:    uint32(len(pluginsJSON)),
			TotalParameterCount: 0,
		}

		for _, pJSON := range pluginsJSON {
			plugin := audiounit.Plugin{
				Name:             pJSON.Name,
				ManufacturerId:   pJSON.ManufacturerID,
				ComponentType:    pJSON.Type,
				ComponentSubtype: pJSON.Subtype,
				Parameters:       make([]audiounit.Parameter, 0, len(pJSON.Parameters)),
			}
			for _, paramJSON := range pJSON.Parameters {
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
				registry.TotalParameterCount++
			}
			registry.Plugins = append(registry.Plugins, plugin)
		}

		// Encode
		buf, err := audiounit.EncodePluginRegistry(&registry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to encode: %v\n", err)
			os.Exit(1)
		}
		encoded = buf

	default:
		fmt.Fprintf(os.Stderr, "Unknown schema: %s\n", *schema)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(*outFile, encoded, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully encoded %d bytes to %s\n", len(encoded), *outFile)
}
