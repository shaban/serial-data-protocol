package complex

import (
	"io"
	"encoding/binary"
	"math"
)

// calculateParameterSize calculates the wire format size for Parameter.
func calculateParameterSize(src *Parameter) int {
	size := 0
	// Field: Id
	size += 4
	// Field: Name
	size += 4 + len(src.Name)
	// Field: Value
	size += 4
	// Field: Min
	size += 4
	// Field: Max
	size += 4
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
	// Field: Id
	size += 4
	// Field: Name
	size += 4 + len(src.Name)
	// Field: Manufacturer
	size += 4 + len(src.Manufacturer)
	// Field: Version
	size += 4
	// Field: Enabled
	size += 1
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

// calculateAudioDeviceSize calculates the wire format size for AudioDevice.
func calculateAudioDeviceSize(src *AudioDevice) int {
	size := 0
	// Field: DeviceId
	size += 4
	// Field: DeviceName
	size += 4 + len(src.DeviceName)
	// Field: SampleRate
	size += 4
	// Field: BufferSize
	size += 4
	// Field: InputChannels
	size += 2
	// Field: OutputChannels
	size += 2
	// Field: IsDefault
	size += 1
	// Field: ActivePlugins
	size += 4
	for i := range src.ActivePlugins {
		size += calculatePluginSize(&src.ActivePlugins[i])
	}
	return size
}

// EncodeAudioDevice encodes a AudioDevice to wire format.
// It returns the encoded bytes or an error.
func EncodeAudioDevice(src *AudioDevice) ([]byte, error) {
	size := calculateAudioDeviceSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeAudioDevice(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}


// encodeParameter is the helper function that encodes Parameter fields.
func encodeParameter(src *Parameter, buf []byte, offset *int) error {
	// Field: Id (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.Id)
	*offset += 4

	// Field: Name (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Name)))
	*offset += 4
	copy(buf[*offset:], src.Name)
	*offset += len(src.Name)
	// Field: Value (f32)
	binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.Value))
	*offset += 4

	// Field: Min (f32)
	binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.Min))
	*offset += 4

	// Field: Max (f32)
	binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.Max))
	*offset += 4

	return nil
}

// encodePlugin is the helper function that encodes Plugin fields.
func encodePlugin(src *Plugin, buf []byte, offset *int) error {
	// Field: Id (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.Id)
	*offset += 4

	// Field: Name (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Name)))
	*offset += 4
	copy(buf[*offset:], src.Name)
	*offset += len(src.Name)
	// Field: Manufacturer (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Manufacturer)))
	*offset += 4
	copy(buf[*offset:], src.Manufacturer)
	*offset += len(src.Manufacturer)
	// Field: Version (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.Version)
	*offset += 4

	// Field: Enabled (bool)
	if src.Enabled {
		buf[*offset] = 1
	} else {
		buf[*offset] = 0
	}
	*offset++

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

// encodeAudioDevice is the helper function that encodes AudioDevice fields.
func encodeAudioDevice(src *AudioDevice, buf []byte, offset *int) error {
	// Field: DeviceId (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.DeviceId)
	*offset += 4

	// Field: DeviceName (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.DeviceName)))
	*offset += 4
	copy(buf[*offset:], src.DeviceName)
	*offset += len(src.DeviceName)
	// Field: SampleRate (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.SampleRate)
	*offset += 4

	// Field: BufferSize (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.BufferSize)
	*offset += 4

	// Field: InputChannels (u16)
	binary.LittleEndian.PutUint16(buf[*offset:], src.InputChannels)
	*offset += 2

	// Field: OutputChannels (u16)
	binary.LittleEndian.PutUint16(buf[*offset:], src.OutputChannels)
	*offset += 2

	// Field: IsDefault (bool)
	if src.IsDefault {
		buf[*offset] = 1
	} else {
		buf[*offset] = 0
	}
	*offset++

	// Field: ActivePlugins ([]Plugin)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.ActivePlugins)))
	*offset += 4

	for i := range src.ActivePlugins {
		if err := encodePlugin(&src.ActivePlugins[i], buf, offset); err != nil {
			return err
		}
	}
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

// EncodeAudioDeviceMessage encodes a AudioDevice to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeAudioDeviceMessage(src *AudioDevice) ([]byte, error) {
	// Encode payload
	payload, err := EncodeAudioDevice(src)
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

// EncodeAudioDeviceToWriter encodes a AudioDevice to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeAudioDeviceToWriter(src *AudioDevice, w io.Writer) error {
	size := calculateAudioDeviceSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeAudioDevice(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}
