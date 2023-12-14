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

var numericSteps = []int{0, 1, 2}
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

	Client    mqtt.MqttClient          `json:"-"`
	registrar services.DeviceRegistrar `json:"-"`
	mut       sync.RWMutex
	exit      chan bool
	isPending bool
	device    *devices.Device
	max       float64
	min       float64
}

func NewAction() *MqttAction {
	return &MqttAction{Delay: 0, registrar: &mocks.NopDeviceRegistrar{}}
}

func (a *MqttAction) SetRegistrar(registrar services.DeviceRegistrar) {
	a.registrar = registrar

	_, err := a.loadDevice()
	if err != nil {
		utils.LogInfof(err.Error())
	}
}

func (a *MqttAction) Execute(name string, ctx *DeviceContext) {
	// no delay execution
	if a.Delay == 0 {

		payload, err := a.buildPayload(name, ctx)
		if err != nil {
			utils.LogErrorf(err.Error())
			return
		}

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

		if val, ok := a.device.Exposes[a.Property].Attributes["max"]; ok {
			a.max = val.(float64)
		}

		if val, ok := a.device.Exposes[a.Property].Attributes["min"]; ok {
			a.min = val.(float64)
		}
	}

	return a.device, nil
}
func (a *MqttAction) buildPayload(name string, ctx *DeviceContext) ([]byte, error) {

	payloadData := a.Data
	if payloadData == nil {
		payloadData = ctx.Payload[name].Data
	}

	if a.Step > 0 {

		device, err := a.loadDevice()
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("error building action payload"))
		}

		// TODO:
		// maybe we need some ActionStepCalculator
		// or somethig more generic for any type?

		// [1] = + , max
		// [2] = - , min
		// [0] = string

		var newValue = a.Data.(float64)
		if a.Step == 1 { // 1 == increase

			if expose, ok := device.Exposes[a.Property]; ok {
				if expose.Data != nil {
					newValue = expose.Data.(float64) + a.Data.(float64) // calculation
					if a.max != 0 {
						newValue = math.Min(newValue, a.max)
					}
				}
			}
		} else if a.Step == 2 { // 2 == decrease
			if expose, ok := device.Exposes[a.Property]; ok {
				if expose.Data != nil {
					newValue = expose.Data.(float64) - a.Data.(float64) // calculation
					if a.min != 0 {
						newValue = math.Max(newValue, a.min)
					}
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
