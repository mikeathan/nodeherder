package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/models/devices"
	"sync"
)

// examples
// sensor name = condition 1 & condiiton 1  && condtion n.... = action
// presence = (true) && (lux <= 30) = turn on
// presence = (false) && timer condition = turn off
// presence = true = turn on
// presence = false = turn off

var contextIgnoreList = []string{"action"}

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

// Device Context
type DeviceContext struct {
	currentData map[string]any
	payload     map[string]*devices.Entity
	pendingData map[string]any
	mu          sync.RWMutex
}

func NewDeviceContext() *DeviceContext {
	return &DeviceContext{
		currentData: map[string]any{},
		payload:     make(map[string]*devices.Entity),
		pendingData: map[string]any{},
	}
}

func (d *DeviceContext) SetPayload(payload map[string]*devices.Entity) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.payload = payload
}

// wIP
func (d *DeviceContext) SetPending(name string, value any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pendingData[name] = value
}

func (d *DeviceContext) GetPending(name string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.pendingData[name]
}
// wIP


func (d *DeviceContext) GetPayload(name string) (*devices.Entity, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	value, exists := d.payload[name]
	return value, exists
}

func (d *DeviceContext) GetCurrent(name string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentData[name]
}

func (d *DeviceContext) SetCurrent(name string, value any) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// if trigger is in ignore list, we  want to trigger it again
	for _, item := range contextIgnoreList {
		if item == name {
			return
		}
	}

	d.currentData[name] = value
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
	d.ctx.SetPayload(device.Exposes)

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
	d.ctx.SetPayload(device.Exposes) // rename to set currentData 

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
