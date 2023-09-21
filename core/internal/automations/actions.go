package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
	"node-herder/utils"
	"sync"
	"time"
)

type Action interface {
	Run() error
}

type MqttAction struct {
	Friendlyname string `json:"friendlyname"`

	Type     string          `json:"type"`
	Property string          `json:"name"`
	Value    any             `json:"value"`
	Client   mqtt.MqttClient `json:"-"`
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

	utils.LogInfof("Action triggered. Message %s published in %s", string(payload), a.Friendlyname)

	return nil
}

type ActionRunner struct {
	Friendlyname string          `json:"friendlyname"`
	Type         string          `json:"type"`
	Property     string          `json:"name"`
	Value        any             `json:"value"`
	Client       mqtt.MqttClient `json:"-"`

	Delay     time.Duration `json:"delay"`
	mut       sync.RWMutex
	exit      chan bool
	isPending bool
}

func NewActionRunner() *ActionRunner {
	return &ActionRunner{Delay: 0}
}

func (a *ActionRunner) Stop() {
	if a.isPending {
		// stop it and exit
		a.mut.Lock()
		defer a.mut.Unlock()

		a.exit <- true
		a.isPending = false
	}
}

func (a *ActionRunner) Execute(onSuccess func()) {

	if a.isPending {
		a.Stop()
		return
	}

	// no delay execution
	if a.Delay == 0 {
		a.run()
		onSuccess()

		return
	}

	a.mut.Lock()
	defer a.mut.Unlock()

	// with delay execution
	a.exit = make(chan bool, 1)
	go func() {

		utils.LogInfo("time constraint started")

		timestamp := time.Now().Add(a.Delay)
		diff := time.Until(timestamp).Milliseconds()

		duration := time.Duration(diff)
		ticker := *time.NewTicker(duration * time.Millisecond)
		a.isPending = true

		defer func() {
			close(a.exit)
			a.isPending = false
		}()

		select {
		case <-ticker.C:
			a.run()
			onSuccess()

			utils.LogInfo("timer constraint finished")
			return

		case <-a.exit:

			utils.LogInfo("timer constraint stopped")
			return
		}
	}()

}

func (a *ActionRunner) run() {

	jp := map[string]any{
		a.Property: a.Value,
	}
	payload, _ := json.Marshal(jp)

	msg := fmt.Sprintf("%s/set", a.Friendlyname)
	a.Client.Publish(msg, payload)

	utils.LogInfof("Action triggered. Message %s published in %s", string(payload), a.Friendlyname)
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
	triggers := newMockMqttTriggerPresenseWithLux(false)
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

			utils.LogInfof("Loaded MqttTrigger %s, Enabled=%t", mqttTrigger.Description, mqttTrigger.Enabled)
			continue
		}

		return errors.New("unsupported trigger type ")
	}

	a.configured = true
	return nil
}

// REMOVE
// used for testing only!!!!!!!!!!
func newMockMqttTriggerPresenseWithLux(enabled bool) []interface{} {

	deviceName := "Human presence"
	luxSensor := "illuminance_lux"
	conditionType := "presence"
	offConditionValue := false
	onConditionValue := true
	luxConstraintValue := 210
	constraintOp := "<"
	timerValue := 1 * time.Minute

	trigger := newMqttTrigger()
	trigger.DeviceName = deviceName
	trigger.Description = fmt.Sprintf("automation for device %s", trigger.DeviceName)
	trigger.Enabled = enabled

	// sensor off condition
	offCondition := createConditionWithTimerConstraint("Attic light", "light", conditionType, offConditionValue, "state", false, timerValue, nil)
	trigger.Conditions = append(trigger.Conditions, offCondition)

	// sensor on condition
	turnOnCondition := createTriggerConditionWithSensorConstraint("Attic light", "light", conditionType, onConditionValue, "state", true, luxSensor, luxConstraintValue, constraintOp, nil)
	trigger.Conditions = append(trigger.Conditions, turnOnCondition)

	return []interface{}{trigger}
}

func createTriggerConditionWithSensorConstraint(actionDeviceName string, actionType string, sensor string, conditionValue any, actionProperty string, actionValue any, constraintProperty string, constraintValue any, constraintOp string, mqtt mqtt.MqttClient) *DeviceCondition {

	// create condition
	condition := newMockMqttCondition(sensor, conditionValue)
	action := newMockMqttAction(actionDeviceName, actionProperty, actionType, actionValue)

	action.Client = mqtt
	condition.Action = action

	//
	// add device constrains
	deviceConstraint := NewDeviceConstraint()
	deviceConstraint.Sensor = constraintProperty
	deviceConstraint.Value = constraintValue
	deviceConstraint.EqualityOperator = constraintOp
	condition.Constraint = deviceConstraint
	return condition
}
func createConditionWithTimerConstraint(actionDeviceName string, actionType string, sensor string, conditionValue any, actionProperty string, actionValue any, constraintDuration time.Duration, mqtt mqtt.MqttClient) *DeviceCondition {

	// create condition
	condition := newMockMqttCondition(sensor, conditionValue)
	action := newMockMqttAction(actionDeviceName, actionProperty, actionType, actionValue)

	action.Client = mqtt
	condition.Action = action

	//
	// add timer constrains
	timerConstraint := NewTimerConstraint()
	timerConstraint.Duration = constraintDuration
	condition.Constraint = timerConstraint

	return condition
}
func newMockMqttAction(friendlyName string, property string, actionType string, value any) *MqttAction {
	action := &MqttAction{}
	action.Friendlyname = friendlyName
	action.Property = property
	action.Type = actionType
	action.Value = value

	return action
}

func newMockMqttCondition(sensor string, value any) *DeviceCondition {
	condition := NewDeviceCondition()
	condition.Friendlyname = sensor
	condition.Type = sensor
	condition.Value = value
	return condition
}
