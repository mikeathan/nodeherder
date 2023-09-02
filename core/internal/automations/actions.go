package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
	"node-herder/utils"
)

type MqttAction struct {
	Friendlyname string `json:"friendlyname"`

	Type     string `json:"type"`
	Property string `json:"name"`
	Value    any    `json:"value"`
	Client   mqtt.MqttClient
}

func (a *MqttAction) Run() error {

	jp := map[string]any{
		a.Property: a.Value,
	}
	payload, err := json.Marshal(jp)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("%s/set", a.Friendlyname)
	a.Client.Publish(msg, payload)

	//fmt.Printf("[DEBUG] publishing topic: %s => %s \n", msg, string(payload))
	return nil
}

type Engine interface { // TODO: might need to move it to Models????
	HandleDevice(id string, data map[string]any)
	Load(bridgeDevices []*devices.BridgeDevice) error
}

type AutomationEngine struct {
	mqttTriggers map[string]*MqttTrigger
	mqttClient   mqtt.MqttClient
	configured   bool
}

func NewEngine(mqtt mqtt.MqttClient) *AutomationEngine {
	return &AutomationEngine{
		mqttTriggers: map[string]*MqttTrigger{},
		mqttClient:   mqtt,
	}
}
func (a *AutomationEngine) HandleDevice(id string, data map[string]any) {
	if t, ok := a.mqttTriggers[id]; ok {
		t.Evaluate(data)
	}
}

func (a *AutomationEngine) Load(bridgeDevices []*devices.BridgeDevice) error {

	// fake input data - TEST ONLY
	// tha needs to come from a file and loaded
	triggers := newMockMqttTrigger()
	//

	utils.LogInfof("Loading automations")
	for _, trigger := range triggers {

		mqttTrigger, ok := trigger.(*MqttTrigger)
		if ok {
			err := mqttTrigger.configure(bridgeDevices, a.mqttClient)
			if err != nil {
				return err
			}

			a.mqttTriggers[mqttTrigger.DeviceName] = mqttTrigger
			utils.LogInfof("MqttTrigger %s loaded", mqttTrigger.Description)
			continue
		}

		return errors.New("unsupported trigger type ")
	}

	a.configured = true
	return nil
}

// REMOVE
// used for testing only!!!!!!!!!!
func newMockMqttTrigger() []interface{} {

	trigger := newMqttTrigger()
	trigger.DeviceName = "Human presence"
	trigger.Description = "Turn on light 1 mqtt automation"
	trigger.Enabled = false // disabled

	// new condition
	mcOn := &DeviceCondition{}
	mcOn.Friendlyname = "Human presence" // ? we dont need that now
	mcOn.Type = "presence"
	mcOn.Value = true

	// new action
	ma := &MqttAction{}
	ma.Client = nil

	ma.Friendlyname = "Attic light"
	ma.Property = "state"
	ma.Type = "light"
	ma.Value = true
	mcOn.Action = ma

	// ##########################
	mcOff := &DeviceCondition{}
	mcOff.Friendlyname = "Human presence"
	mcOff.Type = "presence"
	mcOff.Value = false

	// TODO:
	// add timer condition

	//example sensor says falss and start timer, after eg 15 min call action to turn off light

	// new action
	ma2 := &MqttAction{}
	ma2.Client = nil

	ma2.Friendlyname = "Attic light"
	ma2.Property = "state"
	ma2.Type = "light"
	ma2.Value = false
	mcOff.Action = ma2

	////////////////////////////////////////////////////

	trigger.Conditions = append(trigger.Conditions, mcOn)
	trigger.Conditions = append(trigger.Conditions, mcOff)
	return []interface{}{trigger}
}
