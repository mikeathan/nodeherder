package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
	"node-herder/utils"
	"time"
)

type MqttAction struct {
	Friendlyname string `json:"friendlyname"`
	Type         string `json:"type"`
	Property     string `json:"name"`
	Value        any    `json:"value"`
	client       mqtt.MqttClient
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
	a.client.Publish(msg, payload)
	return nil
}

type Engine interface { // TODO: might need to move it to Models????
	HandleDevice(device *devices.Device)
	Load(bridgeDevices []*devices.BridgeDevice) error
}

type AutomationEngine struct {
	mqttTriggers map[string]*MqttTrigger
	timeTriggers map[string]*TimerTrigger
	mqttClient   mqtt.MqttClient
	configured   bool
}

func NewEngine(mqtt mqtt.MqttClient) *AutomationEngine {
	return &AutomationEngine{
		mqttTriggers: map[string]*MqttTrigger{},
		timeTriggers: map[string]*TimerTrigger{},
		mqttClient:   mqtt,
	}
}
func (a *AutomationEngine) HandleDevice(device *devices.Device) {
	if t, ok := a.mqttTriggers[device.Id]; ok {
		t.Evaluate(device)
	}
}

func (a *AutomationEngine) Load(bridgeDevices []*devices.BridgeDevice) error {

	// fake input data - TEST ONLY
	// tha needs to come from a file and loaded
	triggers := createMockConfiguration()
	//

	utils.LogInfof("Loading automations")
	for _, trigger := range triggers {

		timerTrigger, ok := trigger.(*TimerTrigger)
		if ok {
			err := timerTrigger.configure(bridgeDevices, a.mqttClient)
			if err != nil {
				return err
			}
			a.timeTriggers[timerTrigger.Name] = timerTrigger
			utils.LogInfof("TimerTrigger %s loaded", timerTrigger.Name)
			continue
		}

		mqttTrigger, ok := trigger.(*MqttTrigger)
		if ok {
			err := mqttTrigger.configure(bridgeDevices, a.mqttClient)
			if err != nil {
				return err
			}

			a.mqttTriggers[mqttTrigger.Name] = mqttTrigger
			utils.LogInfof("MqttTrigger %s loaded", mqttTrigger.Name)
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
	trigger.Name = "Turn on light 1 mqtt automation"
	trigger.Enabled = false

	// new condition
	mcOn := &MqttCondition{}
	mcOn.Friendlyname = "Human presence"
	mcOn.Type = "presence"
	mcOn.Value = true

	// new action
	ma := &MqttAction{}
	ma.client = nil

	ma.Friendlyname = "Attic light"
	ma.Property = "state"
	ma.Type = "light"
	ma.Value = false

	//////////////////////////////////////////////////////////////
	trigger2 := newMqttTrigger()
	trigger2.Name = "Turn off light 1 mqtt automation"
	trigger2.Enabled = false

	mcOff := &MqttCondition{}
	mcOff.Friendlyname = "Human presence"
	mcOff.Type = "presence"
	mcOn.Value = false

	// TODO:
	// add timer condition

	//example sensor says falss and start timer, after eg 15 min call action to turn off light

	// new action
	ma2 := &MqttAction{}
	ma2.client = nil

	ma2.Friendlyname = "Attic light"
	ma2.Property = "state"
	ma2.Type = "light"
	ma2.Value = false

	////////////////////////////////////////////////////

	return []interface{}{trigger}
}

// REMOVE
// used for testing only!!!!!!!!!!
func createMockConfiguration() []interface{} {

	trigger := newTimerTrigger()
	trigger.Name = "Turn on light 1 timer automation"
	trigger.Enabled = false

	// new condition
	tc := &TimeDurationCondition{}
	tc.Duration = 5 * time.Second
	trigger.Condition = tc

	// new action
	ma := &MqttAction{}
	ma.client = nil

	ma.Friendlyname = "Attic light"
	ma.Property = "state"
	ma.Type = "light"
	ma.Value = false
	trigger.Action = ma

	return []interface{}{trigger}
}
