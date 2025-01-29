package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/utils"
	"reflect"
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

//////////////////////////////////////////////////////////////

type MqttTriggerActionExpose struct {
	Name string `json:"name"`
	Data any    `json:"data,omitempty"`
}

type MqttTriggerAction struct {
	MqttBaseAction
	Exposes   []*MqttTriggerActionExpose `json:"exposes"`
	Delay     *utils.TimeInterval        `json:"delay"`
	operation actionOperation
}

func NewTriggerAction() *MqttTriggerAction {
	return &MqttTriggerAction{
		Exposes: make([]*MqttTriggerActionExpose, 0),
		Delay:   utils.IntervalFromMilliseconds(0),
		MqttBaseAction: MqttBaseAction{
			Type: TriggerAction,
		},
	}
}

func (a *MqttTriggerAction) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {
	bridgeInfo, err := registrar.FindBridgeInfo(a.Id)
	if err != nil {
		return err
	}
	for _, property := range a.Exposes {
		sanitizedData, err := bridgeInfo.SanitiseProperty(property.Name, property.Data)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to sanitize data for action %s: %s", a.Id, err.Error()))
		}
		property.Data = sanitizedData
	}

	a.operation = CreateTriggerOperation(a)

	// common logic
	device, err := registrar.LookupById(a.Id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %s ", a.Id, err.Error())
	}
	a.Id = device.Id
	// NOTE: if device is renamed we might need to register the automations again
	a.friendlyName = device.FriendlyName
	a.Client = client
	a.registrar = registrar
	return nil
}

func (a *MqttTriggerAction) Execute(ctx *DeviceContext) error {
	a.mut.Lock()
	defer a.mut.Unlock()

	if a.isPending {
		return nil
	}

	// no delay execution
	if a.Delay.Value == 0 {
		defer func() {
			a.isPending = false
		}()

		a.isPending = true

		payload, err := a.operation.CreatePayload()
		if err != nil {
			return err
		}
		bytes, _ := json.Marshal(payload)
		a.emit(bytes)

		// on success callback
		// update sensor current value so we dont have to query the device again
		for key, value := range payload {
			ctx.SetCurrent(key, value)
		}

		return nil
	}

	// with delay execution
	a.exit = make(chan bool, 1)
	go func() {

		var delay = a.Delay.Duration()
		timestamp := time.Now().Add(delay)
		diff := time.Until(timestamp).Milliseconds()

		duration := time.Duration(diff)
		ticker := *time.NewTicker(duration * time.Millisecond)
		a.isPending = true
		//utils.LogInfof("time constraint started Delay: %d ms", a.Delay)

		defer func() {
			close(a.exit)
			a.isPending = false
		}()

		select {
		case <-ticker.C:

			payload, err := a.operation.CreatePayload()
			if err != nil {
				return
			}

			bytes, _ := json.Marshal(payload)
			a.emit(bytes)

			// on success callback
			// update sensor current value so we dont have to query the device again
			for key, value := range payload {
				ctx.SetCurrent(key, value)
			}
			//utils.LogInfo("timer constraint finished")
			return

		case <-a.exit:

			//utils.LogInfo("timer constraint stopped")
			return
		}
	}()
	return nil
}

type MqttStepAction struct {
	MqttBaseAction
	Property  string  `json:"property"`
	Steps     []*Step `json:"steps,omitempty"`
	Data      any     `json:"data,omitempty"`
	operation actionOperation
}

func NewStepAction() *MqttStepAction {
	return &MqttStepAction{
		Steps: make([]*Step, 0),
		MqttBaseAction: MqttBaseAction{
			Type: StepAction,
		},
	}
}

func (a *MqttStepAction) Execute(ctx *DeviceContext) error {

	a.mut.Lock()
	defer a.mut.Unlock()

	if a.isPending {
		return nil
	}

	defer func() {
		a.isPending = false
	}()

	a.isPending = true
	payload, err := a.operation.CreatePayload()
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		utils.LogErrorf("step action failed %s ", err.Error())
		return err
	}

	a.emit(bytes)

	for key, value := range payload {
		ctx.SetCurrent(key, value)
	}

	return nil
}
func (a *MqttStepAction) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

	device, err := registrar.LookupById(a.Id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %s ", a.Id, err.Error())
	}
	expose := device.Exposes[a.Property]
	a.operation = CreateStepOperation(expose, a)

	// common logic
	a.Id = device.Id
	a.friendlyName = device.FriendlyName
	a.Client = client
	a.registrar = registrar

	return nil
}

type MqttPresetCyclingAction struct {
	MqttBaseAction
	Property  string   `json:"property"`
	Presets   []string `json:"presets,omitempty"`
	operation actionOperation
}

func NewPresetCyclingAction() *MqttPresetCyclingAction {
	return &MqttPresetCyclingAction{
		Presets: make([]string, 0),
		MqttBaseAction: MqttBaseAction{
			Type: PresetRotationAction,
		},
	}
}

func (a *MqttPresetCyclingAction) Execute(ctx *DeviceContext) error {
	a.mut.Lock()
	defer a.mut.Unlock()

	if a.isPending {
		return nil
	}

	defer func() {
		a.isPending = false
	}()

	a.isPending = true
	payload, err := a.operation.CreatePayload()
	if err != nil {
		return err
	}

	// build payload
	bytes, err := json.Marshal(payload)
	if err != nil {
		utils.LogErrorf("preset cycling action failed %s ", err.Error())
		return err
	}

	a.emit(bytes)

	for key, value := range payload {
		ctx.SetCurrent(key, value)
	}
	return nil
}

func (a *MqttPresetCyclingAction) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

	device, err := registrar.LookupById(a.Id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %s ", a.Id, err.Error())
	}

	expose := device.Exposes[a.Property]
	a.operation = CreateRotateOperation(expose)

	// common logic
	a.Id = device.Id
	a.friendlyName = device.FriendlyName
	a.Client = client
	a.registrar = registrar
	return nil
}

type MqttBaseAction struct {
	Id           string                   `json:"id"`
	Type         string                   `json:"type"`
	Client       mqtt.MqttClient          `json:"-"`
	registrar    services.DeviceRegistrar `json:"-"`
	mut          sync.RWMutex             `json:"-"`
	exit         chan bool                `json:"-"`
	isPending    bool                     `json:"-"`
	friendlyName string                   `json:"-"`
}

func (a *MqttBaseAction) emit(payload []byte) {

	msg := fmt.Sprintf("%s/set", a.friendlyName)
	a.Client.Publish(msg, payload)

	utils.LogInfof("Action triggered. Message %s published in %s", string(payload), a.friendlyName)
}

func (b *MqttBaseAction) GetID() string {
	return b.Id
}

func (b *MqttBaseAction) GetType() string {
	return b.Type
}

func (a *MqttBaseAction) Stop() {
	if a.isPending {
		// stop it and exit
		a.mut.Lock()
		defer a.mut.Unlock()

		a.exit <- true
		a.isPending = false
	}
}

func (b *MqttBaseAction) ExecuteBase(ctx *DeviceContext) error {
	// IMPORTANT !!!!
	//we need to check if value is still the same and not emit anything new

	// b.mut.Lock()
	// defer b.mut.Unlock()

	// if b.isPending {
	// 		return fmt.Errorf("action %s is already pending", b.Id) // Return an error
	// }

	// b.isPending = true
	// defer func() { b.isPending = false }() // Ensure isPending is reset

	return nil
}

type MqttAction interface {
	Execute(tx *DeviceContext) error
	Stop()
	GetID() string
	GetType() string
	Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error
}

var typeRegistry = map[string]reflect.Type{
	TriggerAction:        reflect.TypeOf(MqttTriggerAction{}),
	StepAction:           reflect.TypeOf(MqttStepAction{}),
	PresetRotationAction: reflect.TypeOf(MqttPresetCyclingAction{}),
}

// ////////////////////////////////////////////////////////////
// type MqttAction struct {
// 	Id           string `json:"id"`
// 	FriendlyName string `json:"friendlyname"`
// 	Property     string `json:"property"`
// 	Type         string `json:"type"`
// 	Data         any    `json:"data,omitempty"`
// 	Delay        int    `json:"delay,omitempty"`
// 	Steps        []Step `json:"steps,omitempty"`

// 	Client    mqtt.MqttClient          `json:"-"`
// 	registrar services.DeviceRegistrar `json:"-"`

// 	operationAction actionOperation `json:"-"`
// 	mut             sync.RWMutex
// 	exit            chan bool
// 	isPending       bool
// }

// func NewAction() *MqttAction {
// 	return &MqttAction{Delay: 0, Steps: make([]Step, 0)}
// }

// func (a *MqttAction) Execute(name string, ctx *DeviceContext) {

// 	a.mut.Lock()
// 	defer a.mut.Unlock()

// 	if a.isPending {
// 		return
// 	}

// 	// no delay execution
// 	if a.Delay == 0 {
// 		defer func() {
// 			a.isPending = false
// 		}()

// 		a.isPending = true
// 		payload, err := a.buildPayload(name, ctx)
// 		if err != nil {
// 			utils.LogErrorf("build payload failed: %s", err.Error())
// 			return
// 		}

// 		a.emit(payload)

// 		// on success callback
// 		// update sensor current value
// 		ctx.SetCurrent(name, ctx.Payload[name].Data)
// 		return
// 	}

// 	// with delay execution
// 	a.exit = make(chan bool, 1)
// 	go func() {

// 		var delay = time.Duration(float64(a.Delay) * float64(time.Millisecond))
// 		timestamp := time.Now().Add(delay)
// 		diff := time.Until(timestamp).Milliseconds()

// 		duration := time.Duration(diff)
// 		ticker := *time.NewTicker(duration * time.Millisecond)
// 		a.isPending = true
// 		//utils.LogInfof("time constraint started Delay: %d ms", a.Delay)

// 		defer func() {
// 			close(a.exit)
// 			a.isPending = false
// 		}()

// 		select {
// 		case <-ticker.C:

// 			payload, err := a.buildPayload(name, ctx)
// 			if err != nil {
// 				utils.LogErrorf(err.Error())
// 				return
// 			}

// 			a.emit(payload)
// 			ctx.SetCurrent(name, ctx.Payload[name].Data)

// 			// on success callback
// 			// update sensor current value
// 			//utils.LogInfo("timer constraint finished")
// 			return

// 		case <-a.exit:

// 			//utils.LogInfo("timer constraint stopped")
// 			return
// 		}
// 	}()
// }

// func (a *MqttAction) Stop() {
// 	if a.isPending {
// 		// stop it and exit
// 		a.mut.Lock()
// 		defer a.mut.Unlock()

// 		a.exit <- true
// 		a.isPending = false
// 	}
// }

// func (a *MqttAction) emit(payload []byte) {

// 	msg := fmt.Sprintf("%s/set", a.FriendlyName)
// 	a.Client.Publish(msg, payload)

// 	utils.LogInfof("Action triggered. Message %s published in %s", string(payload), a.FriendlyName)
// }

// func (a *MqttAction) buildPayload(name string, ctx *DeviceContext) ([]byte, error) {

// 	if a.operationAction != nil {
// 		newValue, err := a.operationAction.Next()
// 		if err != nil {
// 			return nil, err
// 		}

// 		return createJson(a.Property, newValue), nil
// 	}

// 	payloadData := a.Data
// 	if payloadData == nil {
// 		payloadData = ctx.Payload[name].Data
// 	}
// 	return createJson(a.Property, payloadData), nil
// }

// func createJson(property string, data any) []byte {
// 	jp := map[string]any{
// 		property: data,
// 	}

// 	payload, _ := json.Marshal(jp)
// 	return payload
// }
