package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/utils"
	"sync"
	"time"
)

// brightness + direction_time * value = [expose_name] [+/-] [expose_name] [numeric_operator] [numeric value]
// brightness - direction_time * value

// action set brighness +/- some value = [value_source] [arithmetic operator] [step_value]
// action set brighness  +/- some other numeric combination  eg direction_time * 0.5

const (
	TriggerAction        = "TriggerAction"
	StepAction           = "StepAction"
	PresetRotationAction = "PresetRotationAction"
)

type Step struct {
	Property string `json:"property"`
	Operator string `json:"operator"` // +,-, *,/
	Id       string `json:"id"`
}

type MqttAction struct {
	Id           string `json:"id"`
	FriendlyName string `json:"friendlyname"`
	Property     string `json:"property"`
	Type         string `json:"type"`
	Data         any    `json:"data,omitempty"`
	Delay        int    `json:"delay,omitempty"`
	Steps        []Step `json:"steps,omitempty"`

	Client    mqtt.MqttClient          `json:"-"`
	registrar services.DeviceRegistrar `json:"-"`

	operationAction actionOperation `json:"-"`
	mut             sync.RWMutex
	exit            chan bool
	isPending       bool
}

func NewAction() *MqttAction {
	return &MqttAction{Delay: 0, Steps: make([]Step, 0)}
}

func (a *MqttAction) configure(registrar services.DeviceRegistrar) error {

	a.registrar = registrar
	device, err := registrar.LookupById(a.Id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %s ", a.Id, err.Error())
	}

	expose := device.Exposes[a.Property]

	// configure special action operations
	switch a.Type {
	case TriggerAction:
		// nothing to do here
		break
	case StepAction:
		a.operationAction = CreateStepOperation(expose, a)
	case PresetRotationAction:
		a.operationAction = CreateRotateOperation(expose, a)
	}
	return nil
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

	if a.operationAction != nil {
		newValue, err := a.operationAction.Next(ctx)
		if err != nil {
			return nil, err
		}

		return createJson(a.Property, newValue), nil
	}

	payloadData := a.Data
	if payloadData == nil {
		payloadData = ctx.Payload[name].Data
	}
	return createJson(a.Property, payloadData), nil
}

func createJson(property string, data any) []byte {
	jp := map[string]any{
		property: data,
	}

	payload, _ := json.Marshal(jp)
	return payload
}
