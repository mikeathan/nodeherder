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
	manual  bool
}

// NewDeviceEvent is the event for a device message with its changed payload.
func NewDeviceEvent(device *devices.Device, payload map[string]interface{}) *DeviceEvent {
	return &DeviceEvent{device: device, payload: payload}
}

// NewManualEvent is the event for a manual trigger of device's automation.
func NewManualEvent(device *devices.Device) *DeviceEvent {
	return &DeviceEvent{device: device, manual: true}
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

// IsManual reports whether the event is a manual trigger rather than a device message.
func (de *DeviceEvent) IsManual() bool {
	return de.manual
}

// runContext isolates one run's exposes and origin over the persistent state.
func (de *DeviceEvent) runContext(state *DeviceContext) AutomationContext {
	var exposes map[string]*devices.Entity
	if de.device != nil {
		exposes = de.device.Exposes
	}
	return NewRunContext(state, exposes, de.manual)
}

// Device Automation

// ErrAutomationSourceDisabled distinguishes an explicit source disable from a
// transient configuration failure, which may retain the last working generation.
var ErrAutomationSourceDisabled = errors.New("automation source is disabled")

type Device struct {
	BaseAutomation
	state  *DeviceContext
	inputs *triggerDependencies
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
		state: NewDeviceContext(opts...),
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

	if d.state == nil {
		d.state = NewDeviceContext()
	}
	return nil
}

// Device's only serialized fields are its base; ctx remains private.
func (d *Device) MarshalJSON() ([]byte, error) {
	return d.BaseAutomation.MarshalJSON()
}

// Evaluate selects triggers for one device event and starts actions only for a
// primary-property change. A condition-dependency-only change evaluates conditions
// and cancels pending actions, but never commands a device. It returns false for
// non-device events and when an action fails. Manual runs bypass selection through
// EvaluateTrigger.
func (d *Device) Evaluate(event TriggerEvent) bool {
	deviceEvent, ok := event.(*DeviceEvent)
	if !ok {
		return false
	}

	ctx := deviceEvent.runContext(d.state)
	payload := deviceEvent.Payload()
	success := true
	for _, trigger := range d.Triggers {
		if !trigger.AffectedBy(ctx, payload) {
			continue
		}
		// Starting actions on a dependency-only change re-creates the device-control
		// loop documented in specs/068-automation-feedback-loop-guard.
		if _, primary := payload[trigger.GetName()]; primary {
			if err := trigger.Process(ctx); err != nil {
				utils.LogErrorf("automation %s: %v", d.Id, err)
				success = false
			}
			continue
		}
		if dependent, ok := trigger.(dependencyOnlyTrigger); ok {
			dependent.processDependencyChange(ctx)
		}
	}
	return success
}

// EvaluateTrigger runs every trigger with the given name directly for a manual
// event, bypassing device selection. It joins action failures and reports an error
// when the event is not a device event or no trigger matches.
func (d *Device) EvaluateTrigger(event TriggerEvent, triggerName string) error {

	deviceEvent, ok := event.(*DeviceEvent)
	if !ok {
		return fmt.Errorf("manual trigger requires a device event")
	}

	ctx := deviceEvent.runContext(d.state)

	matched := false
	var failures error
	for _, trigger := range d.Triggers {
		if trigger.GetName() == triggerName {
			matched = true
			failures = errors.Join(failures, trigger.Process(ctx))
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

	if err := d.BaseAutomation.Configure(registrar, client); err != nil {
		return err
	}
	d.inputs = newTriggerDependencies(d.Triggers)
	return nil
}

// AcceptsDeviceUpdate avoids constructing a run context for unrelated telemetry.
// Unconfigured instances intentionally keep the conservative full-selection path.
func (d *Device) AcceptsDeviceUpdate(changes map[string]interface{}) bool {
	return d.inputs.accepts(changes)
}
