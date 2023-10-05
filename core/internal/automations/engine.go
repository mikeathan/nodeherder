package automations

import (
	"encoding/json"
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
	"node-herder/utils"
	"sort"
	"time"
)

type Engine interface { // TODO: might need to move it to Models????
	HandleDevice(id string, data map[string]any)
	Initialize(bridgeDevices []*devices.BridgeDevice) error
	GetAllTriggers() []byte
}

type AutomationEngine struct {
	deviceTriggers map[string]*Device
	mqttClient     mqtt.MqttClient
	configured     bool
}

func NewEngine(mqtt mqtt.MqttClient) *AutomationEngine {
	return &AutomationEngine{
		deviceTriggers: map[string]*Device{},
		mqttClient:     mqtt,
	}
}
func (a *AutomationEngine) HandleDevice(id string, data map[string]any) {
	if t, ok := a.deviceTriggers[id]; ok {
		t.Evaluate(data)
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

func (a *AutomationEngine) Initialize(bridgeDevices []*devices.BridgeDevice) error {

	// fake input data - TEST ONLY
	// tha needs to come from a file and loaded
	//triggers := newMockMqttTriggerPresenseWithLux(true)

	triggers := LoadTriggers()
	utils.LogInfof("Initialize automations")
	for _, deviceTrigger := range triggers {

		err := deviceTrigger.configure(bridgeDevices, a.mqttClient)
		if err != nil {
			return err
		}

		a.deviceTriggers[deviceTrigger.Name] = deviceTrigger

		utils.LogInfof("Loaded MqttTrigger %s, Enabled=%t", deviceTrigger.Description, deviceTrigger.Enabled)
		continue
	}

	a.configured = true
	return nil
}

// REMOVE
// used for testing only!!!!!!!!!!
func newMockMqttTriggerPresenseWithLux(enabled bool) []*Device {

	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff("Attic light", 5*time.Minute)
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux("Attic light", 30)

	// create device trigger
	deviceTrigger := NewDevice("Human presence")
	deviceTrigger.Description = "Attic light test automation"
	deviceTrigger.Enabled = enabled
	deviceTrigger.Triggers = []*Trigger{}
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	return []*Device{deviceTrigger}
}

func createTriggerTurnOnLightWithPresenceOnAndLux(name string, lux any) *Trigger {
	// action = turn off light
	turnOnAction := &MqttAction{}
	turnOnAction.Friendlyname = name
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

func createTriggerDelayTurnOffLightWithPresenceOff(name string, delay time.Duration) *Trigger {
	// action = turn off light
	turnOffAction := &MqttAction{}
	turnOffAction.Friendlyname = name
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
