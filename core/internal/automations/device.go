package automations

import (
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

type DeviceContext struct {
	currentData map[string]any
	Payload     map[string]*devices.Entity
}

func NewDeviceContext() *DeviceContext {
	return &DeviceContext{currentData: map[string]any{}, Payload: make(map[string]*devices.Entity)}
}

func (d *DeviceContext) GetCurrent(name string) any {
	return d.currentData[name]
}

func (d *DeviceContext) SetCurrent(name string, value any) {
	d.currentData[name] = value
}

type Device struct {
	Id           string     `json:"id"`
	FriendlyName string     `json:"friendlyname"`
	Description  string     `json:"description"`
	Enabled      bool       `json:"enabled"`
	Triggers     []*Trigger `json:"triggers"`
	ctx          *DeviceContext
}

func newDevice() *Device {

	d := &Device{
		Id:           "",
		FriendlyName: "",
		Description:  "",
		Enabled:      false,
		Triggers:     []*Trigger{},
		ctx:          NewDeviceContext(),
	}

	return d
}

func NewDevice(id string) *Device {

	d := &Device{
		Id:           id,
		FriendlyName: "",
		Description:  "",
		Enabled:      false,
		Triggers:     []*Trigger{},
		ctx:          NewDeviceContext(),
	}

	return d
}

func (d *Device) Evaluate(device *devices.Device) bool {
	d.ctx.Payload = device.Exposes
	for _, trigger := range d.Triggers {
		if _, ok := device.Exposes[trigger.Name]; ok {
			trigger.process(d.ctx)
		}
	}
	return false
}

func (d *Device) configure(registrar *services.DeviceRegistrar, client mqtt.MqttClient) error {

	//  check if device with automation id exists. friendyname can change
	bridgeInfo := registrar.FindBridgeInfo(d.Id)
	if bridgeInfo == nil {
		return fmt.Errorf("automation id %s not found", d.Id)
	}

	if bridgeInfo.Disabled {
		return fmt.Errorf("device %s is disabled ", bridgeInfo.FriendlyName)
	}

	// validate conditions
	for _, trigger := range d.Triggers {

		// validate actions
		err := configureAction(registrar, trigger.Action, client)
		if err != nil {
			return err
		}
	}

	d.FriendlyName = bridgeInfo.FriendlyName
	return nil
}

// validate actions
func configureAction(registrar *services.DeviceRegistrar, action *MqttAction, client mqtt.MqttClient) error {

	bridgeInfo := registrar.FindBridgeInfo(action.Id)
	if bridgeInfo == nil {
		return fmt.Errorf("action id %s not found", action.Id)
	}

	for _, e := range bridgeInfo.Definition.Exposes {
		for _, f := range e.Features {

			if action.Property == f.Property {

				if action.Type != e.Type {
					return fmt.Errorf("type=%s for action=%s not found", action.Type, action.Id)
				}

				// sanitize data
				if f.Type == "binary" {
					if value, ok := action.Data.(bool); ok {
						if value {
							action.Data = f.ValueOn
						} else {
							action.Data = f.ValueOff
						}
					}
				} else if f.Type == "numeric" {
					if value, ok := action.Data.(int); ok {

						if min, ok := f.ValueMin.(int); ok {
							if value < min {
								return fmt.Errorf("value=%d for action=%s smaller than Minimum %d", value, action.Id, min)
							}
						}

						if max, ok := f.ValueMax.(int); ok {
							if value > max {
								return fmt.Errorf("value=%d for action=%s bigger than Maximum %d", value, action.Id, max)
							}
						}
					}
				} else {
					return fmt.Errorf("type=%s for action=%s not implemented", f.Type, action.Id)
				}

				action.FriendlyName = bridgeInfo.FriendlyName
				action.Client = client
				return nil
			}
		}
	}

	return fmt.Errorf("property=%s for action=%s not found", action.Property, action.Id)
}
