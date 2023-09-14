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
	automationDir = "../../configs/automations"
	automationExt = ".config"
)

type MqttTrigger struct {
	Type        string             `json:"trigger"`
	DeviceName  string             `json:"devicename"`
	Description string             `json:"description"`
	Conditions  []*DeviceCondition `json:"conditions"`
	Enabled     bool               `json:"enabled"`
}

func newMqttTrigger() *MqttTrigger {
	return &MqttTrigger{
		Type:        "mqtt",
		DeviceName:  "",
		Description: "",
		Conditions:  []*DeviceCondition{},
		Enabled:     true,
	}
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

func (t *MqttTrigger) Save(name string, pretty bool) error {
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

func LoadTrigger(name string) (*MqttTrigger, error) {
	filePath := getFilePath(name)

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	t := newMqttTrigger()
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

func (t *MqttTrigger) configure(bridgeDevices []*devices.BridgeDevice, client mqtt.MqttClient) error {
	fmt.Println("Loading mqtt trigger:", t.Description)
	if t.Type != "mqtt" {
		return errors.New("unsupported mqtt type ")
	}

	// validate conditions
	for _, condition := range t.Conditions {

		// validate actions
		err := validateAction(bridgeDevices, condition.Action, client)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *MqttTrigger) Evaluate(data map[string]any) {

	if !t.Enabled {
		utils.LogDebugf("Trigger %s is disabled\n", t.Description)
		return
	}

	for _, condition := range t.Conditions {
		condition.Evaluate(data)
	}
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
