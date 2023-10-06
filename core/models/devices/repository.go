package devices

type Repository interface {
	Store(deviceName string, Device *Device)
	ListAllDevices() []*Device
	FindDevice(deviceName string) (*Device, error)

	StoreV2(deviceName string, device *DeviceV2)
	FindDeviceV2(deviceName string) (*DeviceV2, error)
}
