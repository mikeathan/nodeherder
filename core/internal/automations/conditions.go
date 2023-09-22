package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/utils"
	"sync"
	"time"
)

func toFloat(value any) float32 {
	switch v := value.(type) {
	case int:
		return float32(v)
	case float64:
		return float32(v)
	case float32:
		return float32(v)
	default:
		return float32(0)
	}
}

var Equalityoperators = map[string]func(any, any) bool{
	"=": func(v1 any, v2 any) bool {
		return v1 == v2
	},
	">=": func(v1 any, v2 any) bool {
		return toFloat(v1) >= toFloat(v2)
	},
	"<=": func(v1 any, v2 any) bool {
		return toFloat(v1) <= toFloat(v2)
	},
	">": func(v1 any, v2 any) bool {
		return toFloat(v1) > toFloat(v2)
	},
	"<": func(v1 any, v2 any) bool {
		return toFloat(v1) < toFloat(v2)
	},
}

type SensorCondition struct {
	Name             string `json:"name"`
	Value            any    `json:"value"`
	EqualityOperator string `json:"equalityoperator"`
}

func (s *SensorCondition) Evaluate(data map[string]any) bool {

	value, ok := data[s.Name]
	if !ok {
		utils.LogDebugf("sensor %s not found in payload", s.Name)
		return false
	}

	if Equalityoperators[s.EqualityOperator](value, s.Value) {
		return true
	}

	return false
}

type SensorTrigger struct {
	Name       string             `json:"name"`
	Conditions []*SensorCondition `json:"conditions"`
	Action     *MqttAction        `json:"action"`
}

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

func NewAction() *MqttAction {
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
