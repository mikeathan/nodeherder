package automations

import (
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
	"node-herder/utils"
	"time"
)

type Engine interface { // TODO: might need to move it to Models????
	HandleDevice(id string, data map[string]any)
	Load(bridgeDevices []*devices.BridgeDevice) error
}

type AutomationEngine struct {
	deviceTriggers map[string]*DeviceTrigger
	mqttClient     mqtt.MqttClient
	configured     bool
}

func NewEngine(mqtt mqtt.MqttClient) *AutomationEngine {
	return &AutomationEngine{
		deviceTriggers: map[string]*DeviceTrigger{},
		mqttClient:     mqtt,
	}
}
func (a *AutomationEngine) HandleDevice(id string, data map[string]any) {
	if t, ok := a.deviceTriggers[id]; ok {
		t.Evaluate(data)
	}
}

func (a *AutomationEngine) Load(bridgeDevices []*devices.BridgeDevice) error {

	// fake input data - TEST ONLY
	// tha needs to come from a file and loaded
	//triggers := newMockMqttTriggerPresenseWithLux(true)

	triggers := LoadTriggers()
	utils.LogInfof("Loading automations")
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
func newMockMqttTriggerPresenseWithLux(enabled bool) []*DeviceTrigger {

	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff("Attic light", 5*time.Minute)
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux("Attic light", 30)

	// create device trigger
	deviceTrigger := NewDeviceTrigger("Human presence")
	deviceTrigger.Description = "Attic light test automation"
	deviceTrigger.Enabled = enabled
	deviceTrigger.SensorTriggers = make(map[string][]*SensorTrigger)
	deviceTrigger.SensorTriggers[turnOffTrigger.Name] = append(deviceTrigger.SensorTriggers[turnOffTrigger.Name], turnOffTrigger)
	deviceTrigger.SensorTriggers[turnOnTrigger.Name] = append(deviceTrigger.SensorTriggers[turnOnTrigger.Name], turnOnTrigger)

	return []*DeviceTrigger{deviceTrigger}
}

func createTriggerTurnOnLightWithPresenceOnAndLux(name string, lux any) *SensorTrigger {
	// action = turn off light
	turnOnAction := &MqttAction{}
	turnOnAction.Friendlyname = name
	turnOnAction.Type = "light"
	turnOnAction.Property = "state"
	turnOnAction.Value = true
	turnOnAction.Delay = 0

	// Turn on sensor trigger
	turnOnTrigger := &SensorTrigger{}
	turnOnTrigger.Name = "presence"
	turnOnTrigger.Action = turnOnAction

	// condition = presence = off && lux <= 30
	turnOnCondition := &SensorCondition{}
	turnOnCondition.Name = "presence"
	turnOnCondition.EqualityOperator = "="
	turnOnCondition.Value = true

	luxCondition := &SensorCondition{}
	luxCondition.Name = "lux"
	luxCondition.EqualityOperator = "<="
	luxCondition.Value = lux

	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, luxCondition)

	return turnOnTrigger
}

func createTriggerDelayTurnOffLightWithPresenceOff(name string, delay time.Duration) *SensorTrigger {
	// action = turn off light
	turnOffAction := &MqttAction{}
	turnOffAction.Friendlyname = name
	turnOffAction.Type = "light"
	turnOffAction.Property = "state"
	turnOffAction.Value = false
	turnOffAction.Delay = delay

	// Turn off sensor trigger
	turnOffTrigger := &SensorTrigger{}
	turnOffTrigger.Name = "presence"
	turnOffTrigger.Action = turnOffAction

	// condition = presence == false
	turnOffCondition := &SensorCondition{}
	turnOffCondition.Name = "presence"
	turnOffCondition.EqualityOperator = "="
	turnOffCondition.Value = false

	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffCondition)

	return turnOffTrigger
}
