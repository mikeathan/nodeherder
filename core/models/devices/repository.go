package devices

type Repository interface {
	Store(deviceName string, Device *Device)
	ListAllDevices() []*Device
	FindDevice(deviceName string) (*Device, error)

	ListAllDevicesV2() []*DeviceV2
	StoreV2(friendlyName string, device *DeviceV2)
	FindDeviceV2(friendlyName string) (*DeviceV2, error)
	RegisterBridge(bridgeInfoList []*BridgeInfo)
	FindBridgeInfo(id string) *BridgeInfo
	ResolveId(friendlyName string) string
}
