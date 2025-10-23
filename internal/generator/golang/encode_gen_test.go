package golang

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// TestGenerateEncoderSimple verifies basic encoder generation
func TestGenerateEncoderSimple(t *testing.T) {
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

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check size calculation function
	if !strings.Contains(result, "func calculateDeviceSize(src *Device) int {") {
		t.Error("missing size calculation function")
	}

	// Check public encoder function
	if !strings.Contains(result, "func EncodeDevice(src *Device) ([]byte, error) {") {
		t.Error("missing public encoder function")
	}

	// Check size calculation call
	if !strings.Contains(result, "size := calculateDeviceSize(src)") {
		t.Error("missing size calculation call")
	}

	// Check buffer allocation
	if !strings.Contains(result, "buf := make([]byte, size)") {
		t.Error("missing buffer allocation")
	}

	// Check helper call
	if !strings.Contains(result, "if err := encodeDevice(src, buf, &offset); err != nil {") {
		t.Error("missing helper function call")
	}
}

// TestGenerateEncoderMultipleStructs verifies multiple struct encoding
func TestGenerateEncoderMultipleStructs(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Device",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
			{
				Name: "Plugin",
				Fields: []parser.Field{
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check both size functions
	if !strings.Contains(result, "func calculateDeviceSize(src *Device) int {") {
		t.Error("missing Device size calculation")
	}
	if !strings.Contains(result, "func calculatePluginSize(src *Plugin) int {") {
		t.Error("missing Plugin size calculation")
	}

	// Check both encoder functions
	if !strings.Contains(result, "func EncodeDevice(src *Device) ([]byte, error) {") {
		t.Error("missing Device encoder")
	}
	if !strings.Contains(result, "func EncodePlugin(src *Plugin) ([]byte, error) {") {
		t.Error("missing Plugin encoder")
	}
}

// TestGenerateEncoderSnakeCaseConversion verifies name conversion
func TestGenerateEncoderSnakeCaseConversion(t *testing.T) {
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

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check PascalCase conversion for struct name
	if !strings.Contains(result, "calculateAudioDeviceSize") {
		t.Error("size function name not converted to PascalCase")
	}
	if !strings.Contains(result, "func EncodeAudioDevice(src *AudioDevice)") {
		t.Error("encoder function name not converted to PascalCase")
	}

	// Check field name conversion in size calculation
	// Field: DeviceId comment should be present
	if !strings.Contains(result, "// Field: DeviceId") {
		t.Error("field comment not converted to PascalCase")
	}
}

// TestGenerateEncoderDocComments verifies documentation comments
func TestGenerateEncoderDocComments(t *testing.T) {
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

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check size function doc comment
	if !strings.Contains(result, "// calculateDeviceSize calculates the wire format size for Device") {
		t.Error("missing size calculation doc comment")
	}

	// Check encoder function doc comment
	if !strings.Contains(result, "// EncodeDevice encodes a Device to wire format") {
		t.Error("missing encoder doc comment")
	}

	if !strings.Contains(result, "// It returns the encoded bytes or an error") {
		t.Error("missing encoder doc details")
	}
}

// TestGenerateEncoderSignature verifies correct function signatures
func TestGenerateEncoderSignature(t *testing.T) {
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

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify exact signatures
	if !strings.Contains(result, "func calculateDeviceSize(src *Device) int {") {
		t.Error("size function signature incorrect")
	}

	if !strings.Contains(result, "func EncodeDevice(src *Device) ([]byte, error) {") {
		t.Error("encoder function signature incorrect (should return []byte, error)")
	}

	// Verify helper call signature
	if !strings.Contains(result, "encodeDevice(src, buf, &offset)") {
		t.Error("helper function call incorrect (should pass src, buf, &offset)")
	}
}

// TestGenerateEncoderSizeCalculation verifies size calculation logic
func TestGenerateEncoderSizeCalculation(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Mixed",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "count", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u64"}},
					{Name: "flag", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}},
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check size initialization
	if !strings.Contains(result, "size := 0") {
		t.Error("missing size initialization")
	}

	// Check fixed-size fields
	if !strings.Contains(result, "size += 4") { // u32
		t.Error("missing u32 size calculation")
	}
	if !strings.Contains(result, "size += 8") { // u64
		t.Error("missing u64 size calculation")
	}
	if !strings.Contains(result, "size += 1") { // bool
		t.Error("missing bool size calculation")
	}

	// Check string field (dynamic size)
	if !strings.Contains(result, "size += 4 + len(src.Name)") {
		t.Error("missing string size calculation")
	}

	// Check return
	if !strings.Contains(result, "return size") {
		t.Error("missing size return")
	}
}

// TestGenerateEncoderAllPrimitives verifies all primitive type sizes
func TestGenerateEncoderAllPrimitives(t *testing.T) {
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

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count each size addition (1, 2, 4, 8 byte types)
	oneByteCount := strings.Count(result, "size += 1")
	twoByteCount := strings.Count(result, "size += 2")
	fourByteCount := strings.Count(result, "size += 4")
	eightByteCount := strings.Count(result, "size += 8")

	// u8, i8, bool = 3 × 1 byte
	if oneByteCount < 3 {
		t.Errorf("expected at least 3 one-byte fields, got %d", oneByteCount)
	}

	// u16, i16 = 2 × 2 bytes
	if twoByteCount < 2 {
		t.Errorf("expected at least 2 two-byte fields, got %d", twoByteCount)
	}

	// u32, i32, f32 = 3 × 4 bytes (plus string length prefix)
	if fourByteCount < 3 {
		t.Errorf("expected at least 3 four-byte fields, got %d", fourByteCount)
	}

	// u64, i64, f64 = 3 × 8 bytes
	if eightByteCount < 3 {
		t.Errorf("expected at least 3 eight-byte fields, got %d", eightByteCount)
	}

	// String field
	if !strings.Contains(result, "size += 4 + len(src.StrField)") {
		t.Error("missing string field size calculation")
	}
}

// TestGenerateEncoderNilSchema verifies nil schema handling
func TestGenerateEncoderNilSchema(t *testing.T) {
	_, err := GenerateEncoder(nil)
	if err == nil {
		t.Error("expected error for nil schema")
	}
	if !strings.Contains(err.Error(), "schema is nil") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestGenerateEncoderEmptySchema verifies empty schema handling
func TestGenerateEncoderEmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	_, err := GenerateEncoder(schema)
	if err == nil {
		t.Error("expected error for empty schema")
	}
	if !strings.Contains(err.Error(), "no structs") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestGenerateEncoderWithArray verifies array size calculation
func TestGenerateEncoderWithArray(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Container",
				Fields: []parser.Field{
					{
						Name: "items",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"},
						},
					},
				},
			},
		},
	}

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check array count size (4 bytes)
	if !strings.Contains(result, "size += 4") {
		t.Error("missing array count size")
	}

	// Check array elements size (u32 = 4 bytes each)
	if !strings.Contains(result, "size += len(src.Items) * 4") {
		t.Error("missing array elements size calculation")
	}
}

// TestGenerateEncoderWithStringArray verifies string array size calculation
func TestGenerateEncoderWithStringArray(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "StringList",
				Fields: []parser.Field{
					{
						Name: "names",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"},
						},
					},
				},
			},
		},
	}

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check array count
	if !strings.Contains(result, "size += 4") {
		t.Error("missing array count size")
	}

	// Check loop for string sizes
	if !strings.Contains(result, "for i := range src.Names {") {
		t.Error("missing loop for string array")
	}

	// Check string element size (length prefix + bytes)
	if !strings.Contains(result, "size += 4 + len(src.Names[i])") {
		t.Error("missing string element size calculation")
	}
}

// TestGenerateEncoderWithNestedStruct verifies nested struct size calculation
func TestGenerateEncoderWithNestedStruct(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Container",
				Fields: []parser.Field{
					{Name: "inner", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Inner"}},
				},
			},
		},
	}

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check nested struct size calculation call
	if !strings.Contains(result, "size += calculateInnerSize(&src.Inner)") {
		t.Error("missing nested struct size calculation call")
	}
}

// TestGenerateEncoderWithStructArray verifies array of structs size calculation
func TestGenerateEncoderWithStructArray(t *testing.T) {
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

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check array count
	if !strings.Contains(result, "size += 4") {
		t.Error("missing array count size")
	}

	// Check loop for struct array
	if !strings.Contains(result, "for i := range src.Items {") {
		t.Error("missing loop for struct array")
	}

	// Check struct element size calculation
	if !strings.Contains(result, "size += calculateItemSize(&src.Items[i])") {
		t.Error("missing struct element size calculation call")
	}
}

// TestGenerateEncoderFieldComments verifies field comments in size calculation
func TestGenerateEncoderFieldComments(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Device",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateEncoder(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check field comments
	if !strings.Contains(result, "// Field: Id") {
		t.Error("missing Id field comment")
	}
	if !strings.Contains(result, "// Field: Name") {
		t.Error("missing Name field comment")
	}
}

func TestGenerateEncodeHelpers_SimplePrimitives(t *testing.T) {
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

	result, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check helper function signature
	if !strings.Contains(result, "func encodeDevice(src *Device, buf []byte, offset *int) error {") {
		t.Error("missing encodeDevice helper function")
	}

	// Check u32 encoding
	if !strings.Contains(result, "binary.LittleEndian.PutUint32(buf[*offset:], src.Id)") {
		t.Error("missing u32 encoding for Id")
	}
	if !strings.Contains(result, "*offset += 4") {
		t.Error("missing offset increment for u32")
	}

	// Check u16 encoding
	if !strings.Contains(result, "binary.LittleEndian.PutUint16(buf[*offset:], src.Count)") {
		t.Error("missing u16 encoding for Count")
	}
	if !strings.Contains(result, "*offset += 2") {
		t.Error("missing offset increment for u16")
	}

	// Check bool encoding
	if !strings.Contains(result, "if src.Active {") {
		t.Error("missing bool if statement")
	}
	if !strings.Contains(result, "buf[*offset] = 1") {
		t.Error("missing bool true encoding")
	}
	if !strings.Contains(result, "buf[*offset] = 0") {
		t.Error("missing bool false encoding")
	}
	if !strings.Contains(result, "*offset++") {
		t.Error("missing offset increment for bool")
	}
}

func TestGenerateEncodeHelpers_AllPrimitiveTypes(t *testing.T) {
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
				},
			},
		},
	}

	result, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check u8 (single byte)
	if !strings.Contains(result, "buf[*offset] = src.U8Field") {
		t.Error("missing u8 encoding")
	}

	// Check signed integer casts
	if !strings.Contains(result, "binary.LittleEndian.PutUint16(buf[*offset:], uint16(src.I16Field))") {
		t.Error("missing i16 cast and encoding")
	}
	if !strings.Contains(result, "binary.LittleEndian.PutUint32(buf[*offset:], uint32(src.I32Field))") {
		t.Error("missing i32 cast and encoding")
	}
	if !strings.Contains(result, "binary.LittleEndian.PutUint64(buf[*offset:], uint64(src.I64Field))") {
		t.Error("missing i64 cast and encoding")
	}

	// Check float conversions
	if !strings.Contains(result, "math.Float32bits(src.F32Field)") {
		t.Error("missing f32 conversion")
	}
	if !strings.Contains(result, "math.Float64bits(src.F64Field)") {
		t.Error("missing f64 conversion")
	}
}

func TestGenerateEncodeHelpers_StringField(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Message",
				Fields: []parser.Field{
					{Name: "content", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check length prefix
	if !strings.Contains(result, "binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Content)))") {
		t.Error("missing string length prefix")
	}

	// Check copy
	if !strings.Contains(result, "copy(buf[*offset:], src.Content)") {
		t.Error("missing string copy")
	}

	// Check offset updates
	if !strings.Contains(result, "*offset += 4") {
		t.Error("missing offset increment for length prefix")
	}
	if !strings.Contains(result, "*offset += len(src.Content)") {
		t.Error("missing offset increment for string bytes")
	}
}

func TestGenerateEncodeHelpers_PrimitiveArray(t *testing.T) {
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

	result, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check count encoding
	if !strings.Contains(result, "binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Values)))") {
		t.Error("missing array count encoding")
	}

	// Check for either bulk copy optimization OR element-by-element loop
	hasBulkCopy := strings.Contains(result, "// Bulk copy optimization") || strings.Contains(result, "unsafe.Slice")
	hasLoop := strings.Contains(result, "for i := range src.Values {")

	if !hasBulkCopy && !hasLoop {
		t.Error("missing array encoding (neither bulk copy nor loop found)")
	}

	// If using loop, check element encoding
	if hasLoop && !strings.Contains(result, "binary.LittleEndian.PutUint32(buf[*offset:], src.Values[i])") {
		t.Error("missing array element encoding in loop")
	}
}

func TestGenerateEncodeHelpers_StringArray(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Tags",
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

	result, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check array count
	if !strings.Contains(result, "binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Items)))") {
		t.Error("missing array count")
	}

	// Check string element length
	if !strings.Contains(result, "binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Items[i])))") {
		t.Error("missing string element length")
	}

	// Check string element copy
	if !strings.Contains(result, "copy(buf[*offset:], src.Items[i])") {
		t.Error("missing string element copy")
	}

	// Check offset updates
	if !strings.Contains(result, "*offset += len(src.Items[i])") {
		t.Error("missing offset increment for string element")
	}
}

func TestGenerateEncodeHelpers_NestedStruct(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Address",
				Fields: []parser.Field{
					{Name: "street", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
			{
				Name: "Person",
				Fields: []parser.Field{
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{Name: "address", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Address"}},
				},
			},
		},
	}

	result, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check recursive call
	if !strings.Contains(result, "if err := encodeAddress(&src.Address, buf, offset); err != nil {") {
		t.Error("missing recursive encode call for nested struct")
	}

	// Check error propagation
	if !strings.Contains(result, "return err") {
		t.Error("missing error propagation for nested struct")
	}

	// Check both helper functions exist
	if !strings.Contains(result, "func encodeAddress(") {
		t.Error("missing encodeAddress helper")
	}
	if !strings.Contains(result, "func encodePerson(") {
		t.Error("missing encodePerson helper")
	}
}

func TestGenerateEncodeHelpers_StructArray(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Item",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
			{
				Name: "Container",
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

	result, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check array count
	if !strings.Contains(result, "binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Items)))") {
		t.Error("missing array count for struct array")
	}

	// Check loop with recursive call
	if !strings.Contains(result, "for i := range src.Items {") {
		t.Error("missing loop for struct array")
	}
	if !strings.Contains(result, "if err := encodeItem(&src.Items[i], buf, offset); err != nil {") {
		t.Error("missing recursive call for struct array element")
	}

	// Check error propagation
	if !strings.Contains(result, "return err") {
		t.Error("missing error propagation for struct array")
	}
}

func TestGenerateEncodeHelpers_MixedFields(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Complex",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
					{
						Name: "values",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"},
						},
					},
					{Name: "active", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}},
				},
			},
		},
	}

	result, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all field types are encoded
	if !strings.Contains(result, "src.Id") {
		t.Error("missing Id field encoding")
	}
	if !strings.Contains(result, "src.Name") {
		t.Error("missing Name field encoding")
	}
	if !strings.Contains(result, "src.Values") {
		t.Error("missing Values field encoding")
	}
	if !strings.Contains(result, "src.Active") {
		t.Error("missing Active field encoding")
	}

	// Check field comments
	if !strings.Contains(result, "// Field: Id (u32)") {
		t.Error("missing Id field comment")
	}
	if !strings.Contains(result, "// Field: Name (str)") {
		t.Error("missing Name field comment")
	}
	if !strings.Contains(result, "// Field: Values ([]f64)") {
		t.Error("missing Values field comment")
	}
	if !strings.Contains(result, "// Field: Active (bool)") {
		t.Error("missing Active field comment")
	}
}

func TestGenerateEncodeHelpers_SnakeCaseConversion(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "test_struct",
				Fields: []parser.Field{
					{Name: "some_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "another_long_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check function name conversion
	if !strings.Contains(result, "func encodeTestStruct(") {
		t.Error("missing snake_case to camelCase conversion for function name")
	}

	// Check field name conversion
	if !strings.Contains(result, "src.SomeField") {
		t.Error("missing snake_case to PascalCase conversion for field name")
	}
	if !strings.Contains(result, "src.AnotherLongField") {
		t.Error("missing snake_case to PascalCase conversion for long field name")
	}
}

func TestGenerateEncodeHelpers_MultipleStructs(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "First",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
			{
				Name: "Second",
				Fields: []parser.Field{
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
			{
				Name: "Third",
				Fields: []parser.Field{
					{Name: "active", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}},
				},
			},
		},
	}

	result, err := GenerateEncodeHelpers(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all helper functions exist
	if !strings.Contains(result, "func encodeFirst(") {
		t.Error("missing encodeFirst helper")
	}
	if !strings.Contains(result, "func encodeSecond(") {
		t.Error("missing encodeSecond helper")
	}
	if !strings.Contains(result, "func encodeThird(") {
		t.Error("missing encodeThird helper")
	}
}

func TestGenerateEncodeHelpers_EmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	_, err := GenerateEncodeHelpers(schema)
	if err == nil {
		t.Fatal("expected error for empty schema, got nil")
	}

	if !strings.Contains(err.Error(), "schema has no structs") {
		t.Errorf("expected 'schema has no structs' error, got: %v", err)
	}
}
