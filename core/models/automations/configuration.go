package automations

import (
	"errors"
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
	"time"
)

type AutomationService struct {
	mqttTriggers map[string]*MqttTrigger
	timeTriggers map[string]*TimerTrigger
	mqttClient   mqtt.MqttClient
}

func Create(mqtt mqtt.MqttClient) *AutomationService {
	return &AutomationService{
		mqttTriggers: map[string]*MqttTrigger{},
		timeTriggers: map[string]*TimerTrigger{},
		mqttClient:   mqtt,
	}
}
func (a *AutomationService) HandleDevice(device *devices.Device) {
	if t, ok := a.mqttTriggers[device.Id]; ok {
		t.Evaluate(device)
		return
	}
}

func (a *AutomationService) Load(bridgeDevices []*devices.BridgeDevice) error {

	// fake input data
	triggers := createMockConfiguration()
	//

	for _, trigger := range triggers {

		timerTrigger, ok := trigger.(*TimerTrigger)
		if ok {
			err := timerTrigger.configure(bridgeDevices, a.mqttClient)
			if err != nil {
				return err
			}
			a.timeTriggers[timerTrigger.Name] = timerTrigger
			continue
		}

		mqttTrigger, ok := trigger.(*MqttTrigger)
		if ok {
			err := mqttTrigger.configure(bridgeDevices, a.mqttClient)
			if err != nil {
				return err
			}

			a.mqttTriggers[mqttTrigger.Name] = mqttTrigger
			continue
		}

		return errors.New("unsupported trigger type ")
	}

	return nil
}

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
