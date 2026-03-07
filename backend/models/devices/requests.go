package devices

import "time"

type DashboardGroupRenameRequest struct {
	OldName string `json:"oldName"`
	NewName string `json:"newName"`
}

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
	AvailabilityTimeout         time.Duration
	OnNewDevice                 func(device *Device)
	OnDeviceUpdated             func(device *Device, data *UpdatePackage)
	OnDeviceMeasurementsUpdated func(device *Device, dataMap map[string]interface{})
	OnDeviceAutomationTriggered func(device *Device)
	OnDeviceAvailabilityChanged func(p *UpdatePackage)
}

func NewDeviceRequestEvents() *DeviceRequestEvents {
	return &DeviceRequestEvents{
		AvailabilityTimeout:         time.Duration(24) * time.Hour,
		OnNewDevice:                 nil,
		OnDeviceUpdated:             nil,
		OnDeviceAvailabilityChanged: nil,
	}
}

func (d *DeviceRequestEvents) WithAvailabilityTimeout(timeout time.Duration) *DeviceRequestEvents {
	d.AvailabilityTimeout = timeout
	return d
}

func (d *DeviceRequestEvents) WithOnDeviceAvailabilityChanged(f func(p *UpdatePackage)) *DeviceRequestEvents {
	d.OnDeviceAvailabilityChanged = f
	return d
}

func (d *DeviceRequestEvents) WithOnNewDevice(f func(device *Device)) *DeviceRequestEvents {
	d.OnNewDevice = f
	return d
}

func (d *DeviceRequestEvents) WithOnDeviceUpdated(f func(device *Device, data *UpdatePackage)) *DeviceRequestEvents {
	d.OnDeviceUpdated = f
	return d
}

func (d *DeviceRequestEvents) WithOnDeviceMeasurementsUpdated(f func(device *Device, dataMap map[string]interface{})) *DeviceRequestEvents {
	d.OnDeviceMeasurementsUpdated = f
	return d
}

func (d *DeviceRequestEvents) WithOnDeviceAutomationTriggered(f func(device *Device)) *DeviceRequestEvents {
	d.OnDeviceAutomationTriggered = f
	return d
}
