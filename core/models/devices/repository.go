package devices

type Repository interface {
	ListAllDevicesV2() []*DeviceV2
	StoreV2(key string, device *DeviceV2)
	FindDeviceV2(key string) (*DeviceV2, error)
}
