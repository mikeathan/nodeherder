package devices

type Repository interface {
	ListAllDevicesV2() []*DeviceV2
	StoreV2(friendlyName string, device *DeviceV2)
	FindDeviceV2(friendlyName string) (*DeviceV2, error)
	FindDeviceV2ById(id string) (*DeviceV2, error)
	FindBridgeInfo(id string) *BridgeInfo
	ResolveId(friendlyName string) string
	GetBridgeFeatures() []*BridgeFeature
}
