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

type TriggerEvent interface {
	Type() string
}

type Automation interface {
	Evaluate(event TriggerEvent) bool
	Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error
	GetId() string
	GetFriendlyName() string
	IsEnabled() bool
	GetTriggers() []*Trigger
	GetSchedules() []*TimeSchedule

	AddTrigger(trigger *Trigger) error
	RemoveTrigger(index int) error
	SetEnabled(enabled bool)
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

func NewBaseAutomation() *BaseAutomation {
	return &BaseAutomation{
		Id:           "",
		Type:         "",
		FriendlyName: "",
		Description:  "",
		Enabled:      false,
		Triggers:     []*Trigger{},
		Schedules:    []*TimeSchedule{},
	}
}

func (b *BaseAutomation) GetId() string                 { return b.Id }
func (b *BaseAutomation) GetFriendlyName() string       { return b.FriendlyName }
func (b *BaseAutomation) IsEnabled() bool               { return b.Enabled }
func (b *BaseAutomation) GetTriggers() []*Trigger       { return b.Triggers }
func (b *BaseAutomation) GetSchedules() []*TimeSchedule { return b.Schedules }
func (b *BaseAutomation) SetEnabled(enabled bool)       { b.Enabled = enabled }

func (b *BaseAutomation) AddTrigger(trigger *Trigger) error {
	b.Triggers = append(b.Triggers, trigger)
	return nil
}

func (b *BaseAutomation) RemoveTrigger(index int) error {
	if index >= len(b.Triggers) {
		return fmt.Errorf("trigger index out of bounds")
	}

	b.Triggers = append(b.Triggers[:index], b.Triggers[index+1:]...)
	return nil
}

func (b *BaseAutomation) Evaluate(event TriggerEvent) bool {
	return false
}


TODO
func UnmarshalAutomation(data []byte) (Automation, error) {
    var temp struct {
        Type string `json:"type"`
    }
    if err := json.Unmarshal(data, &temp); err != nil {
        return nil, err
    }

    concreteType, ok := automationTypeRegistry[AutomationType(temp.Type)]
    if !ok {
        return nil, fmt.Errorf("unknown automation type: %s", temp.Type)
    }

    concrete := reflect.New(concreteType).Interface()
    if err := json.Unmarshal(data, concrete); err != nil {
        return nil, err
    }

    return concrete.(Automation), nil
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

	concreteType, ok := automationTypeRegistry[AutomationType(typ)]
	if !ok {
		return fmt.Errorf("unknown automation type: %s", typ)
	}

	// prevent recursive call by aliasing BaseAutomation
	type baseAlias BaseAutomation
	var base baseAlias
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	*a = BaseAutomation(base)

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
