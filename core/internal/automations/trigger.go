package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
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

var EqualityOperators = map[string]func(any, any) bool{
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

type Condition struct {
	Name             string `json:"name"`
	Value            any    `json:"value"`
	EqualityOperator string `json:"equality"`
}

func (s *Condition) EvaluateV2(exposes map[string]*devices.Entity) bool {

	entity, ok := exposes[s.Name]
	if !ok {
		utils.LogDebugf("sensor %s not found in payload", s.Name)
		return false
	}

	if EqualityOperators[s.EqualityOperator](entity.Data, s.Value) {
		return true
	}

	return false
}

type Trigger struct {
	Name       string       `json:"name"`
	Conditions []*Condition `json:"conditions"`
	Action     *MqttAction  `json:"action"`
}

func (trigger *Trigger) processV2(ctx *DeviceContext) {

	currValue := ctx.GetCurrentV2(trigger.Name)
	for _, c := range trigger.Conditions {

		isMatched := c.EvaluateV2(ctx.Payload)
		if !isMatched {
			trigger.Action.Stop()
			return
		}

		// avoid calling action again for current trigger if value hasnt changed
		if trigger.Name == c.Name && currValue == c.Value {
			return
		}
	}

	trigger.Action.ExecuteV2(trigger.Name, ctx)
}

type MqttAction struct {
	Id           string          `json:"id"`
	FriendlyName string          `json:"friendlyname"`
	Type         string          `json:"type"`
	Property     string          `json:"property"`
	Data         any             `json:"data,omitempty"`
	Client       mqtt.MqttClient `json:"-"`
	Delay        time.Duration   `json:"delay,omitempty"`

	mut       sync.RWMutex
	exit      chan bool
	isPending bool
}

func NewAction() *MqttAction {
	return &MqttAction{Delay: 0}
}

func (a *MqttAction) ExecuteV2(name string, ctx *DeviceContext) {
	// no delay execution
	if a.Delay == 0 {

		payload := a.buildPayloadV2(name, ctx)

		a.emit(payload)

		// on success callback
		// update sensor current value
		ctx.SetCurrentV2(name, ctx.Payload[name].Data)

		return
	}

	if a.isPending {
		a.Stop()
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

			payload := a.buildPayloadV2(name, ctx)

			a.emit(payload)
			ctx.SetCurrentV2(name, ctx.Payload[name].Data)

			// on success callback
			// update sensor current value
			utils.LogInfo("timer constraint finished")
			return

		case <-a.exit:

			utils.LogInfo("timer constraint stopped")
			return
		}
	}()
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

func (a *MqttAction) emit(payload []byte) {

	msg := fmt.Sprintf("%s/set", a.FriendlyName)
	a.Client.Publish(msg, payload)

	utils.LogInfof("Action triggered. Message %s published in %s", string(payload), a.FriendlyName)
}

func (a *MqttAction) buildPayloadV2(name string, ctx *DeviceContext) []byte {

	payloadData := a.Data
	if payloadData == nil {
		payloadData = ctx.Payload[name].Data
	}

	jp := map[string]any{
		a.Property: payloadData,
	}

	payload, _ := json.Marshal(jp)
	return payload

}
