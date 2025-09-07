package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"reflect"
)

type AutomationType string

const (
	DeviceAutomationType AutomationType = "device"
	SystemAutomationType AutomationType = "system"
)

var automationTypeRegistry = map[AutomationType]reflect.Type{
	"device": reflect.TypeOf(Device{}),
}

type BaseAutomation struct {
	Id           string          `json:"id"`
	Type         AutomationType  `json:"type"`
	FriendlyName string          `json:"friendlyname"`
	Description  string          `json:"description"`
	Enabled      bool            `json:"enabled"`
	Triggers     []*Trigger      `json:"triggers"`
	Schedules    []*TimeSchedule `json:"schedules"`
}

func (a *BaseAutomation) UnmarshalJSON(data []byte) error {
	var temp map[string]json.RawMessage
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	var typ string
	if t, ok := temp["type"]; ok {
		if err := json.Unmarshal(t, &typ); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("missing type field in automation")
	}

	// find automation type
	concreteType, ok := automationTypeRegistry[AutomationType(typ)]
	if !ok {
		return fmt.Errorf("unknown automation type: %s", typ)
	}

	concrete := reflect.New(concreteType).Interface()
	if err := json.Unmarshal(data, concrete); err != nil {
		return err
	}

	return nil
}
func (d *BaseAutomation) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

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

	return nil
}
