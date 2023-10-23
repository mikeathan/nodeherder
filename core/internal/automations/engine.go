package automations

import (
	"errors"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/models/devices"
	"node-herder/utils"
	"node-herder/utils/storage"
	"time"
)

const (
	automationDir = "configs/automations"
	automationExt = ".config"
)

type Engine interface { // TODO: might need to move it to Models????
	HandleDeviceV2(device *devices.DeviceV2)
	Add(automation *Device) error
	Delete(id string) error
	DeleteTrigger(id string, triggerId int) error
	Initialize()
	Load(id string) (*Device, error)
	GetAllTriggers() []*Device
}

type AutomationEngine struct {
	mqttClient        mqtt.MqttClient
	registrar         *services.DeviceRegistrar
	automationStorage storage.Storage[Device]
}

func NewEngine(registrar *services.DeviceRegistrar, mqtt mqtt.MqttClient) *AutomationEngine {
	return &AutomationEngine{
		mqttClient:        mqtt,
		registrar:         registrar,
		automationStorage: storage.NewJsonDiskStorage[Device](automationDir),
	}
}

func (a *AutomationEngine) WithStorage(storage storage.Storage[Device]) {
	a.automationStorage = storage
}

func (a *AutomationEngine) HandleDeviceV2(device *devices.DeviceV2) {
	trigger, err := a.automationStorage.Load(device.Id)
	if err == nil {
		trigger.EvaluateV2(device)
	}
}

func (a *AutomationEngine) GetAllTriggers() []*Device {
	return a.automationStorage.LoadAll()
}

func (a *AutomationEngine) Load(id string) (*Device, error) {
	return a.automationStorage.Load(id)
}

func (a *AutomationEngine) Add(automation *Device) error {

	utils.LogInfof("adding automation id=%s, friendlyName=%s, enabled=%v", automation.Id, automation.FriendlyName, automation.Enabled)
	err := automation.configure(a.registrar, a.mqttClient)
	if err != nil {
		utils.LogErrorf("configure automation id %s failed. Error=%s", automation.Id, err.Error())
		return err
	}

	a.automationStorage.Store(automation.Id, automation)

	return nil
}

func (a *AutomationEngine) DeleteTrigger(id string, triggerId int) error {
	automation, err := a.automationStorage.Load(id)
	if err != nil {
		return err
	}

	if triggerId >= len(automation.Triggers) {
		return errors.New("trigger index out of bounds")
	}

	// remove trigger
	automation.Triggers = append(automation.Triggers[:triggerId], automation.Triggers[triggerId+1:]...)

	// store
	a.automationStorage.Store(id, automation)
	return nil
}

func (a *AutomationEngine) Delete(id string) error {

	a.automationStorage.Delete(id)
	return nil
}

func (a *AutomationEngine) Initialize() {

	// fake input data - TEST ONLY
	// that needs to come from a file and loaded
	//automations := newMockMqttTriggerPresenseWithLux(true)
	//automations[0].Save("human_presence", true)

	utils.LogInfof("Initialize automations")
	automations, err := a.automationStorage.Initialize()
	if err != nil {
		utils.LogErrorf("Load automations failed. Error=%s", err.Error())
		return
	}

	for _, automation := range automations {

		utils.LogInfof("Loading automation id= %s, friendlyName=%s, Enabled=%t", automation.Id, automation.FriendlyName, automation.Enabled)
		if err != nil {
			utils.LogErrorf("configure automation id %s failed. Error=%s", automation.Id, err.Error())
			continue
		}
	}

	//automations[0].Save(, true)
}

// REMOVE
// used for testing only!!!!!!!!!!
func newMockMqttTriggerPresenseWithLux(enabled bool) []*Device {

	// 0x70ac08fffefafeca = Attic light
	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff("0x70ac08fffefafeca", 5*time.Minute)
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux("0x70ac08fffefafeca", 30)

	// create device trigger
	// 0xa4c13894070052fc = Human Presence
	deviceTrigger := NewDevice("0xa4c13894070052fc")
	deviceTrigger.Description = "Attic light test automation"
	deviceTrigger.Enabled = enabled
	deviceTrigger.Triggers = []*Trigger{}
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	return []*Device{deviceTrigger}
}

func createTriggerTurnOnLightWithPresenceOnAndLux(id string, lux any) *Trigger {
	// action = turn off light
	turnOnAction := &MqttAction{}
	turnOnAction.Id = id
	turnOnAction.Type = "light"
	turnOnAction.Property = "state"
	turnOnAction.Data = true
	turnOnAction.Delay = 0

	// Turn on sensor trigger
	turnOnTrigger := &Trigger{}
	turnOnTrigger.Name = "presence"
	turnOnTrigger.Action = turnOnAction

	// condition = presence = off && lux <= 30
	turnOnCondition := &Condition{}
	turnOnCondition.Name = "presence"
	turnOnCondition.EqualityOperator = "="
	turnOnCondition.Value = true

	luxCondition := &Condition{}
	luxCondition.Name = "lux"
	luxCondition.EqualityOperator = "<="
	luxCondition.Value = lux

	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, luxCondition)

	return turnOnTrigger
}

func createTriggerDelayTurnOffLightWithPresenceOff(id string, delay time.Duration) *Trigger {
	// action = turn off light
	turnOffAction := &MqttAction{}
	turnOffAction.Id = id
	turnOffAction.Type = "light"
	turnOffAction.Property = "state"
	turnOffAction.Data = false
	turnOffAction.Delay = delay

	// Turn off sensor trigger
	turnOffTrigger := &Trigger{}
	turnOffTrigger.Name = "presence"
	turnOffTrigger.Action = turnOffAction

	// condition = presence == false
	turnOffCondition := &Condition{}
	turnOffCondition.Name = "presence"
	turnOffCondition.EqualityOperator = "="
	turnOffCondition.Value = false

	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffCondition)

	return turnOffTrigger
}
