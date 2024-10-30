package automations

import (
	"context"
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

var contextIgnoreList = []string{"action"}

type DeviceContext struct {
	currentData map[string]any
	Payload     map[string]*devices.Entity
	scheduler   *Scheduler
}

func NewDeviceContext(ctx context.Context) *DeviceContext {
	return &DeviceContext{
		currentData: map[string]any{},
		Payload:     make(map[string]*devices.Entity),
		scheduler:   NewScheduler(ctx),
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
	Id           string        `json:"id"`
	FriendlyName string        `json:"friendlyname"`
	Description  string        `json:"description"`
	Enabled      bool          `json:"enabled"`
	Triggers     []*Trigger    `json:"triggers"`
	Schedule     *TimeSchedule `json:"scheule"`
	ctx          *DeviceContext
}

func newDevice(ctx context.Context) *Device {

	d := &Device{
		Id:           "",
		FriendlyName: "",
		Description:  "",
		Enabled:      false,
		Triggers:     []*Trigger{},
		Schedule:     NewTimeSchedule(),
		ctx:          NewDeviceContext(ctx),
	}

	return d
}

func NewDevice(id string, ctx context.Context) *Device {

	d := &Device{
		Id:           id,
		FriendlyName: "",
		Description:  "",
		Enabled:      false,
		Triggers:     []*Trigger{},
		ctx:          NewDeviceContext(ctx),
	}

	return d
}

func (d *Device) Evaluate(device *devices.Device) bool {

	d.ctx.Payload = device.Exposes

	// triggers = automations for the device
	// NOTE: a trigger can have multiple conditions.
	// e.g presence can have multiple conditions for on and off
	for _, trigger := range d.Triggers {
		if _, ok := device.Exposes[trigger.Name]; ok {
			trigger.process(d.ctx)
		}
	}
	return false
}

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
		err := configureAction(registrar, trigger.Action, client)
		if err != nil {
			return err
		}
	}

	d.FriendlyName = bridgeInfo.FriendlyName

	return configureSchedule(d)
}

func configureSchedule(d *Device) error {

	if d.Schedule == nil || !d.Schedule.Enabled {
		return nil
	}

	// if we have schedule, disable automation and configure scheduler
	d.Enabled = false

	utils.LogInfof("adding schedule for automation id=%s, friendlyName=%s, start=%s, end=%s", d.Id, d.FriendlyName, d.Schedule.Start, d.Schedule.End)

	d.ctx.scheduler.Stop()

	err := d.ctx.scheduler.AddJob(d.Schedule.Start, func() error {
		d.Enabled = true
		return nil
	})

	if err != nil {
		return err
	}

	err = d.ctx.scheduler.AddJob(d.Schedule.End, func() error {
		d.Enabled = false
		return nil
	})

	if err != nil {
		return err
	}

	d.ctx.scheduler.Start()

	return nil
}

// validate actions
func configureAction(registrar services.DeviceRegistrar, action *MqttAction, client mqtt.MqttClient) error {

	bridgeInfo, err := registrar.FindBridgeInfo(action.Id)
	if err != nil {
		return err
	}

	for _, e := range bridgeInfo.Definition.Exposes {
		for _, f := range e.Features {

			if action.Property == f.Property {

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
				return action.Configure(registrar)
			}
		}
	}

	return fmt.Errorf("property=%s for action=%s not found", action.Property, action.Id)
}
