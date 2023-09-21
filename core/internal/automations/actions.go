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

type MqttAction struct {
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

func NewActionRunner() *MqttAction {
	return &MqttAction{Delay: 0}
}

func (a *MqttAction) Stop() {
	if a.isPending {
		// stop it and exit
		a.mut.Lock()
		defer a.mut.Unlock()

		a.exit <- true
		a.isPending = false
	}
}

func (a *MqttAction) Execute(onSuccess func()) {

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

func (a *MqttAction) run() {

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
	triggers := newMockMqttTriggerPresenseWithLux(false)
	//

	utils.LogInfof("Loading automations")
	for _, trigger := range triggers {

		deviceTrigger, ok := trigger.(*DeviceTrigger)
		if ok {
			err := deviceTrigger.configure(bridgeDevices, a.mqttClient)
			if err != nil {
				return err
			}

			a.deviceTriggers[deviceTrigger.Name] = deviceTrigger

			utils.LogInfof("Loaded MqttTrigger %s, Enabled=%t", deviceTrigger.Description, deviceTrigger.Enabled)
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

	return []interface{}{}
}
