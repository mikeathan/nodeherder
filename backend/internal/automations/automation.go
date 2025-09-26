package automations

import (
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/utils/storage"
	"reflect"
)

type AutomationType string

const (
	DeviceAutomationType AutomationType = "device"
)

var automationTypeRegistry = map[AutomationType]reflect.Type{
	DeviceAutomationType: reflect.TypeOf(&Device{})}

type TriggerEvent interface {
	Type() string
}

type Automation interface {
	Evaluate(event TriggerEvent) bool
	EvaluateTrigger(event TriggerEvent, triggerName string) bool
	Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error
	GetId() string
	GetFriendlyName() string
	IsEnabled() bool
	GetTriggers() TriggerList
	GetSchedules() []*TimeSchedule

	AddTrigger(trigger Trigger) error
	RemoveTrigger(index int) error
	SetEnabled(enabled bool)
}

type BaseAutomation struct {
	Id           string          `json:"id"`
	Type         AutomationType  `json:"type"`
	FriendlyName string          `json:"friendlyname"`
	Description  string          `json:"description"`
	Enabled      bool            `json:"enabled"`
	Triggers     TriggerList     `json:"triggers"`
	Schedules    []*TimeSchedule `json:"schedules"`
}

func NewBaseAutomation() *BaseAutomation {
	return &BaseAutomation{
		Id:           "",
		Type:         "",
		FriendlyName: "",
		Description:  "",
		Enabled:      false,
		Triggers:     TriggerList{},
		Schedules:    []*TimeSchedule{},
	}
}

func (b *BaseAutomation) GetId() string                 { return b.Id }
func (b *BaseAutomation) GetFriendlyName() string       { return b.FriendlyName }
func (b *BaseAutomation) IsEnabled() bool               { return b.Enabled }
func (b *BaseAutomation) GetTriggers() TriggerList      { return b.Triggers }
func (b *BaseAutomation) GetSchedules() []*TimeSchedule { return b.Schedules }
func (b *BaseAutomation) SetEnabled(enabled bool)       { b.Enabled = enabled }

func (b *BaseAutomation) AddTrigger(trigger Trigger) error {
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

func (b *BaseAutomation) EvaluateTrigger(event TriggerEvent, triggerName string) bool {
	return false
}

func (d *BaseAutomation) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

	// validate conditions
	for _, trigger := range d.Triggers {

		// validate actions
		for _, action := range trigger.GetActions() {
			err := action.Configure(registrar, client)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AutomationSerialiser
// This is used in storage loader to deserialise the object back to supported automation type

type AututomationSerialiser struct {
}

func NewAutomationSerialiser() storage.Serializer[AutomationType, Automation] {
	return storage.Serializer[AutomationType, Automation]{TypeRegistry: automationTypeRegistry, TypeField: "type"}
}

// type automationSerialiser struct {
// 	Automation Automation `json:"-"`
// }

// func (w *automationSerialiser) MarshalJSON() ([]byte, error) {
// 	if w.Automation != nil {
// 		return json.Marshal(w.Automation)
// 	}
// 	return []byte(`{}`), nil // or return an error if you prefer
// }

// func (w *automationSerialiser) UnmarshalJSON(data []byte) error {
// 	var temp struct {
// 		Type string `json:"type"`
// 	}
// 	if err := json.Unmarshal(data, &temp); err != nil {
// 		return err
// 	}

// 	concreteType, ok := automationTypeRegistry[AutomationType(temp.Type)]
// 	if !ok {
// 		return fmt.Errorf("unknown automation type: %s", temp.Type)
// 	}

// 	concrete := reflect.New(concreteType).Interface().(Automation)
// 	if err := json.Unmarshal(data, concrete); err != nil {
// 		return err
// 	}

// 	w.Automation = concrete
// 	return nil
// }
