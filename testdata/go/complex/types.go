package complex

type Parameter struct {
	Id uint32
	Name string
	Value float32
	Min float32
	Max float32
}

type Plugin struct {
	Id uint32
	Name string
	Manufacturer string
	Version uint32
	Enabled bool
	Parameters []Parameter
}

type AudioDevice struct {
	DeviceId uint32
	DeviceName string
	SampleRate uint32
	BufferSize uint32
	InputChannels uint16
	OutputChannels uint16
	IsDefault bool
	ActivePlugins []Plugin
}
