package devices

type DeviceRemoveRequest struct {
	ID    string `json:"id"`
	Force bool   `json:"force,omitempty"`
}

func NewDeviceRemoveRequest(id string, force bool) *DeviceRemoveRequest {
	return &DeviceRemoveRequest{
		ID:    id,
		Force: force,
	}
}

type BridgePerminJoinRequest struct {
	Value bool `json:"value"`
	Time  int  `json:"time"`
}

func NewBridgePerminJoinRequest(value bool, time int) *BridgePerminJoinRequest {
	return &BridgePerminJoinRequest{
		Value: value,
		Time:  time,
	}
}
