package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/models/devices"
)

// examples
// sensor name = condition 1 & condiiton 1  && condtion n.... = action
// presence = (true) && (lux <= 30) = turn on
// presence = (false) && timer condition = turn off
// presence = true = turn on
// presence = false = turn off

type DeviceEvent struct {
	TriggerEvent
	device *devices.Device
}

func NewDeviceEvent(device *devices.Device) *DeviceEvent {
	return &DeviceEvent{device: device}
}

func (de *DeviceEvent) Device() *devices.Device {
	return de.device
}

func (de *DeviceEvent) Type() string {
	return "device"
}

// Device Automation

type Device struct {
	BaseAutomation
	ctx AutomationContext
}

func newDevice() *Device {
	return NewDevice("")
}

func NewDevice(id string) *Device {

	d := &Device{
		BaseAutomation: BaseAutomation{
			Id:           id,
			Type:         DeviceAutomationType,
			FriendlyName: "",
			Description:  "",
			Enabled:      false,
			Triggers:     TriggerList{},
			Schedules:    []*TimeSchedule{},
		},
		ctx: NewDeviceContext(),
	}

	return d
}

func CreateFromPayload(payload []byte) (*Device, error) {

	device := newDevice()
	err := json.Unmarshal(payload, device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (d *Device) UnmarshalJSON(data []byte) error {
	type Alias Device
	aux := &Alias{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	d.Id = aux.Id
	d.Type = aux.Type
	d.FriendlyName = aux.FriendlyName
	d.Description = aux.Description
	d.Enabled = aux.Enabled
	d.Schedules = aux.Schedules
	d.Triggers = aux.Triggers

	if d.ctx == nil {
		d.ctx = NewDeviceContext()
	}
	return nil
}

func (d *Device) Evaluate(event TriggerEvent) bool {
	// only handle device events
	deviceEvent, ok := event.(*DeviceEvent)
	if !ok {
		return false
	}

	device := deviceEvent.Device()
	d.ctx.SetDevicePayload(device.Exposes)

	// NOTE: a trigger can have multiple conditions.
	// e.g presence can have multiple conditions for on and off
	for _, trigger := range d.Triggers {
		if _, ok := device.Exposes[trigger.GetName()]; ok {
			trigger.Process(d.ctx)
		}
	}
	return true
}

func (d *Device) EvaluateTrigger(event TriggerEvent, triggerName string) bool {

	deviceEvent, ok := event.(*DeviceEvent)
	if !ok {
		return false
	}

	device := deviceEvent.Device()

	// TODO: can pass the Device event directly
	// payload is the current device expose
	d.ctx.SetDevicePayload(device.Exposes) // rename to set currentData

	for _, trigger := range d.Triggers {
		if trigger.GetName() == triggerName {
			trigger.Process(d.ctx)
		}
	}

	return true
}

func (d *Device) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

	//  check if device with automation id exists. friendyname can change
	bridgeInfo, err := registrar.FindBridgeInfo(d.Id)
	if err != nil {
		return err
	}

	if bridgeInfo.Disabled {
		return fmt.Errorf("device %s is disabled ", bridgeInfo.FriendlyName)
	}

	d.FriendlyName = bridgeInfo.FriendlyName

	return d.BaseAutomation.Configure(registrar, client)
}
