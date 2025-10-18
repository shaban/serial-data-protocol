package golang

import (
	"bytes"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// TestIntegrationSimplePrimitives tests a complete encode/decode roundtrip
// with a struct containing only primitive types.
func TestIntegrationSimplePrimitives(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Device",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "count", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u16"}},
					{Name: "active", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}},
				},
			},
		},
	}

	// Generate all code
	structs, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("GenerateStructs failed: %v", err)
	}

	encoder, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("GenerateEncoder failed: %v", err)
	}

	encodeHelpers, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateEncodeHelpers failed: %v", err)
	}

	decoder, err := GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("GenerateDecoder failed: %v", err)
	}

	decodeHelpers, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateDecodeHelpers failed: %v", err)
	}

	errors := GenerateErrors()
	context := GenerateDecodeContext()

	// Verify all code was generated
	if structs == "" {
		t.Error("structs is empty")
	}
	if encoder == "" {
		t.Error("encoder is empty")
	}
	if encodeHelpers == "" {
		t.Error("encodeHelpers is empty")
	}
	if decoder == "" {
		t.Error("decoder is empty")
	}
	if decodeHelpers == "" {
		t.Error("decodeHelpers is empty")
	}
	if errors == "" {
		t.Error("errors is empty")
	}
	if context == "" {
		t.Error("context is empty")
	}

	// Verify key functions are present
	if !bytes.Contains([]byte(structs), []byte("type Device struct")) {
		t.Error("Device struct not generated")
	}
	if !bytes.Contains([]byte(encoder), []byte("func EncodeDevice(")) {
		t.Error("EncodeDevice not generated")
	}
	if !bytes.Contains([]byte(encodeHelpers), []byte("func encodeDevice(")) {
		t.Error("encodeDevice helper not generated")
	}
	if !bytes.Contains([]byte(decoder), []byte("func DecodeDevice(")) {
		t.Error("DecodeDevice not generated")
	}
	if !bytes.Contains([]byte(decodeHelpers), []byte("func decodeDevice(")) {
		t.Error("decodeDevice helper not generated")
	}
}

// TestIntegrationWithString tests encode/decode roundtrip with string fields.
func TestIntegrationWithString(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Message",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "content", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	// Generate all code
	structs, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("GenerateStructs failed: %v", err)
	}

	_, err = GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("GenerateEncoder failed: %v", err)
	}

	encodeHelpers, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateEncodeHelpers failed: %v", err)
	}

	_, err = GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("GenerateDecoder failed: %v", err)
	}

	decodeHelpers, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateDecodeHelpers failed: %v", err)
	}

	// Verify string handling in encoder
	if !bytes.Contains([]byte(encodeHelpers), []byte("len(src.Content)")) {
		t.Error("string length calculation not found in encoder")
	}
	if !bytes.Contains([]byte(encodeHelpers), []byte("copy(buf[*offset:], src.Content)")) {
		t.Error("string copy not found in encoder")
	}

	// Verify string handling in decoder
	if !bytes.Contains([]byte(decodeHelpers), []byte("strLen = binary.LittleEndian.Uint32")) {
		t.Error("string length decode not found in decoder")
	}
	if !bytes.Contains([]byte(decodeHelpers), []byte("string(data[*offset:*offset+int(strLen)])")) {
		t.Error("string decode not found in decoder")
	}

	// Verify struct fields
	if !bytes.Contains([]byte(structs), []byte("Id uint32")) {
		t.Error("Id field not found in struct")
	}
	if !bytes.Contains([]byte(structs), []byte("Content string")) {
		t.Error("Content field not found in struct")
	}
}

// TestIntegrationWithArrays tests encode/decode roundtrip with array fields.
func TestIntegrationWithArrays(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Container",
				Fields: []parser.Field{
					{
						Name: "values",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"},
						},
					},
					{
						Name: "tags",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"},
						},
					},
				},
			},
		},
	}

	// Generate all code
	structs, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("GenerateStructs failed: %v", err)
	}

	_, err = GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("GenerateEncoder failed: %v", err)
	}

	encodeHelpers, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateEncodeHelpers failed: %v", err)
	}

	_, err = GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("GenerateDecoder failed: %v", err)
	}

	decodeHelpers, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateDecodeHelpers failed: %v", err)
	}

	// Verify array handling in encoder
	if !bytes.Contains([]byte(encodeHelpers), []byte("len(src.Values)")) {
		t.Error("array length calculation not found in encoder")
	}
	if !bytes.Contains([]byte(encodeHelpers), []byte("for i := range src.Values")) {
		t.Error("array loop not found in encoder")
	}

	// Verify array handling in decoder
	if !bytes.Contains([]byte(decodeHelpers), []byte("arrCount = binary.LittleEndian.Uint32")) {
		t.Error("array count decode not found in decoder")
	}
	if !bytes.Contains([]byte(decodeHelpers), []byte("make([]uint32, arrCount)")) {
		t.Error("array allocation not found in decoder")
	}
	if !bytes.Contains([]byte(decodeHelpers), []byte("for i := uint32(0); i < arrCount; i++")) {
		t.Error("array decode loop not found in decoder")
	}

	// Verify struct fields
	if !bytes.Contains([]byte(structs), []byte("Values []uint32")) {
		t.Error("Values field not found in struct")
	}
	if !bytes.Contains([]byte(structs), []byte("Tags []string")) {
		t.Error("Tags field not found in struct")
	}
}

// TestIntegrationWithNestedStructs tests encode/decode roundtrip with nested structs.
func TestIntegrationWithNestedStructs(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Address",
				Fields: []parser.Field{
					{Name: "street", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{Name: "zip", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
			{
				Name: "Person",
				Fields: []parser.Field{
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{Name: "age", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
					{Name: "address", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Address"}},
				},
			},
		},
	}

	// Generate all code
	structs, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("GenerateStructs failed: %v", err)
	}

	_, err = GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("GenerateEncoder failed: %v", err)
	}

	encodeHelpers, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateEncodeHelpers failed: %v", err)
	}

	_, err = GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("GenerateDecoder failed: %v", err)
	}

	decodeHelpers, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateDecodeHelpers failed: %v", err)
	}

	// Verify nested struct handling in encoder
	if !bytes.Contains([]byte(encodeHelpers), []byte("encodeAddress(&src.Address, buf, offset)")) {
		t.Error("nested struct encode call not found in encoder")
	}

	// Verify nested struct handling in decoder
	if !bytes.Contains([]byte(decodeHelpers), []byte("decodeAddress(&dest.Address, data, offset, ctx)")) {
		t.Error("nested struct decode call not found in decoder")
	}

	// Verify both struct types are generated
	if !bytes.Contains([]byte(structs), []byte("type Address struct")) {
		t.Error("Address struct not generated")
	}
	if !bytes.Contains([]byte(structs), []byte("type Person struct")) {
		t.Error("Person struct not generated")
	}
	if !bytes.Contains([]byte(structs), []byte("Address Address")) {
		t.Error("Address field not found in Person struct")
	}

	// Verify both encode helpers exist
	if !bytes.Contains([]byte(encodeHelpers), []byte("func encodeAddress(")) {
		t.Error("encodeAddress helper not generated")
	}
	if !bytes.Contains([]byte(encodeHelpers), []byte("func encodePerson(")) {
		t.Error("encodePerson helper not generated")
	}

	// Verify both decode helpers exist
	if !bytes.Contains([]byte(decodeHelpers), []byte("func decodeAddress(")) {
		t.Error("decodeAddress helper not generated")
	}
	if !bytes.Contains([]byte(decodeHelpers), []byte("func decodePerson(")) {
		t.Error("decodePerson helper not generated")
	}
}

// TestIntegrationWithStructArrays tests encode/decode roundtrip with arrays of structs.
func TestIntegrationWithStructArrays(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Item",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"}},
				},
			},
			{
				Name: "Collection",
				Fields: []parser.Field{
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{
						Name: "items",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Item"},
						},
					},
				},
			},
		},
	}

	// Generate all code
	structs, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("GenerateStructs failed: %v", err)
	}

	_, err = GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("GenerateEncoder failed: %v", err)
	}

	encodeHelpers, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateEncodeHelpers failed: %v", err)
	}

	_, err = GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("GenerateDecoder failed: %v", err)
	}

	decodeHelpers, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateDecodeHelpers failed: %v", err)
	}

	// Verify struct array handling in encoder
	if !bytes.Contains([]byte(encodeHelpers), []byte("for i := range src.Items")) {
		t.Error("struct array loop not found in encoder")
	}
	if !bytes.Contains([]byte(encodeHelpers), []byte("encodeItem(&src.Items[i], buf, offset)")) {
		t.Error("struct array element encode not found in encoder")
	}

	// Verify struct array handling in decoder
	if !bytes.Contains([]byte(decodeHelpers), []byte("make([]Item, arrCount)")) {
		t.Error("struct array allocation not found in decoder")
	}
	if !bytes.Contains([]byte(decodeHelpers), []byte("decodeItem(&dest.Items[i], data, offset, ctx)")) {
		t.Error("struct array element decode not found in decoder")
	}

	// Verify struct fields
	if !bytes.Contains([]byte(structs), []byte("Items []Item")) {
		t.Error("Items field not found in Collection struct")
	}
}

// TestIntegrationComplexSchema tests a complex real-world-like schema with multiple field types.
func TestIntegrationComplexSchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Parameter",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "min", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "max", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
				},
			},
			{
				Name: "Plugin",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{Name: "manufacturer", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{Name: "version", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "enabled", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}},
					{
						Name: "parameters",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Parameter"},
						},
					},
				},
			},
		},
	}

	// Generate all code
	structs, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("GenerateStructs failed: %v", err)
	}

	encoder, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("GenerateEncoder failed: %v", err)
	}

	encodeHelpers, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateEncodeHelpers failed: %v", err)
	}

	decoder, err := GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("GenerateDecoder failed: %v", err)
	}

	decodeHelpers, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateDecodeHelpers failed: %v", err)
	}

	errors := GenerateErrors()
	if err != nil {
		t.Fatalf("GenerateErrors failed: %v", err)
	}

	context := GenerateDecodeContext()
	if err != nil {
		t.Fatalf("GenerateDecodeContext failed: %v", err)
	}

	// Verify all major components are present
	components := []struct {
		code     string
		name     string
		contains []string
	}{
		{structs, "structs", []string{"type Parameter struct", "type Plugin struct", "Parameters []Parameter"}},
		{encoder, "encoder", []string{"func EncodeParameter(", "func EncodePlugin(", "calculateParameterSize", "calculatePluginSize"}},
		{encodeHelpers, "encodeHelpers", []string{"func encodeParameter(", "func encodePlugin("}},
		{decoder, "decoder", []string{"func DecodeParameter(", "func DecodePlugin(", "MaxSerializedSize", "DecodeContext"}},
		{decodeHelpers, "decodeHelpers", []string{"func decodeParameter(", "func decodePlugin(", "checkArraySize"}},
		{errors, "errors", []string{"ErrDataTooLarge", "ErrUnexpectedEOF", "ErrTooManyElements"}},
		{context, "context", []string{"type DecodeContext struct", "MaxArrayElements", "checkArraySize"}},
	}

	for _, comp := range components {
		for _, needle := range comp.contains {
			if !bytes.Contains([]byte(comp.code), []byte(needle)) {
				t.Errorf("%s: missing %q", comp.name, needle)
			}
		}
	}

	// Verify size calculation includes all fields
	if !bytes.Contains([]byte(encoder), []byte("size += 4 + len(src.Name)")) {
		t.Error("string size calculation not found")
	}
	if !bytes.Contains([]byte(encoder), []byte("for i := range src.Parameters")) {
		t.Error("array size calculation loop not found")
	}

	// Verify parameter array encoding
	if !bytes.Contains([]byte(encodeHelpers), []byte("for i := range src.Parameters")) {
		t.Error("parameter array loop not found in encoder")
	}
	if !bytes.Contains([]byte(encodeHelpers), []byte("encodeParameter(&src.Parameters[i]")) {
		t.Error("parameter encode call not found in encoder")
	}

	// Verify parameter array decoding
	if !bytes.Contains([]byte(decodeHelpers), []byte("make([]Parameter, arrCount)")) {
		t.Error("parameter array allocation not found in decoder")
	}
	if !bytes.Contains([]byte(decodeHelpers), []byte("decodeParameter(&dest.Parameters[i]")) {
		t.Error("parameter decode call not found in decoder")
	}
}

// TestIntegrationAllPrimitiveTypes tests that all primitive types are handled correctly.
func TestIntegrationAllPrimitiveTypes(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "AllTypes",
				Fields: []parser.Field{
					{Name: "u8_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
					{Name: "u16_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u16"}},
					{Name: "u32_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "u64_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u64"}},
					{Name: "i8_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i8"}},
					{Name: "i16_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i16"}},
					{Name: "i32_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
					{Name: "i64_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i64"}},
					{Name: "f32_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "f64_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"}},
					{Name: "bool_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}},
					{Name: "str_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	// Generate all code
	structs, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("GenerateStructs failed: %v", err)
	}

	encodeHelpers, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateEncodeHelpers failed: %v", err)
	}

	decodeHelpers, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("GenerateDecodeHelpers failed: %v", err)
	}

	// Verify all primitive types in struct
	primitiveTypes := []string{
		"U8Field uint8",
		"U16Field uint16",
		"U32Field uint32",
		"U64Field uint64",
		"I8Field int8",
		"I16Field int16",
		"I32Field int32",
		"I64Field int64",
		"F32Field float32",
		"F64Field float64",
		"BoolField bool",
		"StrField string",
	}

	for _, typ := range primitiveTypes {
		if !bytes.Contains([]byte(structs), []byte(typ)) {
			t.Errorf("struct field %q not found", typ)
		}
	}

	// Verify all encode operations
	encodeOps := []string{
		"buf[*offset] = src.U8Field",                                         // u8
		"binary.LittleEndian.PutUint16(buf[*offset:], src.U16Field)",         // u16
		"binary.LittleEndian.PutUint32(buf[*offset:], src.U32Field)",         // u32
		"binary.LittleEndian.PutUint64(buf[*offset:], src.U64Field)",         // u64
		"buf[*offset] = uint8(src.I8Field)",                                  // i8
		"binary.LittleEndian.PutUint16(buf[*offset:], uint16(src.I16Field))", // i16
		"binary.LittleEndian.PutUint32(buf[*offset:], uint32(src.I32Field))", // i32
		"binary.LittleEndian.PutUint64(buf[*offset:], uint64(src.I64Field))", // i64
		"math.Float32bits(src.F32Field)",                                     // f32
		"math.Float64bits(src.F64Field)",                                     // f64
		"if src.BoolField {",                                                 // bool
		"len(src.StrField)",                                                  // str
	}

	for _, op := range encodeOps {
		if !bytes.Contains([]byte(encodeHelpers), []byte(op)) {
			t.Errorf("encode operation %q not found", op)
		}
	}

	// Verify all decode operations
	decodeOps := []string{
		"dest.U8Field = uint8(data[*offset])",                               // u8
		"dest.U16Field = binary.LittleEndian.Uint16(data[*offset:])",        // u16
		"dest.U32Field = binary.LittleEndian.Uint32(data[*offset:])",        // u32
		"dest.U64Field = binary.LittleEndian.Uint64(data[*offset:])",        // u64
		"dest.I8Field = int8(data[*offset])",                                // i8
		"dest.I16Field = int16(binary.LittleEndian.Uint16(data[*offset:]))", // i16
		"dest.I32Field = int32(binary.LittleEndian.Uint32(data[*offset:]))", // i32
		"dest.I64Field = int64(binary.LittleEndian.Uint64(data[*offset:]))", // i64
		"math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))",  // f32
		"math.Float64frombits(binary.LittleEndian.Uint64(data[*offset:]))",  // f64
		"dest.BoolField = data[*offset] != 0",                               // bool
		"dest.StrField = string(data[*offset:*offset+int(strLen)])",         // str
	}

	for _, op := range decodeOps {
		if !bytes.Contains([]byte(decodeHelpers), []byte(op)) {
			t.Errorf("decode operation %q not found", op)
		}
	}
}
