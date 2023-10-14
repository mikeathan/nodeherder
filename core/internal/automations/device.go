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
	"strings"
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

type DeviceContextV2 struct {
	currentData map[string]any
	Payload     map[string]*devices.Entity
}

type DeviceContext struct {
	currentData map[string]any
	Payload     map[string]any
}

func (d *DeviceContext) GetCurrent(name string) any {
	return d.currentData[name]
}

func (d *DeviceContext) SetCurrent(name string, value any) {
	d.currentData[name] = value
}

func NewDeviceContext() *DeviceContext {
	return &DeviceContext{currentData: map[string]any{}, Payload: map[string]any{}}
}

func NewDeviceContextV2() *DeviceContextV2 {
	return &DeviceContextV2{currentData: map[string]any{}, Payload: make(map[string]*devices.Entity)}
}

func (d *DeviceContextV2) GetCurrentV2(name string) any {
	return d.currentData[name]
}

func (d *DeviceContextV2) SetCurrentV2(name string, value any) {
	d.currentData[name] = value
}

type Device struct {
	Id           string     `json:"id"`
	FriendlyName string     `json:"friendlyName"`
	Description  string     `json:"description"`
	Enabled      bool       `json:"enabled"`
	Triggers     []*Trigger `json:"triggers"`
	ctx          *DeviceContext
	ctxV2        *DeviceContextV2
}

func NewDevice(id string) *Device {

	d := &Device{
		Id:           id,
		FriendlyName: "",
		Description:  "",
		Enabled:      false,
		Triggers:     []*Trigger{},
		ctx:          NewDeviceContext(),
		ctxV2:        &DeviceContextV2{},
	}

	return d
}

func (d *Device) Evaluate(data map[string]any) bool {

	d.ctx.Payload = data
	for _, trigger := range d.Triggers {
		if _, ok := data[trigger.Name]; ok {
			trigger.process(d.ctx)
		}
	}

	return false
}

func (d *Device) EvaluateV2(device *devices.DeviceV2) bool {
	d.ctxV2.Payload = device.Exposes
	for _, trigger := range d.Triggers {
		if _, ok := device.Exposes[trigger.Name]; ok {
			trigger.processV2(d.ctxV2)
		}
	}
	return false
}

func (t *Device) Save(name string, pretty bool) error {

	//sanitize
	name = strings.Replace(name, " ", "_", -1)
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

func Saveutomations(triggers []*Device) {
	for _, trigger := range triggers {
		trigger.Save(trigger.FriendlyName, true)
	}
}
func LoadAutomations() []*Device {

	triggers := []*Device{}
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

func LoadTrigger(name string) (*Device, error) {
	filePath := getFilePath(name)

	return load(filePath)
}

func load(filePath string) (*Device, error) {

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	t := &Device{}
	err = json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func DeleteTrigger(name string) error {
	// sanitize
	name = strings.Replace(name, " ", "_", -1)
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

func (d *Device) configure(repo devices.Repository, client mqtt.MqttClient) error {

	//  check if device with automation id exists. friendyname can change
	bridgeInfo := repo.FindBridgeInfo(d.Id)
	if bridgeInfo == nil {
		return fmt.Errorf("automation id %s not found", d.Id)
	}

	if bridgeInfo.Disabled {
		return fmt.Errorf("device %s is disabled ", bridgeInfo.FriendlyName)
	}

	// validate conditions
	for _, trigger := range d.Triggers {

		// validate actions
		err := configureAction(repo, trigger.Action, client)
		if err != nil {
			return err
		}
	}

	d.FriendlyName = bridgeInfo.FriendlyName
	return nil
}

// validate actions
func configureAction(repo devices.Repository, action *MqttAction, client mqtt.MqttClient) error {

	bridgeInfo := repo.FindBridgeInfo(action.Id)
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
