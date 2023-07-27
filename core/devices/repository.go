package devices

type Repository interface {
	Store(deviceName string, Device *Device)
	ListAllDevices() []*Device
	FindDevice(deviceName string) (*Device, error)
}
