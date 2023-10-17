package automations

import (
	"encoding/json"
	"node-herder/internal/mqtt"
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

	Initialize()
	GetAllTriggers() []byte
}

type AutomationEngine struct {
	mqttClient        mqtt.MqttClient
	repo              devices.Repository
	automationStorage storage.Storage[Device]
}

func NewEngine(mqtt mqtt.MqttClient, repo devices.Repository) *AutomationEngine {
	return &AutomationEngine{
		mqttClient:        mqtt,
		repo:              repo,
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

func (a *AutomationEngine) GetAllTriggers() []byte {
	triggers := a.automationStorage.LoadAll()

	bytes, err := json.Marshal(triggers)
	if err != nil {
		utils.LogErrorf("marshal automation triggers failed. error %s", err.Error())
	}
	return bytes
}

func (a *AutomationEngine) Add(automation *Device) error {

	utils.LogInfof("adding automation id=%s, friendlyName=%s, enabled=%v", automation.Id, automation.FriendlyName, automation.Enabled)
	err := automation.configure(a.repo, a.mqttClient)
	if err != nil {
		utils.LogErrorf("configure automation id %s failed. Error=%s", automation.Id, err.Error())
		return err
	}

	a.automationStorage.Store(automation.Id, automation)

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
		err := automation.configure(a.repo, a.mqttClient)
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
