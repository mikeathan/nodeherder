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

var numericSteps = []string{"increase", "decrease"}
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

func (s *Condition) Evaluate(exposes map[string]*devices.Entity) bool {

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

// NOTE:
// handle action type - eg for hue tap switch
// first case press of button (btn1_press, btn1_release)

// logic needs to be in evaluate, maybe have differnt evaluate according to exposed type
// map of evaluators for execution, eg default and action states
// if expose is action and enum execute the stateful action evaluator
// question : how do we access the extra device information for current action
// eg
//
//		"action_direction": "right",
//	    "action_time": 15,
//	    "action_type": "step",
//
// not required for simple case but good to have an idea

// statefull action evaluator
// read action = btn1_press
// compare with automation conditions if it matches

// example of conditions, the need to be in correct sequence
// btn1_press
// btn1_release

// if action matched do we have another one ?
// if no then evalued is true
// if we do have another one
// set current action in context and return false

// if we have a differnt button eg btn2_press we need to reset context guess goes thought same code path
// what about btn1_hold or other eventws coming from same button but should not clear the context ???

func (trigger *Trigger) process(ctx *DeviceContext) {

	currValue := ctx.GetCurrent(trigger.Name)
	for _, c := range trigger.Conditions {

		isMatched := c.Evaluate(ctx.Payload)
		if !isMatched {
			trigger.Action.Stop()
			return
		}

		// avoid calling action again for current trigger if value hasnt changed
		if trigger.Name == c.Name && currValue == c.Value {
			return
		}
	}

	trigger.Action.Execute(trigger.Name, ctx)
}

type MqttAction struct {
	Id           string          `json:"id"`
	FriendlyName string          `json:"friendlyname"`
	Type         string          `json:"type"`
	Property     string          `json:"property"`
	Data         any             `json:"data,omitempty"`
	Delay        int             `json:"delay,omitempty"`
	Client       mqtt.MqttClient `json:"-"`
	Step         string          `json:"step,omitempty"`

	mut       sync.RWMutex
	exit      chan bool
	isPending bool
}

func NewAction() *MqttAction {
	return &MqttAction{Delay: 0}
}

func (a *MqttAction) Execute(name string, ctx *DeviceContext) {
	// no delay execution
	if a.Delay == 0 {

		payload := a.buildPayload(name, ctx)

		a.emit(payload)

		// on success callback
		// update sensor current value
		ctx.SetCurrent(name, ctx.Payload[name].Data)

		return
	}

	if a.isPending {
		return
	}

	a.mut.Lock()
	defer a.mut.Unlock()

	// with delay execution
	a.exit = make(chan bool, 1)
	go func() {

		var delay = time.Duration(float64(a.Delay) * float64(time.Millisecond))
		timestamp := time.Now().Add(delay)
		diff := time.Until(timestamp).Milliseconds()

		duration := time.Duration(diff)
		ticker := *time.NewTicker(duration * time.Millisecond)
		a.isPending = true
		utils.LogInfof("time constraint started Delay: %d ms", a.Delay)

		defer func() {
			close(a.exit)
			a.isPending = false
		}()

		select {
		case <-ticker.C:

			payload := a.buildPayload(name, ctx)

			a.emit(payload)
			ctx.SetCurrent(name, ctx.Payload[name].Data)

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

func (a *MqttAction) buildPayload(name string, ctx *DeviceContext) []byte {

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
