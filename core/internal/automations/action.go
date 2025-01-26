package automations

import (
	"encoding/json"
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

type MqttTrigerAction struct {
	MqttBaseAction
	Exposes []*MqttTriggerActionExpose `json:"exposes"`
	Delay   *utils.TimeInterval        `json:"delay"`
}

func NewTriggerAction() *MqttTrigerAction {
	return &MqttTrigerAction{
		Exposes: make([]*MqttTriggerActionExpose, 0),
		Delay:   utils.IntervalFromMilliseconds(0),
	}
}

func (a *MqttTrigerAction) buildPayload() []byte {

	actionData := map[string]any{}
	for _, expose := range a.Exposes {
		actionData[expose.Name] = expose.Data
	}

	payload, _ := json.Marshal(actionData)
	return payload
}

func (a *MqttTrigerAction) Execute(ctx *DeviceContext) error {
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
		payload := a.buildPayload()

		a.emit(payload)

		// on success callback
		// update sensor current value so we dont have to query the device again
		for _, expose := range a.Exposes {
			ctx.SetCurrent(expose.Name, expose.Data)
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

			payload := a.buildPayload()

			a.emit(payload)

			// on success callback
			// update sensor current value so we dont have to query the device again
			for _, expose := range a.Exposes {
				ctx.SetCurrent(expose.Name, expose.Data)
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
	newValue, err := a.operation.Next()
	if err != nil {
		return err
	}

	// build payload
	actionData := map[string]any{
		a.Property: newValue,
	}

	payload, _ := json.Marshal(actionData)
	if err != nil {
		utils.LogErrorf("step action failed %s ", err.Error())
		return err
	}

	a.emit(payload)

	ctx.SetCurrent(a.Property, newValue)

	return nil
}

type MqttPresetCyclingAction struct {
	MqttBaseAction
	Property  string   `json:"property"`
	Presets   []string `json:"presets,omitempty"`
	operation actionOperation
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
	newValue, err := a.operation.Next()
	if err != nil {
		return err
	}

	// build payload
	actionData := map[string]any{
		a.Property: newValue,
	}

	payload, _ := json.Marshal(actionData)
	if err != nil {
		utils.LogErrorf("preset cycling action failed %s ", err.Error())
		return err
	}

	a.emit(payload)

	ctx.SetCurrent(a.Property, newValue)
	return nil
}

type MqttBaseAction struct {
	Id           string                   `json:"id"`
	FriendlyName string                   `json:"friendlyname"`
	Type         string                   `json:"type"`
	Client       mqtt.MqttClient          `json:"-"`
	registrar    services.DeviceRegistrar `json:"-"`
	mut          sync.RWMutex             `json:"-"`
	exit         chan bool                `json:"-"`
	isPending    bool                     `json:"-"`
}

func (a *MqttBaseAction) emit(payload []byte) {

	msg := fmt.Sprintf("%s/set", a.FriendlyName)
	a.Client.Publish(msg, payload)

	utils.LogInfof("Action triggered. Message %s published in %s", string(payload), a.FriendlyName)
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

	we need to check if value is still the same and not emit anything new
	// b.mut.Lock()
	// defer b.mut.Unlock()

	// if b.isPending {
	// 		return fmt.Errorf("action %s is already pending", b.Id) // Return an error
	// }

	// b.isPending = true
	// defer func() { b.isPending = false }() // Ensure isPending is reset

	return nil
}

// trigger action
// could have delay
// can have multiple exposes to generate payload

// step action
// has expose property
// has min/max limits
// has sinlge expose Daya
// can have multiple steps

// rotation action
// has preset
// has current index for rotation

type MqttActionTest interface {
	Execute(tx *DeviceContext) error
	Stop()
	GetID() string
	GetType() string
}

var typeRegistry = map[string]reflect.Type{
	"trigger":  reflect.TypeOf(MqttTrigerAction{}),
	"step":     reflect.TypeOf(MqttStepAction{}),
	"rotation": reflect.TypeOf(MqttPresetCyclingAction{}),
}

func UnmarshalAction(data []byte) (MqttActionTest, error) {
	var baseAction MqttBaseAction
	if err := json.Unmarshal(data, &baseAction); err != nil {
		return nil, fmt.Errorf("unmarshaling base action: %w", err)
	}

	// Look up the concrete type in the registry
	concreteType, ok := typeRegistry[baseAction.Type]
	if !ok {
		return nil, fmt.Errorf("unknown action type: %s", baseAction.Type)
	}

	// Create a new value of the concrete type
	action := reflect.New(concreteType).Interface().(MqttActionTest)

	// Unmarshal the full JSON into the concrete type
	if err := json.Unmarshal(data, action); err != nil {
		return nil, fmt.Errorf("unmarshaling concrete action: %w", err)
	}

	return action, nil
}

// ////////////////////////////////////////////////////////////
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

func (a *MqttAction) Configure(registrar services.DeviceRegistrar) error {

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
			utils.LogErrorf("build payload failed: %s", err.Error())
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
		//utils.LogInfof("time constraint started Delay: %d ms", a.Delay)

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
			//utils.LogInfo("timer constraint finished")
			return

		case <-a.exit:

			//utils.LogInfo("timer constraint stopped")
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
		newValue, err := a.operationAction.Next()
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
