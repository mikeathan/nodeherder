package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/models/devices"
	"node-herder/utils"
	"node-herder/utils/storage"
	"reflect"
)

type AutomationType string

const (
	automationDir                       = "configs/automations"
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
	// Temporary struct to read the type
	var temp struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	concreteType, ok := automationTypeRegistry[AutomationType(temp.Type)]
	if !ok {
		return fmt.Errorf("unknown automation type: %s", temp.Type)
	}

	// Allocate a concrete struct
	concrete := reflect.New(concreteType).Interface()

	// Unmarshal into concrete type
	if err := json.Unmarshal(data, concrete); err != nil {
		return err
	}

	// Copy all fields back into BaseAutomation
	if conv, ok := concrete.(interface{ Base() *BaseAutomation }); ok {
		*a = *conv.Base()
	}

	return nil
}

type Engine interface { // TODO: might need to move it to Models????
	HandleDevice(device *devices.Device)
	Add(automation *Device) error
	Delete(id string) error
	DeleteTrigger(id string, triggerId int) error
	Initialize()
	Load(id string) (*Device, error)
	IsAutomationEnabled(id string) bool
	GetAllTriggers() []*Device
	WithStorage(storage storage.Storage[Device])
}

type AutomationEngine struct {
	mqttClient mqtt.MqttClient
	registrar  services.DeviceRegistrar
	storage    storage.Storage[Device]
	handlers   []AutomationHandler
}

func NewEngine(handlerFactory []AutomationHandler, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient) *AutomationEngine {

	return &AutomationEngine{
		mqttClient: mqtt,
		registrar:  registrar,
		storage: storage.NewJsonDiskStorage(automationDir, func() *Device {
			return newDevice()
		}),
		handlers: handlerFactory,
	}
}

func (a *AutomationEngine) WithStorage(storage storage.Storage[Device]) {
	a.storage = storage
}

func (a *AutomationEngine) IsAutomationEnabled(id string) bool {
	automation, err := a.storage.LoadFromCache(id)
	if err == nil {
		return automation.Enabled
	}

	return false
}

func (a *AutomationEngine) HandleDevice(device *devices.Device) {
	automation, err := a.storage.LoadFromCache(device.Id)
	if err == nil && automation.Enabled {
		automation.Evaluate(device)
	}
}

func (a *AutomationEngine) GetAllTriggers() []*Device {
	return a.storage.LoadAll()
}

func (a *AutomationEngine) Load(id string) (*Device, error) {
	return a.storage.Load(id)
}

func (a *AutomationEngine) Add(automation *Device) error {

	utils.LogInfof("adding automation id=%s, friendlyName=%s, enabled=%v", automation.Id, automation.FriendlyName, automation.Enabled)
	err := a.configureAutomation(automation)
	if err != nil {
		utils.LogErrorf("configure automation id %s failed. Error=%s", automation.Id, err.Error())
		return err
	}

	a.storage.Store(automation.Id, automation)

	return nil
}

func (a *AutomationEngine) DeleteTrigger(id string, triggerId int) error {
	automation, err := a.storage.Load(id)
	if err != nil {
		return err
	}

	if triggerId >= len(automation.Triggers) {
		return errors.New("trigger index out of bounds")
	}

	// remove trigger
	automation.Triggers = append(automation.Triggers[:triggerId], automation.Triggers[triggerId+1:]...)

	// store
	a.storage.Store(id, automation)
	return nil
}

func (a *AutomationEngine) Delete(id string) error {

	a.storage.Delete(id)
	return nil
}

func (a *AutomationEngine) Initialize() {

	utils.LogInfof("Initialize automations")
	automations, err := a.storage.Initialize() // <--------------- check if we clear any internal cache in automations after reloading
	if err != nil {
		utils.LogErrorf("Load automations failed. Error=%s", err.Error())
		return
	}

	for _, automation := range automations {

		utils.LogInfof("Loading automation id= %s, friendlyName=%s, Enabled=%t", automation.Id, automation.FriendlyName, automation.Enabled)

		err := a.configureAutomation(automation)
		if err != nil {
			utils.LogErrorf("configure automation id %s failed. Error=%s", automation.Id, err.Error())
			continue
		}
	}
}

func (a *AutomationEngine) configureAutomation(automation *Device) error {

	err := automation.configure(a.registrar, a.mqttClient)
	if err != nil {
		return err
	}

	for _, handler := range a.handlers {
		err = handler.Process(automation)
		if err != nil {
			return err
		}
	}

	return nil
}
