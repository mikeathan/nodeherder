package automations

import (
	"encoding/json"
	"errors"
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

var contextIgnoreList = []string{"action"}

type DeviceContext struct {
	currentData map[string]any
	Payload     map[string]*devices.Entity
}

func NewDeviceContext() *DeviceContext {
	return &DeviceContext{
		currentData: map[string]any{},
		Payload:     make(map[string]*devices.Entity),
	}
}

func (d *DeviceContext) GetCurrent(name string) any {
	return d.currentData[name]
}

func (d *DeviceContext) SetCurrent(name string, value any) {

	// if trigger is in ignore list, we  want to trigger it again
	for _, item := range contextIgnoreList {
		if item == name {
			return
		}
	}

	d.currentData[name] = value
}

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
	// err := json.Unmarshal(payload, device)
	// if err != nil {
	// 	return nil, err
	// }

}

func (d *Device) UnmarshalJSON(data []byte) error {
	type Alias Device // Prevent infinite recursion
	aux := &Alias{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	d.Id = aux.Id
	d.FriendlyName = aux.FriendlyName
	d.Description = aux.Description
	d.Enabled = aux.Enabled
	d.Schedules = aux.Schedules

	var rawTriggers []json.RawMessage
	if err := json.Unmarshal(data, &struct {
		Triggers *[]json.RawMessage `json:"triggers"`
	}{Triggers: &rawTriggers}); err != nil {
		return err
	}

	d.Triggers = make([]*Trigger, len(rawTriggers))
	for i, rawTrigger := range rawTriggers {
		trigger := &Trigger{}
		if err := json.Unmarshal(rawTrigger, trigger); err != nil { // Crucial: Use Trigger's UnmarshalJSON
			return fmt.Errorf("unmarshaling trigger %d: %w", i, err)
		}
		d.Triggers[i] = trigger
	}

	return nil
}

func (d *Device) Evaluate(device *devices.Device) bool {

	d.ctx.Payload = device.Exposes

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
			err := configureAction(registrar, action, client)
			if err != nil {
				return err
			}
		}
	}

	d.FriendlyName = bridgeInfo.FriendlyName

	return nil
}

// TODO:
// THIS CAN BE AUTOMATION HANDLE
func configureAction(registrar services.DeviceRegistrar, action *MqttAction, client mqtt.MqttClient) error {

	bridgeInfo, err := registrar.FindBridgeInfo(action.Id)
	if err != nil {
		return err
	}

	for _, e := range bridgeInfo.Definition.Exposes {
		for _, f := range e.Features {

			if action.Property == f.Property {

				sanitizedData, err := f.SanitizeData(action.Data)
				if err != nil {
					return errors.Join(fmt.Errorf("failed to sanitize feature data for action %s: %s", action.Id, err.Error()))
				}
				action.Data = sanitizedData
				action.FriendlyName = bridgeInfo.FriendlyName
				action.Client = client
				return action.Configure(registrar)
			}

		}

		// NOTE:
		// there are devices tha tcan be triggered in an action but dont have features.
		if e.Property == action.Property {

			sanitizedData, err := e.SanitizeData(action.Data)
			if err != nil {
				return errors.Join(fmt.Errorf("failed to sanitize expose data for action %s: %s", action.Id, err.Error()))
			}
			action.Data = sanitizedData
			action.FriendlyName = bridgeInfo.FriendlyName
			action.Client = client
			return action.Configure(registrar)

		}

	}

	return fmt.Errorf("property=%s for action=%s not found", action.Property, action.Id)
}
