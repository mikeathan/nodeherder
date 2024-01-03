package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/mocks"
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

type numericOperator struct {
	Operator string
	Limit    string
}

var stepsOperators = map[int]*numericOperator{
	1: {Operator: "+", Limit: "max"},
	2: {Operator: "-", Limit: "min"},
}

var numericOperations = map[string]func(float64, float64, float64) float64{
	"+": func(v1 float64, v2 float64, limit float64) float64 {

		newValue := v1 + v2
		if limit != 0 {
			newValue = math.Min(newValue, limit)
		}
		return newValue
	},
	"-": func(v1 float64, v2 float64, limit float64) float64 {

		newValue := v1 - v2
		newValue = math.Max(newValue, limit)

		return newValue
	},
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
	Type         string `json:"type"`
	Property     string `json:"property"`
	Data         any    `json:"data,omitempty"`
	Delay        int    `json:"delay,omitempty"`
	Step         int    `json:"step,omitempty"`
	PresetRotate bool   `json:"preset_rotate,omitempty"`

	Client    mqtt.MqttClient          `json:"-"`
	registrar services.DeviceRegistrar `json:"-"`
	mut       sync.RWMutex
	exit      chan bool
	isPending bool

	device    *devices.Device
	limits    map[string]float64
	presets   []any
	presetPos int
}

func NewAction() *MqttAction {
	return &MqttAction{Delay: 0, registrar: &mocks.NopDeviceRegistrar{}, limits: make(map[string]float64)}
}

func (a *MqttAction) SetRegistrar(registrar services.DeviceRegistrar) {
	a.registrar = registrar

	_, err := a.loadDevice()
	if err != nil {
		utils.LogInfof(err.Error())
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

func (a *MqttAction) loadDevice() (*devices.Device, error) {
	if a.device == nil {
		var err error
		a.device, err = a.registrar.LookupById(a.Id)
		if err != nil {
			return nil, fmt.Errorf("Device %s not found for action %s", a.Id, a.FriendlyName)
		}

		a.limits = map[string]float64{}
		if val, ok := a.device.Exposes[a.Property].Attributes["max"]; ok {
			a.limits["max"] = val.(float64)
		}

		if val, ok := a.device.Exposes[a.Property].Attributes["min"]; ok {
			a.limits["min"] = val.(float64)
		}

		if a.PresetRotate {
			for _, value := range a.device.Exposes[a.Property].Presets {
				a.presets = append(a.presets, value)
			}
		}
	}

	return a.device, nil
}

func (a *MqttAction) buildPayload(name string, ctx *DeviceContext) ([]byte, error) {

	payloadData := a.Data
	if payloadData == nil {
		payloadData = ctx.Payload[name].Data
	}

	// TODO:
	// if we have a a preset_cycle flag
	// then we can move to next preset to get value
	// increment the index and store it
	if a.PresetRotate {
		_, err := a.loadDevice()
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("error building action payload"))
		}

		v := a.presets[a.presetPos]
	}

	if a.Step > 0 {

		device, err := a.loadDevice()
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("error building action payload"))
		}
		// value is modified by a step up/down value
		// it has to be within max/min limits

		// [1] = + , max
		// [2] = - , min
		op := stepsOperators[a.Step]

		var newValue = a.Data.(float64)
		if expose, ok := device.Exposes[a.Property]; ok {
			if expose.Data != nil {
				limit := a.limits[op.Limit]
				newValue = numericOperations[op.Operator](expose.Data.(float64), a.Data.(float64), limit)

				if expose.Data == newValue {
					return nil, errors.New("same value, skipping")
				}
			}
		}

		payloadData = newValue
	}

	jp := map[string]any{
		a.Property: payloadData,
	}

	payload, _ := json.Marshal(jp)

	return payload, nil

}
