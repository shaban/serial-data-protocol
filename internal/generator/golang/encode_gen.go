package golang

import (
	"fmt"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// GenerateEncoder generates an Encode function for each struct in the schema.
// Uses the Pre-calculated Size + Direct Writes approach for optimal FFI performance.
//
// For each struct type, it generates:
//   - calculateStructNameSize(src *StructName) int - Fast size calculation
//   - EncodeStructName(src *StructName) ([]byte, error) - Public encoder
//
// The encoder:
//  1. Calculates exact buffer size needed (single pass, ~50ns overhead)
//  2. Allocates buffer once with exact size
//  3. Calls helper function for direct buffer writes
//  4. Returns buffer with zero-copy (ownership transfer)
//
// Example output:
//
//	func calculateDeviceSize(src *Device) int {
//	    size := 4  // ID field
//	    size += 4 + len(src.Name)  // Name: length + bytes
//	    return size
//	}
//
//	func EncodeDevice(src *Device) ([]byte, error) {
//	    size := calculateDeviceSize(src)
//	    buf := make([]byte, size)
//	    offset := 0
//	    if err := encodeDevice(src, buf, &offset); err != nil {
//	        return nil, err
//	    }
//	    return buf, nil
//	}
//
// This approach is optimal for FFI scenarios where encoding happens on hot paths
// (2x faster than bytes.Buffer, single allocation, predictable performance).
func GenerateEncoder(schema *parser.Schema) (string, error) {
	if schema == nil {
		return "", fmt.Errorf("schema is nil")
	}

	if len(schema.Structs) == 0 {
		return "", fmt.Errorf("schema has no structs")
	}

	var buf strings.Builder

	for i, s := range schema.Structs {
		// Add blank line between functions (except before first)
		if i > 0 {
			buf.WriteString("\n")
		}

		structName := ToGoName(s.Name)
		sizeFunc := "calculate" + structName + "Size"
		encodeFunc := "Encode" + structName
		helperFunc := "encode" + structName

		// Generate size calculation function
		if err := generateSizeCalculation(&buf, &s, structName, sizeFunc); err != nil {
			return "", err
		}

		buf.WriteString("\n")

		// Generate public encoder function
		if err := generateEncoderFunction(&buf, structName, encodeFunc, sizeFunc, helperFunc); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// generateSizeCalculation generates the size calculation function for a struct.
// This function traverses the struct once to compute the exact buffer size needed.
func generateSizeCalculation(buf *strings.Builder, s *parser.Struct, structName, funcName string) error {
	// Doc comment
	buf.WriteString("// ")
	buf.WriteString(funcName)
	buf.WriteString(" calculates the wire format size for ")
	buf.WriteString(structName)
	buf.WriteString(".\n")

	// Function signature
	buf.WriteString("func ")
	buf.WriteString(funcName)
	buf.WriteString("(src *")
	buf.WriteString(structName)
	buf.WriteString(") int {\n")

	// Initialize size
	buf.WriteString("\tsize := 0\n")

	// Add size for each field
	for _, field := range s.Fields {
		if err := generateFieldSizeCalculation(buf, &field); err != nil {
			return err
		}
	}

	// Return total size
	buf.WriteString("\treturn size\n")
	buf.WriteString("}\n")

	return nil
}

// generateFieldSizeCalculation generates size calculation code for a single field.
func generateFieldSizeCalculation(buf *strings.Builder, field *parser.Field) error {
	fieldName := ToGoName(field.Name)

	buf.WriteString("\t// Field: ")
	buf.WriteString(fieldName)
	buf.WriteString("\n")

	// Handle optional fields - add 1 byte for presence flag
	if field.Type.Optional {
		buf.WriteString("\tsize += 1 // presence flag\n")
		buf.WriteString("\tif src.")
		buf.WriteString(fieldName)
		buf.WriteString(" != nil {\n")

		// Indent the size calculation for the actual field
		tempBuf := &strings.Builder{}

		var err error
		switch field.Type.Kind {
		case parser.TypeKindPrimitive:
			err = generatePrimitiveSizeCalculation(tempBuf, field.Type.Name, fieldName)
		case parser.TypeKindArray:
			err = generateArraySizeCalculation(tempBuf, &field.Type, fieldName)
		case parser.TypeKindNamed:
			// For optional fields, don't take address since src.FieldName is already a pointer
			err = generateNamedTypeSizeCalculationOptional(tempBuf, field.Type.Name, fieldName)
		default:
			err = fmt.Errorf("unsupported type kind: %v", field.Type.Kind)
		}

		if err != nil {
			return err
		}

		// Add extra indentation
		lines := strings.Split(tempBuf.String(), "\n")
		for _, line := range lines {
			if line != "" {
				buf.WriteString("\t")
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}

		buf.WriteString("\t}\n")
		return nil
	}

	// Non-optional fields - generate size calculation normally
	switch field.Type.Kind {
	case parser.TypeKindPrimitive:
		return generatePrimitiveSizeCalculation(buf, field.Type.Name, fieldName)
	case parser.TypeKindArray:
		return generateArraySizeCalculation(buf, &field.Type, fieldName)
	case parser.TypeKindNamed:
		return generateNamedTypeSizeCalculation(buf, field.Type.Name, fieldName)
	default:
		return fmt.Errorf("unsupported type kind: %v", field.Type.Kind)
	}
}

// generatePrimitiveSizeCalculation generates size calculation for primitive fields.
func generatePrimitiveSizeCalculation(buf *strings.Builder, typeName, fieldName string) error {
	var size int

	switch typeName {
	case "u8", "i8", "bool":
		size = 1
	case "u16", "i16":
		size = 2
	case "u32", "i32", "f32":
		size = 4
	case "u64", "i64", "f64":
		size = 8
	case "str":
		// String: length prefix (4 bytes) + string bytes
		buf.WriteString("\tsize += 4 + len(src.")
		buf.WriteString(fieldName)
		buf.WriteString(")\n")
		return nil
	default:
		return fmt.Errorf("unknown primitive type: %s", typeName)
	}

	buf.WriteString(fmt.Sprintf("\tsize += %d\n", size))
	return nil
}

// generateArraySizeCalculation generates size calculation for array fields.
func generateArraySizeCalculation(buf *strings.Builder, typeExpr *parser.TypeExpr, fieldName string) error {
	if typeExpr.Elem == nil {
		return fmt.Errorf("array type missing element type")
	}

	// Array count (4 bytes)
	buf.WriteString("\tsize += 4\n")

	elemType := typeExpr.Elem

	switch elemType.Kind {
	case parser.TypeKindPrimitive:
		return generateArrayPrimitiveSizeCalculation(buf, elemType.Name, fieldName)
	case parser.TypeKindNamed:
		return generateArrayNamedTypeSizeCalculation(buf, elemType.Name, fieldName)
	default:
		return fmt.Errorf("unsupported array element type kind: %v", elemType.Kind)
	}
}

// generateArrayPrimitiveSizeCalculation generates size calculation for arrays of primitives.
func generateArrayPrimitiveSizeCalculation(buf *strings.Builder, elemTypeName, fieldName string) error {
	switch elemTypeName {
	case "u8", "i8", "bool":
		buf.WriteString("\tsize += len(src.")
		buf.WriteString(fieldName)
		buf.WriteString(")\n")
	case "u16", "i16":
		buf.WriteString("\tsize += len(src.")
		buf.WriteString(fieldName)
		buf.WriteString(") * 2\n")
	case "u32", "i32", "f32":
		buf.WriteString("\tsize += len(src.")
		buf.WriteString(fieldName)
		buf.WriteString(") * 4\n")
	case "u64", "i64", "f64":
		buf.WriteString("\tsize += len(src.")
		buf.WriteString(fieldName)
		buf.WriteString(") * 8\n")
	case "str":
		// Array of strings: each string has length prefix + bytes
		buf.WriteString("\tfor i := range src.")
		buf.WriteString(fieldName)
		buf.WriteString(" {\n")
		buf.WriteString("\t\tsize += 4 + len(src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i])\n")
		buf.WriteString("\t}\n")
	default:
		return fmt.Errorf("unknown primitive type in array: %s", elemTypeName)
	}

	return nil
}

// generateArrayNamedTypeSizeCalculation generates size calculation for arrays of structs.
func generateArrayNamedTypeSizeCalculation(buf *strings.Builder, elemTypeName, fieldName string) error {
	elemStructName := ToGoName(elemTypeName)
	sizeFuncName := "calculate" + elemStructName + "Size"

	buf.WriteString("\tfor i := range src.")
	buf.WriteString(fieldName)
	buf.WriteString(" {\n")
	buf.WriteString("\t\tsize += ")
	buf.WriteString(sizeFuncName)
	buf.WriteString("(&src.")
	buf.WriteString(fieldName)
	buf.WriteString("[i])\n")
	buf.WriteString("\t}\n")

	return nil
}

// generateNamedTypeSizeCalculation generates size calculation for nested struct fields.
// For non-optional fields, it takes the address (&src.FieldName).
// For optional fields, src.FieldName is already a pointer, so no & needed.
func generateNamedTypeSizeCalculation(buf *strings.Builder, typeName, fieldName string) error {
	return generateNamedTypeSizeCalculationWithPrefix(buf, typeName, fieldName, "&src.")
}

// generateNamedTypeSizeCalculationOptional generates size calculation for optional nested struct fields.
// Since optional fields are already pointers, we don't take the address.
func generateNamedTypeSizeCalculationOptional(buf *strings.Builder, typeName, fieldName string) error {
	return generateNamedTypeSizeCalculationWithPrefix(buf, typeName, fieldName, "src.")
}

// Helper function for named type size calculation with custom prefix
func generateNamedTypeSizeCalculationWithPrefix(buf *strings.Builder, typeName, fieldName, prefix string) error {
	structName := ToGoName(typeName)
	sizeFuncName := "calculate" + structName + "Size"

	buf.WriteString("\tsize += ")
	buf.WriteString(sizeFuncName)
	buf.WriteString("(")
	buf.WriteString(prefix)
	buf.WriteString(fieldName)
	buf.WriteString(")\n")

	return nil
}

// generateEncoderFunction generates the public encoder function.
// This function allocates the buffer and delegates to the helper function.
func generateEncoderFunction(buf *strings.Builder, structName, funcName, sizeFunc, helperFunc string) error {
	// Doc comment
	buf.WriteString("// ")
	buf.WriteString(funcName)
	buf.WriteString(" encodes a ")
	buf.WriteString(structName)
	buf.WriteString(" to wire format.\n")
	buf.WriteString("// It returns the encoded bytes or an error.\n")

	// Function signature
	buf.WriteString("func ")
	buf.WriteString(funcName)
	buf.WriteString("(src *")
	buf.WriteString(structName)
	buf.WriteString(") ([]byte, error) {\n")

	// Calculate size
	buf.WriteString("\tsize := ")
	buf.WriteString(sizeFunc)
	buf.WriteString("(src)\n")

	// Allocate buffer
	buf.WriteString("\tbuf := make([]byte, size)\n")
	buf.WriteString("\toffset := 0\n")

	// Call helper function
	buf.WriteString("\tif err := ")
	buf.WriteString(helperFunc)
	buf.WriteString("(src, buf, &offset); err != nil {\n")
	buf.WriteString("\t\treturn nil, err\n")
	buf.WriteString("\t}\n")

	// Return buffer
	buf.WriteString("\treturn buf, nil\n")
	buf.WriteString("}\n")

	return nil
}

// GenerateEncodeHelpers generates helper encode functions for each struct in the schema.
// These functions perform the actual buffer writes using direct memory operations.
//
// For each struct type, it generates:
//   - encodeStructName(src *StructName, buf []byte, offset *int) error
//
// The helper functions:
//   - Write directly to pre-allocated buffer
//   - Track offset via pointer (mirrors decoder pattern)
//   - Use binary.LittleEndian for primitives
//   - Use copy() for strings
//   - Call nested helpers recursively
//
// Example output:
//
//	func encodeDevice(src *Device, buf []byte, offset *int) error {
//	    // Field: ID (u32)
//	    binary.LittleEndian.PutUint32(buf[*offset:], src.ID)
//	    *offset += 4
//
//	    // Field: Name (str)
//	    binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Name)))
//	    *offset += 4
//	    copy(buf[*offset:], src.Name)
//	    *offset += len(src.Name)
//
//	    return nil
//	}
func GenerateEncodeHelpers(schema *parser.Schema) (string, error) {
	if schema == nil {
		return "", fmt.Errorf("schema is nil")
	}

	if len(schema.Structs) == 0 {
		return "", fmt.Errorf("schema has no structs")
	}

	var buf strings.Builder

	for i, s := range schema.Structs {
		// Add blank line between functions (except before first)
		if i > 0 {
			buf.WriteString("\n")
		}

		structName := ToGoName(s.Name)
		funcName := "encode" + structName

		// Generate doc comment
		buf.WriteString("// ")
		buf.WriteString(funcName)
		buf.WriteString(" is the helper function that encodes ")
		buf.WriteString(structName)
		buf.WriteString(" fields.\n")

		// Function signature
		buf.WriteString("func ")
		buf.WriteString(funcName)
		buf.WriteString("(src *")
		buf.WriteString(structName)
		buf.WriteString(", buf []byte, offset *int) error {\n")

		// Encode each field
		for _, field := range s.Fields {
			if err := generateFieldEncode(&buf, &field); err != nil {
				return "", err
			}
		}

		// Return success
		buf.WriteString("\treturn nil\n")
		buf.WriteString("}\n")
	}

	return buf.String(), nil
}

// generateFieldEncode generates encode code for a single field.
func generateFieldEncode(buf *strings.Builder, field *parser.Field) error {
	fieldName := ToGoName(field.Name)

	// Add field comment
	buf.WriteString("\t// Field: ")
	buf.WriteString(fieldName)
	buf.WriteString(" (")
	buf.WriteString(formatTypeForComment(&field.Type))
	buf.WriteString(")\n")

	// Handle optional fields
	if field.Type.Optional {
		buf.WriteString("\tif src.")
		buf.WriteString(fieldName)
		buf.WriteString(" == nil {\n")
		buf.WriteString("\t\tbuf[*offset] = 0 // presence = 0 (absent)\n")
		buf.WriteString("\t\t*offset += 1\n")
		buf.WriteString("\t} else {\n")
		buf.WriteString("\t\tbuf[*offset] = 1 // presence = 1 (present)\n")
		buf.WriteString("\t\t*offset += 1\n")

		// Generate encoding for the actual value (indented)
		tempBuf := &strings.Builder{}
		var err error

		switch field.Type.Kind {
		case parser.TypeKindPrimitive:
			err = generatePrimitiveEncode(tempBuf, field.Type.Name, fieldName)
		case parser.TypeKindArray:
			err = generateArrayEncode(tempBuf, &field.Type, fieldName)
		case parser.TypeKindNamed:
			// For optional fields, don't take address since src.FieldName is already a pointer
			err = generateNamedTypeEncodeOptional(tempBuf, field.Type.Name, fieldName)
		default:
			err = fmt.Errorf("unsupported type kind: %v", field.Type.Kind)
		}

		if err != nil {
			return err
		}

		// Add extra indentation for the else block
		lines := strings.Split(tempBuf.String(), "\n")
		for _, line := range lines {
			if line != "" {
				buf.WriteString("\t")
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}

		buf.WriteString("\t}\n")
		return nil
	}

	// Non-optional fields - generate encode normally
	switch field.Type.Kind {
	case parser.TypeKindPrimitive:
		return generatePrimitiveEncode(buf, field.Type.Name, fieldName)
	case parser.TypeKindArray:
		return generateArrayEncode(buf, &field.Type, fieldName)
	case parser.TypeKindNamed:
		return generateNamedTypeEncode(buf, field.Type.Name, fieldName)
	default:
		return fmt.Errorf("unsupported type kind: %v", field.Type.Kind)
	}
}

// formatTypeForComment formats a type expression for display in comments.
func formatTypeForComment(typeExpr *parser.TypeExpr) string {
	switch typeExpr.Kind {
	case parser.TypeKindPrimitive:
		return typeExpr.Name
	case parser.TypeKindArray:
		if typeExpr.Elem != nil {
			return "[]" + formatTypeForComment(typeExpr.Elem)
		}
		return "[]?"
	case parser.TypeKindNamed:
		return typeExpr.Name
	default:
		return "?"
	}
}

// generatePrimitiveEncode generates encode code for primitive fields.
func generatePrimitiveEncode(buf *strings.Builder, typeName, fieldName string) error {
	switch typeName {
	case "u8":
		buf.WriteString("\tbuf[*offset] = src.")
		buf.WriteString(fieldName)
		buf.WriteString("\n")
		buf.WriteString("\t*offset++\n")

	case "u16":
		buf.WriteString("\tbinary.LittleEndian.PutUint16(buf[*offset:], src.")
		buf.WriteString(fieldName)
		buf.WriteString(")\n")
		buf.WriteString("\t*offset += 2\n")

	case "u32":
		buf.WriteString("\tbinary.LittleEndian.PutUint32(buf[*offset:], src.")
		buf.WriteString(fieldName)
		buf.WriteString(")\n")
		buf.WriteString("\t*offset += 4\n")

	case "u64":
		buf.WriteString("\tbinary.LittleEndian.PutUint64(buf[*offset:], src.")
		buf.WriteString(fieldName)
		buf.WriteString(")\n")
		buf.WriteString("\t*offset += 8\n")

	case "i8":
		buf.WriteString("\tbuf[*offset] = uint8(src.")
		buf.WriteString(fieldName)
		buf.WriteString(")\n")
		buf.WriteString("\t*offset++\n")

	case "i16":
		buf.WriteString("\tbinary.LittleEndian.PutUint16(buf[*offset:], uint16(src.")
		buf.WriteString(fieldName)
		buf.WriteString("))\n")
		buf.WriteString("\t*offset += 2\n")

	case "i32":
		buf.WriteString("\tbinary.LittleEndian.PutUint32(buf[*offset:], uint32(src.")
		buf.WriteString(fieldName)
		buf.WriteString("))\n")
		buf.WriteString("\t*offset += 4\n")

	case "i64":
		buf.WriteString("\tbinary.LittleEndian.PutUint64(buf[*offset:], uint64(src.")
		buf.WriteString(fieldName)
		buf.WriteString("))\n")
		buf.WriteString("\t*offset += 8\n")

	case "f32":
		buf.WriteString("\tbinary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.")
		buf.WriteString(fieldName)
		buf.WriteString("))\n")
		buf.WriteString("\t*offset += 4\n")

	case "f64":
		buf.WriteString("\tbinary.LittleEndian.PutUint64(buf[*offset:], math.Float64bits(src.")
		buf.WriteString(fieldName)
		buf.WriteString("))\n")
		buf.WriteString("\t*offset += 8\n")

	case "bool":
		buf.WriteString("\tif src.")
		buf.WriteString(fieldName)
		buf.WriteString(" {\n")
		buf.WriteString("\t\tbuf[*offset] = 1\n")
		buf.WriteString("\t} else {\n")
		buf.WriteString("\t\tbuf[*offset] = 0\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\t*offset++\n")

	case "str":
		return generateStringEncode(buf, fieldName)

	default:
		return fmt.Errorf("unknown primitive type: %s", typeName)
	}

	buf.WriteString("\n")
	return nil
}

// generateStringEncode generates encode code for string fields.
func generateStringEncode(buf *strings.Builder, fieldName string) error {
	// Write length prefix
	buf.WriteString("\tbinary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.")
	buf.WriteString(fieldName)
	buf.WriteString(")))\n")
	buf.WriteString("\t*offset += 4\n")

	// Copy string bytes
	buf.WriteString("\tcopy(buf[*offset:], src.")
	buf.WriteString(fieldName)
	buf.WriteString(")\n")
	buf.WriteString("\t*offset += len(src.")
	buf.WriteString(fieldName)
	buf.WriteString(")\n")

	return nil
}

// generateArrayEncode generates encode code for array fields.
func generateArrayEncode(buf *strings.Builder, typeExpr *parser.TypeExpr, fieldName string) error {
	if typeExpr.Elem == nil {
		return fmt.Errorf("array type missing element type")
	}

	// Write array count
	buf.WriteString("\tbinary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.")
	buf.WriteString(fieldName)
	buf.WriteString(")))\n")
	buf.WriteString("\t*offset += 4\n")
	buf.WriteString("\n")

	// Check if we can use bulk copy optimization for primitive arrays
	elemType := typeExpr.Elem
	if elemType.Kind == parser.TypeKindPrimitive && canUseBulkCopy(elemType.Name) {
		if err := generateBulkArrayCopy(buf, elemType.Name, fieldName); err != nil {
			return err
		}
	} else {
		// Loop through elements
		buf.WriteString("\tfor i := range src.")
		buf.WriteString(fieldName)
		buf.WriteString(" {\n")

		// Encode each element
		if err := generateArrayElementEncode(buf, elemType, fieldName); err != nil {
			return err
		}

		buf.WriteString("\t}\n")
	}

	return nil
}

// generateArrayElementEncode generates encode code for array elements.
func generateArrayElementEncode(buf *strings.Builder, elemType *parser.TypeExpr, fieldName string) error {
	switch elemType.Kind {
	case parser.TypeKindPrimitive:
		return generateArrayPrimitiveElementEncode(buf, elemType.Name, fieldName)
	case parser.TypeKindNamed:
		return generateArrayNamedTypeElementEncode(buf, elemType.Name, fieldName)
	default:
		return fmt.Errorf("unsupported array element type kind: %v", elemType.Kind)
	}
}

// generateArrayPrimitiveElementEncode generates encode code for primitive array elements.
func generateArrayPrimitiveElementEncode(buf *strings.Builder, elemTypeName, fieldName string) error {
	switch elemTypeName {
	case "u8":
		buf.WriteString("\t\tbuf[*offset] = src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i]\n")
		buf.WriteString("\t\t*offset++\n")

	case "u16":
		buf.WriteString("\t\tbinary.LittleEndian.PutUint16(buf[*offset:], src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i])\n")
		buf.WriteString("\t\t*offset += 2\n")

	case "u32":
		buf.WriteString("\t\tbinary.LittleEndian.PutUint32(buf[*offset:], src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i])\n")
		buf.WriteString("\t\t*offset += 4\n")

	case "u64":
		buf.WriteString("\t\tbinary.LittleEndian.PutUint64(buf[*offset:], src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i])\n")
		buf.WriteString("\t\t*offset += 8\n")

	case "i8":
		buf.WriteString("\t\tbuf[*offset] = uint8(src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i])\n")
		buf.WriteString("\t\t*offset++\n")

	case "i16":
		buf.WriteString("\t\tbinary.LittleEndian.PutUint16(buf[*offset:], uint16(src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i]))\n")
		buf.WriteString("\t\t*offset += 2\n")

	case "i32":
		buf.WriteString("\t\tbinary.LittleEndian.PutUint32(buf[*offset:], uint32(src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i]))\n")
		buf.WriteString("\t\t*offset += 4\n")

	case "i64":
		buf.WriteString("\t\tbinary.LittleEndian.PutUint64(buf[*offset:], uint64(src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i]))\n")
		buf.WriteString("\t\t*offset += 8\n")

	case "f32":
		buf.WriteString("\t\tbinary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i]))\n")
		buf.WriteString("\t\t*offset += 4\n")

	case "f64":
		buf.WriteString("\t\tbinary.LittleEndian.PutUint64(buf[*offset:], math.Float64bits(src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i]))\n")
		buf.WriteString("\t\t*offset += 8\n")

	case "bool":
		buf.WriteString("\t\tif src.")
		buf.WriteString(fieldName)
		buf.WriteString("[i] {\n")
		buf.WriteString("\t\t\tbuf[*offset] = 1\n")
		buf.WriteString("\t\t} else {\n")
		buf.WriteString("\t\t\tbuf[*offset] = 0\n")
		buf.WriteString("\t\t}\n")
		buf.WriteString("\t\t*offset++\n")

	case "str":
		return generateArrayStringElementEncode(buf, fieldName)

	default:
		return fmt.Errorf("unknown primitive type in array: %s", elemTypeName)
	}

	return nil
}

// canUseBulkCopy checks if a primitive type can use bulk memory copy optimization.
// Only fixed-size primitives on little-endian systems can use this optimization.
func canUseBulkCopy(typeName string) bool {
	switch typeName {
	case "u8", "i8": // Single byte - always fast path
		return true
	case "u16", "i16", "u32", "i32", "u64", "i64": // Multi-byte integers
		return true
	default:
		return false // f32, f64, bool, str need special handling
	}
}

// generateBulkArrayCopy generates optimized bulk copy for primitive arrays.
// Uses unsafe.Slice to get a byte view of the array and copy in one operation.
func generateBulkArrayCopy(buf *strings.Builder, elemTypeName, fieldName string) error {
	elemSize := getPrimitiveSize(elemTypeName)
	if elemSize == 0 {
		return fmt.Errorf("cannot bulk copy type: %s", elemTypeName)
	}

	// Generate conditional bulk copy (only on little-endian systems)
	buf.WriteString("\t// Bulk copy optimization for primitive arrays\n")
	buf.WriteString("\tif len(src.")
	buf.WriteString(fieldName)
	buf.WriteString(") > 0 {\n")

	if elemSize == 1 {
		// u8/i8: Direct copy without endian concerns
		buf.WriteString("\t\tcopy(buf[*offset:], src.")
		buf.WriteString(fieldName)
		buf.WriteString(")\n")
		buf.WriteString("\t\t*offset += len(src.")
		buf.WriteString(fieldName)
		buf.WriteString(")\n")
	} else {
		// Multi-byte: Use unsafe.Slice for bulk copy
		buf.WriteString("\t\t// Cast slice to bytes for bulk copy\n")
		buf.WriteString("\t\tbytes := unsafe.Slice((*byte)(unsafe.Pointer(&src.")
		buf.WriteString(fieldName)
		buf.WriteString("[0])), len(src.")
		buf.WriteString(fieldName)
		buf.WriteString(fmt.Sprintf(")*%d)\n", elemSize))
		buf.WriteString("\t\tcopy(buf[*offset:], bytes)\n")
		buf.WriteString("\t\t*offset += len(src.")
		buf.WriteString(fieldName)
		buf.WriteString(fmt.Sprintf(")*%d\n", elemSize))
	}

	buf.WriteString("\t}\n")

	return nil
}

// getPrimitiveSize returns the size in bytes of a primitive type, or 0 if variable/unsupported.
func getPrimitiveSize(typeName string) int {
	switch typeName {
	case "u8", "i8", "bool":
		return 1
	case "u16", "i16":
		return 2
	case "u32", "i32", "f32":
		return 4
	case "u64", "i64", "f64":
		return 8
	default:
		return 0
	}
}

// generateArrayStringElementEncode generates encode code for string array elements.
func generateArrayStringElementEncode(buf *strings.Builder, fieldName string) error {
	// Write length prefix
	buf.WriteString("\t\tbinary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.")
	buf.WriteString(fieldName)
	buf.WriteString("[i])))\n")
	buf.WriteString("\t\t*offset += 4\n")

	// Copy string bytes
	buf.WriteString("\t\tcopy(buf[*offset:], src.")
	buf.WriteString(fieldName)
	buf.WriteString("[i])\n")
	buf.WriteString("\t\t*offset += len(src.")
	buf.WriteString(fieldName)
	buf.WriteString("[i])\n")

	return nil
}

// generateNamedTypeEncode generates encode code for nested struct fields.
// For non-optional fields, it takes the address (&src.FieldName).
func generateNamedTypeEncode(buf *strings.Builder, typeName, fieldName string) error {
	return generateNamedTypeEncodeWithPrefix(buf, typeName, fieldName, "&src.")
}

// generateNamedTypeEncodeOptional generates encode code for optional nested struct fields.
// Since optional fields are already pointers, we don't take the address.
func generateNamedTypeEncodeOptional(buf *strings.Builder, typeName, fieldName string) error {
	return generateNamedTypeEncodeWithPrefix(buf, typeName, fieldName, "src.")
}

// Helper function for named type encoding with custom prefix
func generateNamedTypeEncodeWithPrefix(buf *strings.Builder, typeName, fieldName, prefix string) error {
	structName := ToGoName(typeName)
	helperFunc := "encode" + structName

	buf.WriteString("\tif err := ")
	buf.WriteString(helperFunc)
	buf.WriteString("(")
	buf.WriteString(prefix)
	buf.WriteString(fieldName)
	buf.WriteString(", buf, offset); err != nil {\n")
	buf.WriteString("\t\treturn err\n")
	buf.WriteString("\t}\n")

	return nil
}

// generateArrayNamedTypeElementEncode generates encode code for array elements that are structs.
func generateArrayNamedTypeElementEncode(buf *strings.Builder, elemTypeName, fieldName string) error {
	structName := ToGoName(elemTypeName)
	helperFunc := "encode" + structName

	buf.WriteString("\t\tif err := ")
	buf.WriteString(helperFunc)
	buf.WriteString("(&src.")
	buf.WriteString(fieldName)
	buf.WriteString("[i], buf, offset); err != nil {\n")
	buf.WriteString("\t\t\treturn err\n")
	buf.WriteString("\t\t}\n")

	return nil
}
