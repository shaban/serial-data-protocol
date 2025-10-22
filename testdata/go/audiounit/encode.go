package audiounit

import (
	"io"
	"encoding/binary"
	"math"
)

// calculateParameterSize calculates the wire format size for Parameter.
func calculateParameterSize(src *Parameter) int {
	size := 0
	// Field: Address
	size += 8
	// Field: DisplayName
	size += 4 + len(src.DisplayName)
	// Field: Identifier
	size += 4 + len(src.Identifier)
	// Field: Unit
	size += 4 + len(src.Unit)
	// Field: MinValue
	size += 4
	// Field: MaxValue
	size += 4
	// Field: DefaultValue
	size += 4
	// Field: CurrentValue
	size += 4
	// Field: RawFlags
	size += 4
	// Field: IsWritable
	size += 1
	// Field: CanRamp
	size += 1
	return size
}

// EncodeParameter encodes a Parameter to wire format.
// It returns the encoded bytes or an error.
func EncodeParameter(src *Parameter) ([]byte, error) {
	size := calculateParameterSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeParameter(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculatePluginSize calculates the wire format size for Plugin.
func calculatePluginSize(src *Plugin) int {
	size := 0
	// Field: Name
	size += 4 + len(src.Name)
	// Field: ManufacturerId
	size += 4 + len(src.ManufacturerId)
	// Field: ComponentType
	size += 4 + len(src.ComponentType)
	// Field: ComponentSubtype
	size += 4 + len(src.ComponentSubtype)
	// Field: Parameters
	size += 4
	for i := range src.Parameters {
		size += calculateParameterSize(&src.Parameters[i])
	}
	return size
}

// EncodePlugin encodes a Plugin to wire format.
// It returns the encoded bytes or an error.
func EncodePlugin(src *Plugin) ([]byte, error) {
	size := calculatePluginSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodePlugin(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculatePluginRegistrySize calculates the wire format size for PluginRegistry.
func calculatePluginRegistrySize(src *PluginRegistry) int {
	size := 0
	// Field: Plugins
	size += 4
	for i := range src.Plugins {
		size += calculatePluginSize(&src.Plugins[i])
	}
	// Field: TotalPluginCount
	size += 4
	// Field: TotalParameterCount
	size += 4
	return size
}

// EncodePluginRegistry encodes a PluginRegistry to wire format.
// It returns the encoded bytes or an error.
func EncodePluginRegistry(src *PluginRegistry) ([]byte, error) {
	size := calculatePluginRegistrySize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodePluginRegistry(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}


// encodeParameter is the helper function that encodes Parameter fields.
func encodeParameter(src *Parameter, buf []byte, offset *int) error {
	// Field: Address (u64)
	binary.LittleEndian.PutUint64(buf[*offset:], src.Address)
	*offset += 8

	// Field: DisplayName (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.DisplayName)))
	*offset += 4
	copy(buf[*offset:], src.DisplayName)
	*offset += len(src.DisplayName)
	// Field: Identifier (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Identifier)))
	*offset += 4
	copy(buf[*offset:], src.Identifier)
	*offset += len(src.Identifier)
	// Field: Unit (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Unit)))
	*offset += 4
	copy(buf[*offset:], src.Unit)
	*offset += len(src.Unit)
	// Field: MinValue (f32)
	binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.MinValue))
	*offset += 4

	// Field: MaxValue (f32)
	binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.MaxValue))
	*offset += 4

	// Field: DefaultValue (f32)
	binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.DefaultValue))
	*offset += 4

	// Field: CurrentValue (f32)
	binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.CurrentValue))
	*offset += 4

	// Field: RawFlags (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.RawFlags)
	*offset += 4

	// Field: IsWritable (bool)
	if src.IsWritable {
		buf[*offset] = 1
	} else {
		buf[*offset] = 0
	}
	*offset++

	// Field: CanRamp (bool)
	if src.CanRamp {
		buf[*offset] = 1
	} else {
		buf[*offset] = 0
	}
	*offset++

	return nil
}

// encodePlugin is the helper function that encodes Plugin fields.
func encodePlugin(src *Plugin, buf []byte, offset *int) error {
	// Field: Name (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Name)))
	*offset += 4
	copy(buf[*offset:], src.Name)
	*offset += len(src.Name)
	// Field: ManufacturerId (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.ManufacturerId)))
	*offset += 4
	copy(buf[*offset:], src.ManufacturerId)
	*offset += len(src.ManufacturerId)
	// Field: ComponentType (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.ComponentType)))
	*offset += 4
	copy(buf[*offset:], src.ComponentType)
	*offset += len(src.ComponentType)
	// Field: ComponentSubtype (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.ComponentSubtype)))
	*offset += 4
	copy(buf[*offset:], src.ComponentSubtype)
	*offset += len(src.ComponentSubtype)
	// Field: Parameters ([]Parameter)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Parameters)))
	*offset += 4

	for i := range src.Parameters {
		if err := encodeParameter(&src.Parameters[i], buf, offset); err != nil {
			return err
		}
	}
	return nil
}

// encodePluginRegistry is the helper function that encodes PluginRegistry fields.
func encodePluginRegistry(src *PluginRegistry, buf []byte, offset *int) error {
	// Field: Plugins ([]Plugin)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Plugins)))
	*offset += 4

	for i := range src.Plugins {
		if err := encodePlugin(&src.Plugins[i], buf, offset); err != nil {
			return err
		}
	}
	// Field: TotalPluginCount (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.TotalPluginCount)
	*offset += 4

	// Field: TotalParameterCount (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.TotalParameterCount)
	*offset += 4

	return nil
}


// EncodeParameterMessage encodes a Parameter to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeParameterMessage(src *Parameter) ([]byte, error) {
	// Encode payload
	payload, err := EncodeParameter(src)
	if err != nil {
		return nil, err
	}

	// Allocate message buffer (header + payload)
	messageSize := MessageHeaderSize + len(payload)
	message := make([]byte, messageSize)

	// Write header
	copy(message[0:3], MessageMagic)  // Magic bytes 'SDP'
	message[3] = MessageVersion       // Protocol version '2'
	binary.LittleEndian.PutUint16(message[4:6], 1)  // Type ID
	binary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length

	// Copy payload
	copy(message[10:], payload)

	return message, nil
}

// EncodePluginMessage encodes a Plugin to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodePluginMessage(src *Plugin) ([]byte, error) {
	// Encode payload
	payload, err := EncodePlugin(src)
	if err != nil {
		return nil, err
	}

	// Allocate message buffer (header + payload)
	messageSize := MessageHeaderSize + len(payload)
	message := make([]byte, messageSize)

	// Write header
	copy(message[0:3], MessageMagic)  // Magic bytes 'SDP'
	message[3] = MessageVersion       // Protocol version '2'
	binary.LittleEndian.PutUint16(message[4:6], 2)  // Type ID
	binary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length

	// Copy payload
	copy(message[10:], payload)

	return message, nil
}

// EncodePluginRegistryMessage encodes a PluginRegistry to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodePluginRegistryMessage(src *PluginRegistry) ([]byte, error) {
	// Encode payload
	payload, err := EncodePluginRegistry(src)
	if err != nil {
		return nil, err
	}

	// Allocate message buffer (header + payload)
	messageSize := MessageHeaderSize + len(payload)
	message := make([]byte, messageSize)

	// Write header
	copy(message[0:3], MessageMagic)  // Magic bytes 'SDP'
	message[3] = MessageVersion       // Protocol version '2'
	binary.LittleEndian.PutUint16(message[4:6], 3)  // Type ID
	binary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length

	// Copy payload
	copy(message[10:], payload)

	return message, nil
}



// EncodeParameterToWriter encodes a Parameter to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeParameterToWriter(src *Parameter, w io.Writer) error {
	size := calculateParameterSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeParameter(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodePluginToWriter encodes a Plugin to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodePluginToWriter(src *Plugin, w io.Writer) error {
	size := calculatePluginSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodePlugin(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodePluginRegistryToWriter encodes a PluginRegistry to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodePluginRegistryToWriter(src *PluginRegistry, w io.Writer) error {
	size := calculatePluginRegistrySize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodePluginRegistry(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}
