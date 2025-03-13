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

type DeviceRequestEvents struct {
	AvailabilityTimeout int
	OnNewDevice                 func(device *Device, dataMap map[string]interface{})
	OnDeviceUpdated             func(device *Device, data *UpdatePackage)
	OnDeviceAvailabilityChanged func(dataMap map[string]interface{})
}

func NewDeviceRequestEvents(availabilitytimeout int) *DeviceRequestEvents {
	return &DeviceRequestEvents{
		AvailabilityTimeout: availabilitytimeout,
		OnNewDevice:                 nil,
		OnDeviceUpdated:             nil,
		OnDeviceAvailabilityChanged: nil,
	}
}

func (d *DeviceRequestEvents) WithOnDeviceAvailabilityChanged(f func(dataMap map[string]interface{})) *DeviceRequestEvents {
	d.OnDeviceAvailabilityChanged = f
	return d
}

func (d *DeviceRequestEvents) WithOnNewDevice(f func(device *Device, dataMap map[string]interface{})) *DeviceRequestEvents {
	d.OnNewDevice = f
	return d
}

func (d *DeviceRequestEvents) WithOnDeviceUpdated(f func(device *Device, data *UpdatePackage)) *DeviceRequestEvents {
	d.OnDeviceUpdated = f
	return d
}
