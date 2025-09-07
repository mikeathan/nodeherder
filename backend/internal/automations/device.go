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

type DeviceContext struct {
	currentData map[string]any
	payload     map[string]*devices.Entity
	mu          sync.RWMutex
}

func NewDeviceContext() *DeviceContext {
	return &DeviceContext{
		currentData: map[string]any{},
		payload:     make(map[string]*devices.Entity),
	}
}

func (d *DeviceContext) SetPayload(payload map[string]*devices.Entity) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.payload = payload
}

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

will have here a custom automation, not linked to a Device. to start it can be like a Button click or sth or web hook ?
type Device struct {
	Id           string          `json:"id"`
	FriendlyName string          `json:"friendlyname"`
	Description  string          `json:"description"`
	Enabled      bool            `json:"enabled"`
	Triggers     []*Trigger      `json:"triggers"`
	Schedules    []*TimeSchedule `json:"schedules"`
	ctx          *DeviceContext
}

func newDevice() *Device {
	return NewDevice("")
}

func NewDevice(id string) *Device {

	d := &Device{
		Id:           id,
		FriendlyName: "",
		Description:  "",
		Enabled:      false,
		Triggers:     []*Trigger{},
		Schedules:    []*TimeSchedule{},
		ctx:          NewDeviceContext(),
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
	d.FriendlyName = aux.FriendlyName
	d.Description = aux.Description
	d.Enabled = aux.Enabled
	d.Schedules = aux.Schedules
	d.Triggers = aux.Triggers
	return nil
}

func (d *Device) Evaluate(device *devices.Device) bool {

	d.ctx.SetPayload(device.Exposes)

	// NOTE: a trigger can have multiple conditions.
	// e.g presence can have multiple conditions for on and off
	for _, trigger := range d.Triggers {
		if _, ok := device.Exposes[trigger.Name]; ok {
			trigger.process(d.ctx)
		}
	}
	return false
}

// TODO:
// THIS CAN BE AUTOMATION HANDLE
func (d *Device) configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

	//  check if device with automation id exists. friendyname can change
	bridgeInfo, err := registrar.FindBridgeInfo(d.Id)
	if err != nil {
		return err
	}

	if bridgeInfo.Disabled {
		return fmt.Errorf("device %s is disabled ", bridgeInfo.FriendlyName)
	}

	// validate conditions
	for _, trigger := range d.Triggers {

		// validate actions
		for _, action := range trigger.Actions {
			err := action.Configure(registrar, client)
			if err != nil {
				return err
			}
		}
	}

	d.FriendlyName = bridgeInfo.FriendlyName

	return nil
}
