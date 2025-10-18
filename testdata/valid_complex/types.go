package valid_complex

// DeviceList A list of devices
type DeviceList struct {
	Devices []Device
}

// Device A device with parameters
type Device struct {
	Id uint32
	Name string
	Parameters []Parameter
}

// Parameter A configuration parameter
type Parameter struct {
	Name string
	Value float64
}
