package devices

import "github.com/google/uuid"

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

type BridgePermitJoinRequest struct {
	Value       bool   `json:"value"`
	Time        int    `json:"time"`
	Transaction uint32 `json:"transaction"`
}

func NewBridgePermitJoinRequest(value bool, time int) *BridgePermitJoinRequest {
	return &BridgePermitJoinRequest{
		Value:       value,
		Time:        time,
		Transaction: uuid.New().ID(),
	}
}
