package automations

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
	"node-herder/utils"
	"os"
	"path/filepath"
)

const (
	automationDir = "configs/automations"
	automationExt = ".config"
)

// examples
// sensor name = condition 1 & condiiton 1  && condtion n.... = action
// presence = (true) && (lux <= 30) = turn on
// presence = (false) && timer condition = turn off
// presence = true = turn on
// presence = false = turn off

type DeviceContext struct {
	data map[string]any
}

func (d *DeviceContext) Get(name string) any {
	return d.data[name]
}

func (d *DeviceContext) Set(name string, value any) {
	d.data[name] = value
}

func NewDeviceContext() *DeviceContext {
	return &DeviceContext{data: map[string]any{}}
}

type DeviceTrigger struct {
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	Enabled        bool                        `json:"enabled"`
	SensorTriggers map[string][]*SensorTrigger `json:"sensor_triggers"`
	deviceContext  *DeviceContext
}

func NewDeviceTrigger(name string) *DeviceTrigger {
	return &DeviceTrigger{
		Name:           name,
		Description:    "",
		Enabled:        false,
		SensorTriggers: make(map[string][]*SensorTrigger),
		deviceContext:  NewDeviceContext(),
	}
}

func (d *DeviceTrigger) Evaluate(data map[string]any) bool {
	for sensor := range data {
		if triggers, ok := d.SensorTriggers[sensor]; ok {
			for _, trigger := range triggers {
				d.processTrigger(trigger, data)
			}
		}
	}
	return false
}

func (d *DeviceTrigger) processTrigger(trigger *SensorTrigger, data map[string]any) {

	currValue := d.deviceContext.Get(trigger.Name)
	for _, c := range trigger.Conditions {

		isMatched := c.Evaluate(data)
		if !isMatched {
			trigger.Action.Stop()
			return
		}

		// avoid calling action again for sensor if value hasnt changed
		if trigger.Name == c.Name && currValue == c.Value {
			return
		}
	}

	trigger.Action.Execute(func() {
		// on success callback
		// update sensor current value
		d.deviceContext.Set(trigger.Name, data[trigger.Name])
	})
}

func (t *DeviceTrigger) Save(name string, pretty bool) error {
	filePath := getFilePath(name)

	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	if pretty {
		data, err = prettyJson(data)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func getFilePath(name string) string {

	createDirIfNotExists(automationDir)
	return filepath.Join(automationDir, fmt.Sprintf("%s%s", name, automationExt))
}

func prettyJson(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func LoadTriggers() []*DeviceTrigger {

	triggers := []*DeviceTrigger{}
	err := filepath.Walk(automationDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			utils.LogErrorf("Error loading automations %s", err.Error())
			return err
		}
		if info.IsDir() {
			return nil
		}

		trigger, err := load(path)
		if err != nil {
			utils.LogErrorf("Error loading automation %s %s", path, err.Error())
			return err
		}

		triggers = append(triggers, trigger)
		return nil
	})

	if err != nil {
		utils.LogErrorf("Error loading automations %s", err.Error())
	}
	return triggers
}

func LoadTrigger(name string) (*DeviceTrigger, error) {
	filePath := getFilePath(name)

	return load(filePath)
}

func load(filePath string) (*DeviceTrigger, error) {

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	t := &DeviceTrigger{}
	err = json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}

	return t, nil
}
func DeleteTrigger(name string) error {
	filePath := getFilePath(name)

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}

func createDirIfNotExists(name string) {
	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(name, os.ModePerm)
		if err != nil {
			utils.LogErrorf(fmt.Sprintf("Failed to create automations directory %s Error: %v", name, err))
			panic(err)
		}
	}
}

func (t *DeviceTrigger) configure(bridgeDevices []*devices.BridgeDevice, client mqtt.MqttClient) error {

	// validate conditions
	for _, triggers := range t.SensorTriggers {

		for _, trigger := range triggers {
			// validate actions
			err := validateAction(bridgeDevices, trigger.Action, client)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// validate actions
func validateAction(bridgeDevices []*devices.BridgeDevice, action *MqttAction, client mqtt.MqttClient) error {

	for _, device := range bridgeDevices {
		if device.FriendlyName == action.Friendlyname {

			if action.Type != "light" {
				return fmt.Errorf("unsupported action type %s", action.Type)
			}

			if device.Disabled {
				return errors.New("device is disabled ")
			}
			action.Client = client

			// TODO: refactor. no need to keep looping once we found our value
			// need to fix state, convert it to expected one: { "state": "ON" }'
			for _, expose := range device.Definition.Exposes {
				for _, feature := range expose.Features {
					if feature.Property == action.Property { // state property only!

						// for now we only support "state" property
						if action.Value == true {
							action.Value = feature.ValueOn
						} else {
							action.Value = feature.ValueOff
						}
						break
					}
				}
			}
		}
	}
	return nil
}

// func (t *TimerTrigger) configure(bridgeDevices []*devices.BridgeDevice, client mqtt.MqttClient) error {
// 	if t.Type != "timer" {
// 		return fmt.Errorf("unsupported trigger type %s", t.Type)
// 	}

// 	// valdate conditions
// 	_, ok := t.Condition.(*TimeDurationCondition)
// 	if !ok {
// 		_, ok = t.Condition.(*TimestampCondition)
// 		if !ok {
// 			return errors.New("unsupported condition type ")
// 		}
// 	}

// 	// validate actions
// 	err := validateAction(bridgeDevices, t.Action, client)

// 	if err != nil {
// 		return err
// 	}

// 	t.run()
// 	return nil
// }

// type TimerTrigger struct {
// 	Type      string      `json:"trigger"`
// 	Name      string      `json:"name"`
// 	Condition interface{} `json:"condition"`
// 	Action    *MqttAction `json:"action"`
// 	Enabled   bool        `json:"enabled"`
// }

// func newTimerTrigger() *TimerTrigger {
// 	return &TimerTrigger{
// 		Type:    "timer",
// 		Name:    "",
// 		Action:  &MqttAction{},
// 		Enabled: true,
// 	}
// }

// func (t *TimerTrigger) Trigger() error {

// 	fmt.Printf("Trigger Action %s\n", t.Action.Friendlyname)
// 	return t.Action.Run()
// }

// func (t *TimerTrigger) run() {

// 	if !t.Enabled {
// 		fmt.Printf("%s is disabled\n", t.Name)
// 		return
// 	}

// 	tc := t.Condition.(TimerCondition)
// 	go func() {
// 		for {

// 			//
// 			// todo: allow enable/disable triggers
// 			//
// 			diff := time.Until(tc.GetSchedule()).Seconds()
// 			ticker := *time.NewTicker(time.Duration(diff) * time.Second)
// 			select {
// 			case <-ticker.C:
// 				err := t.Trigger()
// 				if err != nil {
// 					fmt.Println(err)
// 				}

// 				if !tc.IsRepeat() {
// 					fmt.Println("timer finished. no repeat. exiting")
// 					return
// 				}
// 			}
// 		}
// 	}()
// }
