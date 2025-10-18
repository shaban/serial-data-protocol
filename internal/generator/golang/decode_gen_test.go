package golang

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// TestGenerateDecoderSimple verifies basic decoder generation
func TestGenerateDecoderSimple(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Device",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	result, err := GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check function signature
	if !strings.Contains(result, "func DecodeDevice(dest *Device, data []byte) error {") {
		t.Errorf("missing correct function signature, got:\n%s", result)
	}

	// Check doc comment
	if !strings.Contains(result, "// DecodeDevice decodes a Device from wire format.") {
		t.Errorf("missing or incorrect doc comment, got:\n%s", result)
	}

	// Check size validation
	if !strings.Contains(result, "if len(data) > MaxSerializedSize {") {
		t.Errorf("missing size validation check, got:\n%s", result)
	}
	if !strings.Contains(result, "return ErrDataTooLarge") {
		t.Errorf("missing ErrDataTooLarge return, got:\n%s", result)
	}

	// Check DecodeContext creation
	if !strings.Contains(result, "ctx := &DecodeContext{}") {
		t.Errorf("missing DecodeContext creation, got:\n%s", result)
	}

	// Check offset initialization
	if !strings.Contains(result, "offset := 0") {
		t.Errorf("missing offset initialization, got:\n%s", result)
	}

	// Check helper function call
	if !strings.Contains(result, "return decodeDevice(dest, data, &offset, ctx)") {
		t.Errorf("missing helper function call, got:\n%s", result)
	}
}

// TestGenerateDecoderMultipleStructs verifies multiple decoder functions
func TestGenerateDecoderMultipleStructs(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Point",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
				},
			},
			{
				Name: "Line",
				Fields: []parser.Field{
					{Name: "start", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Point"}},
				},
			},
		},
	}

	result, err := GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check both functions present
	if !strings.Contains(result, "func DecodePoint(dest *Point, data []byte) error {") {
		t.Errorf("missing DecodePoint function, got:\n%s", result)
	}
	if !strings.Contains(result, "func DecodeLine(dest *Line, data []byte) error {") {
		t.Errorf("missing DecodeLine function, got:\n%s", result)
	}

	// Check both helper calls
	if !strings.Contains(result, "return decodePoint(dest, data, &offset, ctx)") {
		t.Errorf("missing decodePoint call, got:\n%s", result)
	}
	if !strings.Contains(result, "return decodeLine(dest, data, &offset, ctx)") {
		t.Errorf("missing decodeLine call, got:\n%s", result)
	}

	// Verify blank line between functions
	lines := strings.Split(result, "\n")
	foundBlankBetween := false
	for i, line := range lines {
		if strings.Contains(line, "// DecodeLine") {
			// Check if previous line is blank
			if i > 0 && strings.TrimSpace(lines[i-1]) == "" {
				foundBlankBetween = true
			}
		}
	}
	if !foundBlankBetween {
		t.Errorf("expected blank line between decoder functions")
	}
}

// TestGenerateDecoderSnakeCaseConversion verifies name conversion
func TestGenerateDecoderSnakeCaseConversion(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "audio_device",
				Fields: []parser.Field{
					{Name: "device_id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	result, err := GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check function name conversion
	if !strings.Contains(result, "func DecodeAudioDevice(dest *AudioDevice, data []byte) error {") {
		t.Errorf("missing or incorrect function name conversion, got:\n%s", result)
	}

	// Check helper function name conversion
	if !strings.Contains(result, "return decodeAudioDevice(dest, data, &offset, ctx)") {
		t.Errorf("missing or incorrect helper function name, got:\n%s", result)
	}
}

// TestGenerateDecoderValidation verifies all validation checks
func TestGenerateDecoderValidation(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Test",
				Fields: []parser.Field{
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	result, err := GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// List of required validations
	required := []string{
		"if len(data) > MaxSerializedSize {",
		"return ErrDataTooLarge",
		"ctx := &DecodeContext{}",
		"offset := 0",
	}

	for _, req := range required {
		if !strings.Contains(result, req) {
			t.Errorf("missing required validation/setup: %q\ngot:\n%s", req, result)
		}
	}
}

// TestGenerateDecoderDocComments verifies doc comment generation
func TestGenerateDecoderDocComments(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Plugin",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	result, err := GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check doc comment lines
	if !strings.Contains(result, "// DecodePlugin decodes a Plugin from wire format.") {
		t.Errorf("missing first doc comment line, got:\n%s", result)
	}
	if !strings.Contains(result, "// It validates the data size and delegates to the decoder implementation.") {
		t.Errorf("missing second doc comment line, got:\n%s", result)
	}
}

// TestGenerateDecoderComplexSchema verifies realistic schema
func TestGenerateDecoderComplexSchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Parameter",
				Fields: []parser.Field{
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"}},
				},
			},
			{
				Name: "Plugin",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{
						Name: "parameters",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Parameter"},
						},
					},
				},
			},
			{
				Name: "PluginList",
				Fields: []parser.Field{
					{
						Name: "plugins",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Plugin"},
						},
					},
				},
			},
		},
	}

	result, err := GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all three decoder functions
	expectedFuncs := []string{
		"func DecodeParameter(dest *Parameter, data []byte) error {",
		"func DecodePlugin(dest *Plugin, data []byte) error {",
		"func DecodePluginList(dest *PluginList, data []byte) error {",
	}

	for _, expected := range expectedFuncs {
		if !strings.Contains(result, expected) {
			t.Errorf("missing function: %q\ngot:\n%s", expected, result)
		}
	}

	// Verify all helper calls
	expectedHelpers := []string{
		"return decodeParameter(dest, data, &offset, ctx)",
		"return decodePlugin(dest, data, &offset, ctx)",
		"return decodePluginList(dest, data, &offset, ctx)",
	}

	for _, expected := range expectedHelpers {
		if !strings.Contains(result, expected) {
			t.Errorf("missing helper call: %q\ngot:\n%s", expected, result)
		}
	}

	// Count functions (should be 3)
	funcCount := strings.Count(result, "func Decode")
	if funcCount != 3 {
		t.Errorf("expected 3 decoder functions, found %d", funcCount)
	}
}

// TestGenerateDecoderNilSchema verifies error handling
func TestGenerateDecoderNilSchema(t *testing.T) {
	result, err := GenerateDecoder(nil)

	if err == nil {
		t.Errorf("expected error for nil schema, got result: %s", result)
	}
	if !strings.Contains(err.Error(), "schema is nil") {
		t.Errorf("expected 'schema is nil' in error, got: %v", err)
	}
}

// TestGenerateDecoderEmptySchema verifies error handling
func TestGenerateDecoderEmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	result, err := GenerateDecoder(schema)

	if err == nil {
		t.Errorf("expected error for empty schema, got result: %s", result)
	}
	if !strings.Contains(err.Error(), "no structs") {
		t.Errorf("expected 'no structs' in error, got: %v", err)
	}
}

// TestGenerateDecoderFunctionStructure verifies function structure
func TestGenerateDecoderFunctionStructure(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Example",
				Fields: []parser.Field{
					{Name: "data", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
				},
			},
		},
	}

	result, err := GenerateDecoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify function structure in order
	expectedOrder := []string{
		"// DecodeExample",                               // Doc comment first
		"func DecodeExample",                             // Function signature
		"if len(data) > MaxSerializedSize {",             // Size check
		"return ErrDataTooLarge",                         // Error return
		"ctx := &DecodeContext{}",                        // Context creation
		"offset := 0",                                    // Offset init
		"return decodeExample(dest, data, &offset, ctx)", // Helper call
	}

	lastIndex := -1
	for i, expected := range expectedOrder {
		index := strings.Index(result, expected)
		if index == -1 {
			t.Errorf("missing expected element %d: %q", i, expected)
			continue
		}
		if index <= lastIndex {
			t.Errorf("element %d (%q) not in correct order (index %d <= %d)", i, expected, index, lastIndex)
		}
		lastIndex = index
	}
}

// TestGenerateDecodeHelpersSimple verifies basic helper function generation
func TestGenerateDecodeHelpersSimple(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Device",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check function signature
	if !strings.Contains(result, "func decodeDevice(dest *Device, data []byte, offset *int, ctx *DecodeContext) error {") {
		t.Errorf("missing correct function signature, got:\n%s", result)
	}

	// Check doc comment
	if !strings.Contains(result, "// decodeDevice is the helper function that decodes Device fields.") {
		t.Errorf("missing or incorrect doc comment, got:\n%s", result)
	}

	// Check field comment
	if !strings.Contains(result, "// Field: Id (u32)") {
		t.Errorf("missing field comment, got:\n%s", result)
	}

	// Check bounds check
	if !strings.Contains(result, "if *offset + 4 > len(data) {") {
		t.Errorf("missing bounds check, got:\n%s", result)
	}

	// Check EOF error
	if !strings.Contains(result, "return ErrUnexpectedEOF") {
		t.Errorf("missing EOF error return, got:\n%s", result)
	}

	// Check decode logic
	if !strings.Contains(result, "dest.Id = binary.LittleEndian.Uint32(data[*offset:])") {
		t.Errorf("missing decode logic, got:\n%s", result)
	}

	// Check offset increment
	if !strings.Contains(result, "*offset += 4") {
		t.Errorf("missing offset increment, got:\n%s", result)
	}

	// Check return
	if !strings.Contains(result, "return nil") {
		t.Errorf("missing return nil, got:\n%s", result)
	}
}

// TestGenerateDecodeHelpersAllPrimitives verifies all primitive types
func TestGenerateDecodeHelpersAllPrimitives(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "AllTypes",
				Fields: []parser.Field{
					{Name: "u8_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
					{Name: "u16_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u16"}},
					{Name: "u32_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "u64_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u64"}},
					{Name: "i8_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i8"}},
					{Name: "i16_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i16"}},
					{Name: "i32_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
					{Name: "i64_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i64"}},
					{Name: "f32_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "f64_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"}},
					{Name: "bool_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check u8
	if !strings.Contains(result, "if *offset + 1 > len(data) {") {
		t.Errorf("missing u8 bounds check")
	}
	if !strings.Contains(result, "dest.U8Val = uint8(data[*offset])") {
		t.Errorf("missing u8 decode, got:\n%s", result)
	}

	// Check u16
	if !strings.Contains(result, "if *offset + 2 > len(data) {") {
		t.Errorf("missing u16 bounds check")
	}
	if !strings.Contains(result, "dest.U16Val = binary.LittleEndian.Uint16(data[*offset:])") {
		t.Errorf("missing u16 decode, got:\n%s", result)
	}

	// Check u32
	if !strings.Contains(result, "dest.U32Val = binary.LittleEndian.Uint32(data[*offset:])") {
		t.Errorf("missing u32 decode, got:\n%s", result)
	}

	// Check u64
	if !strings.Contains(result, "if *offset + 8 > len(data) {") {
		t.Errorf("missing u64 bounds check")
	}
	if !strings.Contains(result, "dest.U64Val = binary.LittleEndian.Uint64(data[*offset:])") {
		t.Errorf("missing u64 decode, got:\n%s", result)
	}

	// Check i8
	if !strings.Contains(result, "dest.I8Val = int8(data[*offset])") {
		t.Errorf("missing i8 decode, got:\n%s", result)
	}

	// Check i16
	if !strings.Contains(result, "dest.I16Val = int16(binary.LittleEndian.Uint16(data[*offset:]))") {
		t.Errorf("missing i16 decode, got:\n%s", result)
	}

	// Check i32
	if !strings.Contains(result, "dest.I32Val = int32(binary.LittleEndian.Uint32(data[*offset:]))") {
		t.Errorf("missing i32 decode, got:\n%s", result)
	}

	// Check i64
	if !strings.Contains(result, "dest.I64Val = int64(binary.LittleEndian.Uint64(data[*offset:]))") {
		t.Errorf("missing i64 decode, got:\n%s", result)
	}

	// Check f32
	if !strings.Contains(result, "dest.F32Val = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))") {
		t.Errorf("missing f32 decode, got:\n%s", result)
	}

	// Check f64
	if !strings.Contains(result, "dest.F64Val = math.Float64frombits(binary.LittleEndian.Uint64(data[*offset:]))") {
		t.Errorf("missing f64 decode, got:\n%s", result)
	}

	// Check bool
	if !strings.Contains(result, "dest.BoolVal = data[*offset] != 0") {
		t.Errorf("missing bool decode, got:\n%s", result)
	}
}

// TestGenerateDecodeHelpersMultipleFields verifies multiple fields in order
func TestGenerateDecodeHelpersMultipleFields(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Point",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "y", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "z", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all three fields present
	if !strings.Contains(result, "// Field: X (f32)") {
		t.Errorf("missing X field comment")
	}
	if !strings.Contains(result, "// Field: Y (f32)") {
		t.Errorf("missing Y field comment")
	}
	if !strings.Contains(result, "// Field: Z (f32)") {
		t.Errorf("missing Z field comment")
	}

	// Verify order
	xIndex := strings.Index(result, "dest.X =")
	yIndex := strings.Index(result, "dest.Y =")
	zIndex := strings.Index(result, "dest.Z =")

	if xIndex == -1 || yIndex == -1 || zIndex == -1 {
		t.Fatalf("missing field assignments")
	}

	if xIndex >= yIndex || yIndex >= zIndex {
		t.Errorf("fields not in correct order: X@%d, Y@%d, Z@%d", xIndex, yIndex, zIndex)
	}
}

// TestGenerateDecodeHelpersMultipleStructs verifies multiple helper functions
func TestGenerateDecodeHelpersMultipleStructs(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Point",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
				},
			},
			{
				Name: "Color",
				Fields: []parser.Field{
					{Name: "r", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check both functions present
	if !strings.Contains(result, "func decodePoint(dest *Point, data []byte, offset *int, ctx *DecodeContext) error {") {
		t.Errorf("missing decodePoint function")
	}
	if !strings.Contains(result, "func decodeColor(dest *Color, data []byte, offset *int, ctx *DecodeContext) error {") {
		t.Errorf("missing decodeColor function")
	}

	// Verify blank line between functions
	lines := strings.Split(result, "\n")
	foundBlankBetween := false
	for i, line := range lines {
		if strings.Contains(line, "// decodeColor") {
			if i > 0 && strings.TrimSpace(lines[i-1]) == "" {
				foundBlankBetween = true
			}
		}
	}
	if !foundBlankBetween {
		t.Errorf("expected blank line between helper functions")
	}
}

// TestGenerateDecodeHelpersBoundsChecks verifies all bounds checks present
func TestGenerateDecodeHelpersBoundsChecks(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Test",
				Fields: []parser.Field{
					{Name: "a", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
					{Name: "b", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "c", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u64"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count bounds checks (should be 3, one per field)
	boundsCheckCount := strings.Count(result, "if *offset +")
	if boundsCheckCount != 3 {
		t.Errorf("expected 3 bounds checks, found %d", boundsCheckCount)
	}

	// Count EOF errors (should be 3, one per field)
	eofErrorCount := strings.Count(result, "return ErrUnexpectedEOF")
	if eofErrorCount != 3 {
		t.Errorf("expected 3 EOF errors, found %d", eofErrorCount)
	}
}

// TestGenerateDecodeHelpersNilSchema verifies error handling
func TestGenerateDecodeHelpersNilSchema(t *testing.T) {
	result, err := GenerateDecodeHelpers(nil)

	if err == nil {
		t.Errorf("expected error for nil schema, got result: %s", result)
	}
	if !strings.Contains(err.Error(), "schema is nil") {
		t.Errorf("expected 'schema is nil' in error, got: %v", err)
	}
}

// TestGenerateDecodeHelpersEmptySchema verifies error handling
func TestGenerateDecodeHelpersEmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	result, err := GenerateDecodeHelpers(schema)

	if err == nil {
		t.Errorf("expected error for empty schema, got result: %s", result)
	}
	if !strings.Contains(err.Error(), "no structs") {
		t.Errorf("expected 'no structs' in error, got: %v", err)
	}
}

// TestGenerateDecodeHelpersNameConversion verifies snake_case conversion
func TestGenerateDecodeHelpersNameConversion(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "audio_device",
				Fields: []parser.Field{
					{Name: "device_id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "sample_rate", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check function name conversion
	if !strings.Contains(result, "func decodeAudioDevice(dest *AudioDevice") {
		t.Errorf("missing or incorrect function name conversion, got:\n%s", result)
	}

	// Check field name conversions
	if !strings.Contains(result, "dest.DeviceId =") {
		t.Errorf("missing or incorrect DeviceId field, got:\n%s", result)
	}
	if !strings.Contains(result, "dest.SampleRate =") {
		t.Errorf("missing or incorrect SampleRate field, got:\n%s", result)
	}
}

// TestGenerateDecodeHelpersWithString verifies string field decoding
func TestGenerateDecodeHelpersWithString(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Message",
				Fields: []parser.Field{
					{Name: "text", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check field comment
	if !strings.Contains(result, "// Field: Text (str)") {
		t.Errorf("missing field comment, got:\n%s", result)
	}

	// Check length prefix bounds check
	if !strings.Contains(result, "if *offset + 4 > len(data) {") {
		t.Errorf("missing length prefix bounds check, got:\n%s", result)
	}

	// Check length read
	if !strings.Contains(result, "strLen = binary.LittleEndian.Uint32(data[*offset:])") {
		t.Errorf("missing length read, got:\n%s", result)
	}

	// Check length offset increment
	if !strings.Contains(result, "*offset += 4") {
		t.Errorf("missing length offset increment, got:\n%s", result)
	}

	// Check string bytes bounds check
	if !strings.Contains(result, "if *offset + int(strLen) > len(data) {") {
		t.Errorf("missing string bytes bounds check, got:\n%s", result)
	}

	// Check string decode
	if !strings.Contains(result, "dest.Text = string(data[*offset:*offset+int(strLen)])") {
		t.Errorf("missing string decode, got:\n%s", result)
	}

	// Check string offset increment
	if !strings.Contains(result, "*offset += int(strLen)") {
		t.Errorf("missing string offset increment, got:\n%s", result)
	}
}

// TestGenerateDecodeHelpersWithMultipleStrings verifies multiple string fields
func TestGenerateDecodeHelpersWithMultipleStrings(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Plugin",
				Fields: []parser.Field{
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{Name: "vendor", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check both fields present
	if !strings.Contains(result, "// Field: Name (str)") {
		t.Errorf("missing Name field comment")
	}
	if !strings.Contains(result, "// Field: Vendor (str)") {
		t.Errorf("missing Vendor field comment")
	}

	// Check both decode assignments
	if !strings.Contains(result, "dest.Name = string(data[*offset:*offset+int(strLen)])") {
		t.Errorf("missing Name decode, got:\n%s", result)
	}
	if !strings.Contains(result, "dest.Vendor = string(data[*offset:*offset+int(strLen)])") {
		t.Errorf("missing Vendor decode, got:\n%s", result)
	}

	// Verify order (Name before Vendor)
	nameIndex := strings.Index(result, "dest.Name =")
	vendorIndex := strings.Index(result, "dest.Vendor =")
	if nameIndex == -1 || vendorIndex == -1 {
		t.Fatalf("missing field assignments")
	}
	if nameIndex >= vendorIndex {
		t.Errorf("fields not in correct order: Name@%d, Vendor@%d", nameIndex, vendorIndex)
	}
}

// TestGenerateDecodeHelpersMixedTypes verifies mixed primitive and string fields
func TestGenerateDecodeHelpersMixedTypes(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Device",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{Name: "enabled", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all field types present
	if !strings.Contains(result, "// Field: Id (u32)") {
		t.Errorf("missing Id field comment")
	}
	if !strings.Contains(result, "// Field: Name (str)") {
		t.Errorf("missing Name field comment")
	}
	if !strings.Contains(result, "// Field: Enabled (bool)") {
		t.Errorf("missing Enabled field comment")
	}

	// Check decode logic for each type
	if !strings.Contains(result, "dest.Id = binary.LittleEndian.Uint32(data[*offset:])") {
		t.Errorf("missing u32 decode")
	}
	if !strings.Contains(result, "strLen = binary.LittleEndian.Uint32(data[*offset:])") {
		t.Errorf("missing string length read")
	}
	if !strings.Contains(result, "dest.Name = string(data[*offset:*offset+int(strLen)])") {
		t.Errorf("missing string decode")
	}
	if !strings.Contains(result, "dest.Enabled = data[*offset] != 0") {
		t.Errorf("missing bool decode")
	}

	// Verify order
	idIndex := strings.Index(result, "dest.Id =")
	nameIndex := strings.Index(result, "dest.Name =")
	enabledIndex := strings.Index(result, "dest.Enabled =")

	if idIndex >= nameIndex || nameIndex >= enabledIndex {
		t.Errorf("fields not in correct order: Id@%d, Name@%d, Enabled@%d", idIndex, nameIndex, enabledIndex)
	}
}

// TestGenerateDecodeHelpersStringBoundsChecks verifies string has two bounds checks
func TestGenerateDecodeHelpersStringBoundsChecks(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Test",
				Fields: []parser.Field{
					{Name: "text", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 2 bounds checks for string: one for length, one for bytes
	boundsCheckCount := strings.Count(result, "if *offset +")
	if boundsCheckCount != 2 {
		t.Errorf("expected 2 bounds checks for string field, found %d", boundsCheckCount)
	}

	// Should have 2 EOF errors
	eofErrorCount := strings.Count(result, "return ErrUnexpectedEOF")
	if eofErrorCount != 2 {
		t.Errorf("expected 2 EOF errors for string field, found %d", eofErrorCount)
	}
}

// TestGenerateDecodeHelpersAllTypesIncludingString verifies complete type coverage
func TestGenerateDecodeHelpersAllTypesIncludingString(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Complete",
				Fields: []parser.Field{
					{Name: "u8_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
					{Name: "u16_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u16"}},
					{Name: "u32_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "u64_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u64"}},
					{Name: "i8_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i8"}},
					{Name: "i16_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i16"}},
					{Name: "i32_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
					{Name: "i64_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i64"}},
					{Name: "f32_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "f64_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"}},
					{Name: "bool_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}},
					{Name: "str_val", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all 12 primitive types present
	expectedFields := []string{
		"// Field: U8Val (u8)",
		"// Field: U16Val (u16)",
		"// Field: U32Val (u32)",
		"// Field: U64Val (u64)",
		"// Field: I8Val (i8)",
		"// Field: I16Val (i16)",
		"// Field: I32Val (i32)",
		"// Field: I64Val (i64)",
		"// Field: F32Val (f32)",
		"// Field: F64Val (f64)",
		"// Field: BoolVal (bool)",
		"// Field: StrVal (str)",
	}

	for _, expected := range expectedFields {
		if !strings.Contains(result, expected) {
			t.Errorf("missing field comment: %q", expected)
		}
	}

	// Check string decode present
	if !strings.Contains(result, "dest.StrVal = string(data[*offset:*offset+int(strLen)])") {
		t.Errorf("missing string decode for StrVal")
	}
}

// TestGenerateDecodeHelpersWithPrimitiveArray verifies primitive array decoding
func TestGenerateDecodeHelpersWithPrimitiveArray(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Data",
				Fields: []parser.Field{
					{
						Name: "values",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"},
						},
					},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check field comment
	if !strings.Contains(result, "// Field: Values ([]u32)") {
		t.Errorf("missing field comment, got:\n%s", result)
	}

	// Check count bounds check
	if !strings.Contains(result, "if *offset + 4 > len(data) {") {
		t.Errorf("missing count bounds check")
	}

	// Check count read
	if !strings.Contains(result, "arrCount = binary.LittleEndian.Uint32(data[*offset:])") {
		t.Errorf("missing count read")
	}

	// Check array size validation
	if !strings.Contains(result, "err = ctx.checkArraySize(arrCount)") {
		t.Errorf("missing array size check")
	}

	// Check array allocation
	if !strings.Contains(result, "dest.Values = make([]uint32, arrCount)") {
		t.Errorf("missing array allocation, got:\n%s", result)
	}

	// Check loop
	if !strings.Contains(result, "for i := uint32(0); i < arrCount; i++ {") {
		t.Errorf("missing loop, got:\n%s", result)
	}

	// Check element decode
	if !strings.Contains(result, "dest.Values[i] = binary.LittleEndian.Uint32(data[*offset:])") {
		t.Errorf("missing element decode, got:\n%s", result)
	}
}

// TestGenerateDecodeHelpersWithStringArray verifies string array decoding
func TestGenerateDecodeHelpersWithStringArray(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "StringList",
				Fields: []parser.Field{
					{
						Name: "items",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"},
						},
					},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check field comment
	if !strings.Contains(result, "// Field: Items ([]str)") {
		t.Errorf("missing field comment")
	}

	// Check array allocation
	if !strings.Contains(result, "dest.Items = make([]string, arrCount)") {
		t.Errorf("missing array allocation, got:\n%s", result)
	}

	// Check string element decode (length + bytes)
	if !strings.Contains(result, "strLen := binary.LittleEndian.Uint32(data[*offset:])") {
		t.Errorf("missing string length read")
	}
	if !strings.Contains(result, "dest.Items[i] = string(data[*offset:*offset+int(strLen)])") {
		t.Errorf("missing string decode, got:\n%s", result)
	}
}

// TestGenerateDecodeHelpersWithMultiplePrimitiveArrays verifies all primitive array types
func TestGenerateDecodeHelpersWithMultiplePrimitiveArrays(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Arrays",
				Fields: []parser.Field{
					{Name: "bytes", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}}},
					{Name: "ints", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}}},
					{Name: "floats", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"}}},
					{Name: "flags", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all array allocations
	if !strings.Contains(result, "dest.Bytes = make([]uint8, arrCount)") {
		t.Errorf("missing u8 array allocation")
	}
	if !strings.Contains(result, "dest.Ints = make([]int32, arrCount)") {
		t.Errorf("missing i32 array allocation")
	}
	if !strings.Contains(result, "dest.Floats = make([]float64, arrCount)") {
		t.Errorf("missing f64 array allocation")
	}
	if !strings.Contains(result, "dest.Flags = make([]bool, arrCount)") {
		t.Errorf("missing bool array allocation")
	}

	// Check element decodes
	if !strings.Contains(result, "dest.Bytes[i] = uint8(data[*offset])") {
		t.Errorf("missing u8 element decode")
	}
	if !strings.Contains(result, "dest.Ints[i] = int32(binary.LittleEndian.Uint32(data[*offset:]))") {
		t.Errorf("missing i32 element decode")
	}
	if !strings.Contains(result, "dest.Floats[i] = math.Float64frombits(binary.LittleEndian.Uint64(data[*offset:]))") {
		t.Errorf("missing f64 element decode")
	}
	if !strings.Contains(result, "dest.Flags[i] = data[*offset] != 0") {
		t.Errorf("missing bool element decode")
	}
}

// TestGenerateDecodeHelpersMixedFieldsWithArrays verifies mixed field types
func TestGenerateDecodeHelpersMixedFieldsWithArrays(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Mixed",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "values", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}}},
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify order
	idIndex := strings.Index(result, "// Field: Id (u32)")
	valuesIndex := strings.Index(result, "// Field: Values ([]f32)")
	nameIndex := strings.Index(result, "// Field: Name (str)")

	if idIndex == -1 || valuesIndex == -1 || nameIndex == -1 {
		t.Fatalf("missing field comments")
	}

	if idIndex >= valuesIndex || valuesIndex >= nameIndex {
		t.Errorf("fields not in correct order: Id@%d, Values@%d, Name@%d", idIndex, valuesIndex, nameIndex)
	}
}

// TestGenerateDecodeHelpersArrayCheckCount verifies checkArraySize call
func TestGenerateDecodeHelpersArrayCheckCount(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Test",
				Fields: []parser.Field{
					{
						Name: "data",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"},
						},
					},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must call ctx.checkArraySize
	if !strings.Contains(result, "err = ctx.checkArraySize(arrCount)") {
		t.Errorf("missing checkArraySize call")
	}
	if !strings.Contains(result, "if err != nil") && !strings.Contains(result, "return err") {
		t.Errorf("missing error check after checkArraySize")
	}
}

// TestGenerateDecodeHelpersArrayLoopStructure verifies loop structure
func TestGenerateDecodeHelpersArrayLoopStructure(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "List",
				Fields: []parser.Field{
					{
						Name: "items",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u16"},
						},
					},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify loop structure elements in order
	expectedOrder := []string{
		"arrCount = binary.LittleEndian.Uint32(data[*offset:])",
		"ctx.checkArraySize(arrCount)",
		"dest.Items = make([]uint16, arrCount)",
		"for i := uint32(0); i < arrCount; i++ {",
		"dest.Items[i] = binary.LittleEndian.Uint16(data[*offset:])",
	}

	lastIndex := -1
	for i, expected := range expectedOrder {
		index := strings.Index(result, expected)
		if index == -1 {
			t.Errorf("missing expected element %d: %q", i, expected)
			continue
		}
		if index <= lastIndex {
			t.Errorf("element %d (%q) not in correct order (index %d <= %d)", i, expected, index, lastIndex)
		}
		lastIndex = index
	}
}

// TestGenerateDecodeHelpersAllArrayPrimitiveTypes verifies all 12 primitive types in arrays
func TestGenerateDecodeHelpersAllArrayPrimitiveTypes(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "AllArrays",
				Fields: []parser.Field{
					{Name: "u8s", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}}},
					{Name: "u16s", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u16"}}},
					{Name: "u32s", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}}},
					{Name: "u64s", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u64"}}},
					{Name: "i8s", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i8"}}},
					{Name: "i16s", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i16"}}},
					{Name: "i32s", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}}},
					{Name: "i64s", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i64"}}},
					{Name: "f32s", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}}},
					{Name: "f64s", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"}}},
					{Name: "bools", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}}},
					{Name: "strs", Type: parser.TypeExpr{Kind: parser.TypeKindArray, Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all array allocations with correct types
	expectedAllocations := []string{
		"dest.U8s = make([]uint8, arrCount)",
		"dest.U16s = make([]uint16, arrCount)",
		"dest.U32s = make([]uint32, arrCount)",
		"dest.U64s = make([]uint64, arrCount)",
		"dest.I8s = make([]int8, arrCount)",
		"dest.I16s = make([]int16, arrCount)",
		"dest.I32s = make([]int32, arrCount)",
		"dest.I64s = make([]int64, arrCount)",
		"dest.F32s = make([]float32, arrCount)",
		"dest.F64s = make([]float64, arrCount)",
		"dest.Bools = make([]bool, arrCount)",
		"dest.Strs = make([]string, arrCount)",
	}

	for _, expected := range expectedAllocations {
		if !strings.Contains(result, expected) {
			t.Errorf("missing allocation: %q", expected)
		}
	}

	// Check all element decodes
	expectedDecodes := []string{
		"dest.U8s[i] = uint8(data[*offset])",
		"dest.U16s[i] = binary.LittleEndian.Uint16(data[*offset:])",
		"dest.U32s[i] = binary.LittleEndian.Uint32(data[*offset:])",
		"dest.U64s[i] = binary.LittleEndian.Uint64(data[*offset:])",
		"dest.I8s[i] = int8(data[*offset])",
		"dest.I16s[i] = int16(binary.LittleEndian.Uint16(data[*offset:]))",
		"dest.I32s[i] = int32(binary.LittleEndian.Uint32(data[*offset:]))",
		"dest.I64s[i] = int64(binary.LittleEndian.Uint64(data[*offset:]))",
		"dest.F32s[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))",
		"dest.F64s[i] = math.Float64frombits(binary.LittleEndian.Uint64(data[*offset:]))",
		"dest.Bools[i] = data[*offset] != 0",
		"dest.Strs[i] = string(data[*offset:*offset+int(strLen)])",
	}

	for _, expected := range expectedDecodes {
		if !strings.Contains(result, expected) {
			t.Errorf("missing element decode: %q", expected)
		}
	}
}

// TestGenerateDecodeHelpersWithNamedType verifies nested struct decoding
func TestGenerateDecodeHelpersWithNamedType(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Container",
				Fields: []parser.Field{
					{Name: "item", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Item"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check field comment
	if !strings.Contains(result, "// Field: Item (Item)") {
		t.Errorf("missing field comment, got:\n%s", result)
	}

	// Check helper function call
	if !strings.Contains(result, "err = decodeItem(&dest.Item, data, offset, ctx)") {
		t.Errorf("missing helper function call, got:\n%s", result)
	}

	// Check error handling
	if !strings.Contains(result, "if err != nil") || !strings.Contains(result, "return err") {
		t.Errorf("missing error handling")
	}
}

// TestGenerateDecodeHelpersWithMultipleNamedTypes verifies multiple nested structs
func TestGenerateDecodeHelpersWithMultipleNamedTypes(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Complex",
				Fields: []parser.Field{
					{Name: "point", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Point"}},
					{Name: "color", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Color"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check both field comments
	if !strings.Contains(result, "// Field: Point (Point)") {
		t.Errorf("missing Point field comment")
	}
	if !strings.Contains(result, "// Field: Color (Color)") {
		t.Errorf("missing Color field comment")
	}

	// Check both helper calls
	if !strings.Contains(result, "err = decodePoint(&dest.Point, data, offset, ctx)") {
		t.Errorf("missing decodePoint call")
	}
	if !strings.Contains(result, "err = decodeColor(&dest.Color, data, offset, ctx)") {
		t.Errorf("missing decodeColor call")
	}
}

// TestGenerateDecodeHelpersMixedWithNamedTypes verifies mixed field types including named types
func TestGenerateDecodeHelpersMixedWithNamedTypes(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Record",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "data", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Data"}},
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify order
	idIndex := strings.Index(result, "// Field: Id (u32)")
	dataIndex := strings.Index(result, "// Field: Data (Data)")
	nameIndex := strings.Index(result, "// Field: Name (str)")

	if idIndex == -1 || dataIndex == -1 || nameIndex == -1 {
		t.Fatalf("missing field comments")
	}

	if idIndex >= dataIndex || dataIndex >= nameIndex {
		t.Errorf("fields not in correct order: Id@%d, Data@%d, Name@%d", idIndex, dataIndex, nameIndex)
	}

	// Check decode calls
	if !strings.Contains(result, "dest.Id = binary.LittleEndian.Uint32(data[*offset:])") {
		t.Errorf("missing Id decode")
	}
	if !strings.Contains(result, "err = decodeData(&dest.Data, data, offset, ctx)") {
		t.Errorf("missing Data decode")
	}
	if !strings.Contains(result, "dest.Name = string(data[*offset:*offset+int(strLen)])") {
		t.Errorf("missing Name decode")
	}
}

// TestGenerateDecodeHelpersNamedTypeSnakeCase verifies name conversion for nested structs
func TestGenerateDecodeHelpersNamedTypeSnakeCase(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Outer",
				Fields: []parser.Field{
					{Name: "inner_struct", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "inner_data"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check field name conversion
	if !strings.Contains(result, "// Field: InnerStruct (inner_data)") {
		t.Errorf("missing or incorrect field comment, got:\n%s", result)
	}

	// Check helper function name conversion
	if !strings.Contains(result, "err = decodeInnerData(&dest.InnerStruct, data, offset, ctx)") {
		t.Errorf("missing or incorrect helper call, got:\n%s", result)
	}
}

// TestGenerateDecodeHelpersWithNamedTypeArray verifies array of named types
func TestGenerateDecodeHelpersWithNamedTypeArray(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "List",
				Fields: []parser.Field{
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

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check field comment
	if !strings.Contains(result, "// Field: Items ([]Item)") {
		t.Errorf("missing field comment, got:\n%s", result)
	}

	// Check array allocation
	if !strings.Contains(result, "dest.Items = make([]Item, arrCount)") {
		t.Errorf("missing array allocation, got:\n%s", result)
	}

	// Check loop
	if !strings.Contains(result, "for i := uint32(0); i < arrCount; i++ {") {
		t.Errorf("missing loop")
	}

	// Check element decode via helper
	if !strings.Contains(result, "err = decodeItem(&dest.Items[i], data, offset, ctx)") {
		t.Errorf("missing element decode helper call, got:\n%s", result)
	}

	// Check error return in loop
	if !strings.Contains(result, "return err") {
		t.Errorf("missing error return")
	}
}

// TestGenerateDecodeHelpersComplexNesting verifies realistic complex nesting
func TestGenerateDecodeHelpersComplexNesting(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Plugin",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{
						Name: "parameters",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Parameter"},
						},
					},
					{Name: "config", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Config"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all fields present
	expectedComments := []string{
		"// Field: Id (u32)",
		"// Field: Name (str)",
		"// Field: Parameters ([]Parameter)",
		"// Field: Config (Config)",
	}

	for _, expected := range expectedComments {
		if !strings.Contains(result, expected) {
			t.Errorf("missing field comment: %q", expected)
		}
	}

	// Check array of named types
	if !strings.Contains(result, "dest.Parameters = make([]Parameter, arrCount)") {
		t.Errorf("missing Parameters array allocation")
	}
	if !strings.Contains(result, "err = decodeParameter(&dest.Parameters[i], data, offset, ctx)") {
		t.Errorf("missing Parameter element decode")
	}

	// Check nested struct field
	if !strings.Contains(result, "err = decodeConfig(&dest.Config, data, offset, ctx)") {
		t.Errorf("missing Config decode")
	}
}

// TestGenerateDecodeHelpersNamedTypePassesContext verifies context is passed through
func TestGenerateDecodeHelpersNamedTypePassesContext(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Wrapper",
				Fields: []parser.Field{
					{Name: "inner", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Inner"}},
				},
			},
		},
	}

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that data, offset, and ctx are all passed to the helper
	if !strings.Contains(result, "decodeInner(&dest.Inner, data, offset, ctx)") {
		t.Errorf("helper call doesn't pass all required parameters (data, offset, ctx), got:\n%s", result)
	}
}

// TestGenerateDecodeHelpersArrayNamedTypePassesContext verifies context is passed in array elements
func TestGenerateDecodeHelpersArrayNamedTypePassesContext(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Collection",
				Fields: []parser.Field{
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

	result, err := GenerateDecodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify context is passed when decoding array elements
	if !strings.Contains(result, "decodeItem(&dest.Items[i], data, offset, ctx)") {
		t.Errorf("array element helper call doesn't pass all required parameters, got:\n%s", result)
	}
}
