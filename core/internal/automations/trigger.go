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
	Id           string `json:"id"`
	FriendlyName string `json:"friendlyname"`
	Property     string `json:"property"`
	Data         any    `json:"data,omitempty"`
	Delay        int    `json:"delay,omitempty"`
	Step         int    `json:"step,omitempty"`
	PresetRotate bool   `json:"preset_rotate,omitempty"`

	Client    mqtt.MqttClient `json:"-"`
	operation actionOperation
	expose    *devices.Entity
	mut       sync.RWMutex
	exit      chan bool
	isPending bool
}

func NewAction() *MqttAction {
	return &MqttAction{Delay: 0}
}

func (a *MqttAction) configure(expose *devices.Entity) {
	a.expose = expose

	if a.Step > 0 {
		var limit float64
		op := stepsOperators[a.Step]
		if val, ok := expose.Attributes[op.Limit]; ok {
			limit = val.(float64)
		}

		a.operation = newStepOperation(a.Step, a.Data.(float64), limit)
	} else if a.PresetRotate {

		var presets []any
		for _, value := range expose.Presets {
			presets = append(presets, value)
		}

		a.operation = newRotateOperation(presets)
	}
}

func (a *MqttAction) Execute(name string, ctx *DeviceContext) {

	a.mut.Lock()
	defer a.mut.Unlock()

	if a.isPending {
		return
	}

	// no delay execution
	if a.Delay == 0 {
		defer func() {
			a.isPending = false
		}()

		a.isPending = true
		payload, err := a.buildPayload(name, ctx)
		if err != nil {
			return
		}

		a.emit(payload)

		// on success callback
		// update sensor current value
		ctx.SetCurrent(name, ctx.Payload[name].Data)
		return
	}

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

			payload, err := a.buildPayload(name, ctx)
			if err != nil {
				utils.LogErrorf(err.Error())
				return
			}

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

func (a *MqttAction) buildPayload(name string, ctx *DeviceContext) ([]byte, error) {

	payloadData := a.Data
	if payloadData == nil {
		payloadData = ctx.Payload[name].Data
	}

	if a.Step > 0 {
		newValue, err := a.operation.Next(a.expose.Data.(float64))
		if err != nil {
			return nil, err
		}
		payloadData = newValue
	} else if a.PresetRotate {
		newValue, err := a.operation.Next(0)
		if err != nil {
			return nil, err
		}
		payloadData = newValue
	}

	jp := map[string]any{
		a.Property: payloadData,
	}

	payload, _ := json.Marshal(jp)

	return payload, nil

}
