package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
	"node-herder/utils"
	"sort"
	"sync"
	"time"
)

type Engine interface { // TODO: might need to move it to Models????
	HandleDevice(id string, data map[string]any)
	HandleDeviceV2(device *devices.DeviceV2)
	Add(automation *Device) error
	Delete(id string) error

	Initialize()
	GetAllTriggers() []byte
}

type AutomationEngine struct {
	deviceTriggers map[string]*Device
	mqttClient     mqtt.MqttClient
	repo           devices.Repository
	mutex          sync.RWMutex
}

func NewEngine(mqtt mqtt.MqttClient, repo devices.Repository) *AutomationEngine {
	return &AutomationEngine{
		deviceTriggers: map[string]*Device{},
		mqttClient:     mqtt,
		repo:           repo,
		mutex:          sync.RWMutex{},
	}
}

func (a *AutomationEngine) HandleDevice(id string, data map[string]any) {
	if t, ok := a.deviceTriggers[id]; ok {
		t.Evaluate(data)
	}
}

func (a *AutomationEngine) Clear() {
	for k := range a.deviceTriggers {
		delete(a.deviceTriggers, k)
	}
}

func (a *AutomationEngine) HandleDeviceV2(device *devices.DeviceV2) {
	if t, ok := a.deviceTriggers[device.Id]; ok {
		t.EvaluateV2(device)
	}
}

func (a *AutomationEngine) GetAllTriggers() []byte {
	keys := make([]string, 0, len(a.deviceTriggers))
	values := make([]*Device, 0, len(a.deviceTriggers))

	for k, _ := range a.deviceTriggers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		values = append(values, a.deviceTriggers[k])
	}

	bytes, err := json.Marshal(values)
	if err != nil {
		utils.LogErrorf("marshal automation triggers failed. error %s", err.Error())
	}
	return bytes
}

func (a *AutomationEngine) Add(automation *Device) error {

	defer a.mutex.Unlock()
	a.mutex.Lock()

	utils.LogInfof("adding automation id=%s, friendlyName=%s, enabled=%v", automation.Id, automation.FriendlyName, automation.Enabled)
	err := automation.configure(a.repo, a.mqttClient)
	if err != nil {
		utils.LogErrorf("configure automation id %s failed. Error=%s", automation.Id, err.Error())
		return err
	}

	a.deviceTriggers[automation.Id] = automation

	// store to file
	automation.Save(automation.Id, true)
	return nil
}

func (a *AutomationEngine) Delete(id string) error {

	defer a.mutex.Unlock()
	a.mutex.Lock()

	utils.LogInfof("deleting automation id=%s", id)
	if _, ok := a.deviceTriggers[id]; !ok {

		utils.LogErrorf("delete automation id %s failed. Error=automation not found", id)
		return fmt.Errorf("automation %s not found", id)
	}

	delete(a.deviceTriggers, id)

	// delete file
	DeleteTrigger(id)
	return nil
}
func (a *AutomationEngine) Initialize() {

	// fake input data - TEST ONLY
	// that needs to come from a file and loaded
	//automations := newMockMqttTriggerPresenseWithLux(true)
	//automations[0].Save("human_presence", true)

	a.Clear()
	automations := LoadAutomations()

	utils.LogInfof("Initialize automations")
	for _, automation := range automations {

		utils.LogInfof("Loading automation id= %s, friendlyName=%s, Enabled=%t", automation.Id, automation.FriendlyName, automation.Enabled)
		err := automation.configure(a.repo, a.mqttClient)
		if err != nil {
			utils.LogErrorf("configure automation id %s failed. Error=%s", automation.Id, err.Error())
			continue
		}

		a.deviceTriggers[automation.Id] = automation

		continue
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
