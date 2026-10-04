package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/models/devices"
	"node-herder/utils"
)

// examples
// sensor name = condition 1 & condiiton 1  && condtion n.... = action
// presence = (true) && (lux <= 30) = turn on
// presence = (false) && timer condition = turn off
// presence = true = turn on
// presence = false = turn off

type DeviceEvent struct {
	TriggerEvent
	device  *devices.Device
	payload map[string]interface{}
}

func NewDeviceEvent(device *devices.Device, payload map[string]interface{}) *DeviceEvent {
	return &DeviceEvent{device: device, payload: payload}
}

func (de *DeviceEvent) Device() *devices.Device {
	return de.device
}

func (de *DeviceEvent) Payload() map[string]interface{} {
	return de.payload
}

func (de *DeviceEvent) Type() string {
	return "device"
}

// Device Automation

// ErrAutomationSourceDisabled distinguishes an explicit source disable from a
// transient configuration failure, which may retain the last working generation.
var ErrAutomationSourceDisabled = errors.New("automation source is disabled")

type Device struct {
	BaseAutomation
	ctx AutomationContext
}

func newDevice() *Device {
	return NewDevice("")
}

func NewDevice(id string, opts ...func(*DeviceContext)) *Device {

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
		ctx: NewDeviceContext(opts...),
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
	d.SetEnabled(aux.Enabled)
	d.Schedules = aux.Schedules
	d.Triggers = aux.Triggers

	if d.ctx == nil {
		d.ctx = NewDeviceContext()
	}
	return nil
}

// Device's only serialized fields are its base; ctx remains private.
func (d *Device) MarshalJSON() ([]byte, error) {
	return d.BaseAutomation.MarshalJSON()
}

func (d *Device) Evaluate(event TriggerEvent) bool {
	// only handle device events
	deviceEvent, ok := event.(*DeviceEvent)
	if !ok {
		return false
	}

	device := deviceEvent.Device()
	payload := deviceEvent.Payload()
	d.ctx.SetDevicePayload(device.Exposes)

	// This is a device state change (not manual)
	d.ctx.SetManualTrigger(false)

	// NOTE: a trigger can have multiple conditions.
	// e.g presence can have multiple conditions for on and off
	success := true
	for _, trigger := range d.Triggers {
		// Only process this trigger if its property is in the changed payload
		if _, ok := payload[trigger.GetName()]; ok {
			if err := trigger.Process(d.ctx); err != nil {
				utils.LogErrorf("automation %s: %v", d.Id, err)
				success = false
			}
		}
	}
	return success
}

func (d *Device) EvaluateTrigger(event TriggerEvent, triggerName string) error {

	deviceEvent, ok := event.(*DeviceEvent)
	if !ok {
		return fmt.Errorf("manual trigger requires a device event")
	}

	device := deviceEvent.Device()

	// TODO: can pass the Device event directly
	// payload is the current device expose
	d.ctx.SetDevicePayload(device.Exposes)

	// Set manual trigger flag - this method is called for manual triggers
	d.ctx.SetManualTrigger(true)

	matched := false
	var failures error
	for _, trigger := range d.Triggers {
		if trigger.GetName() == triggerName {
			matched = true
			failures = errors.Join(failures, trigger.Process(d.ctx))
		}
	}

	if !matched {
		return fmt.Errorf("trigger %s not found", triggerName)
	}
	return failures
}

func (d *Device) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

	//  check if device with automation id exists. friendyname can change
	bridgeInfo, err := registrar.FindBridgeInfo(d.Id)
	if err != nil {
		return err
	}

	if bridgeInfo.Disabled {
		return fmt.Errorf("device %s: %w", bridgeInfo.FriendlyName, ErrAutomationSourceDisabled)
	}

	d.FriendlyName = bridgeInfo.FriendlyName

	return d.BaseAutomation.Configure(registrar, client)
}
