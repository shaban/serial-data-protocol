package audiounit

type Parameter struct {
	Address uint64
	DisplayName string
	Identifier string
	Unit string
	MinValue float32
	MaxValue float32
	DefaultValue float32
	CurrentValue float32
	RawFlags uint32
	IsWritable bool
	CanRamp bool
}

type Plugin struct {
	Name string
	ManufacturerId string
	ComponentType string
	ComponentSubtype string
	Parameters []Parameter
}

type PluginRegistry struct {
	Plugins []Plugin
	TotalPluginCount uint32
	TotalParameterCount uint32
}
