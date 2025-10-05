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

type PublishMode string

type ActionType string

const (
	PublishBatch  PublishMode = "batch"  // all commands in one payload
	PublishSingle PublishMode = "single" // one payload per command
)

const (
	TriggerAction       ActionType = "trigger"
	StepAction          ActionType = "step"
	PresetCyclingAction ActionType = "preset"
)

var actionTypeRegistry = map[ActionType]reflect.Type{
	TriggerAction:       reflect.TypeOf(MqttTriggerAction{}),
	StepAction:          reflect.TypeOf(MqttStepAction{}),
	PresetCyclingAction: reflect.TypeOf(MqttPresetCyclingAction{}),
}

type Step struct {
	Property string `json:"property"`
	Operator string `json:"operator"` // +,-, *,/
	Id       string `json:"id"`
}

type MqttTriggerActionExpose struct {
	Name string `json:"name"`
	Data any    `json:"data"`
}

type MqttTriggerAction struct {
	MqttBaseAction
	Exposes     []*MqttTriggerActionExpose `json:"exposes"`
	Delay       *utils.TimeInterval        `json:"delay,omitempty"`
	PublishMode PublishMode                `json:"publishMode,omitempty"`
}

func NewTriggerAction() *MqttTriggerAction {
	return &MqttTriggerAction{
		Exposes: make([]*MqttTriggerActionExpose, 0),
		Delay:   nil,
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

	if a.Delay != nil && a.Delay.Value == 0 {
		return fmt.Errorf("delay must be greater than zero")
	}

	// NOTE: to refactor and remove from model. add to some handler to perfom the job
	a.operation = CreateTriggerOperation(a)

	// common logic
	device, err := registrar.LookupById(a.Id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %s ", a.Id, err.Error())
	}
	// NOTE: if device is renamed we might need to register the automations again
	a.friendlyName = device.FriendlyName
	a.Id = device.Id
	a.Client = client
	a.registrar = registrar
	return nil
}

func (a *MqttTriggerAction) Execute(ctx AutomationContext) error {
	if a.Delay == nil {
		return a.executeBase(ctx)
	}

	return a.executeBaseWithDelay(a.Delay, ctx)
}

type MqttStepAction struct {
	MqttBaseAction
	Property string  `json:"property"`
	Steps    []*Step `json:"steps"`
	Data     any     `json:"data"`
}

func NewStepAction() *MqttStepAction {
	return &MqttStepAction{
		Steps: make([]*Step, 0),
		MqttBaseAction: MqttBaseAction{
			Type: StepAction,
		},
	}
}

func (a *MqttStepAction) Execute(ctx AutomationContext) error {
	return a.executeBase(ctx)
}

func (a *MqttStepAction) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

	device, err := registrar.LookupById(a.Id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %s ", a.Id, err.Error())
	}
	expose := device.Exposes[a.Property]

	// NOTE: to refactor and remove from model. add to some handler to perfom the job
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
	Property string `json:"property"`
}

func NewPresetCyclingAction() *MqttPresetCyclingAction {
	return &MqttPresetCyclingAction{
		MqttBaseAction: MqttBaseAction{
			Type: PresetCyclingAction,
		},
	}
}

func (a *MqttPresetCyclingAction) Execute(ctx AutomationContext) error {
	return a.executeBase(ctx)
}

func (a *MqttPresetCyclingAction) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

	device, err := registrar.LookupById(a.Id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %s ", a.Id, err.Error())
	}

	expose := device.Exposes[a.Property]

	// NOTE: to refactor and remove from model. add to some handler to perfom the job
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
	Type         ActionType               `json:"type"`
	Client       mqtt.MqttClient          `json:"-"`
	registrar    services.DeviceRegistrar `json:"-"`
	mut          sync.RWMutex             `json:"-"`
	exit         chan bool                `json:"-"`
	isPending    bool                     `json:"-"`
	friendlyName string                   `json:"-"`
	operation    actionOperation          `json:"-"`
}

func (a *MqttBaseAction) emit(payload []byte) {

	msg := fmt.Sprintf("%s/set", a.friendlyName)
	a.Client.Publish(msg, payload)

	utils.LogInfof("Action triggered. Message %s published in %s", string(payload), a.friendlyName)

}

func (b *MqttBaseAction) processAction(ctx AutomationContext) error {

	payload, err := b.operation.CreatePayload()
	if err != nil {
		return err
	}

	if payload.PublishMode == PublishSingle {
		for key, value := range payload.Commands {
			single := map[string]any{key: value}

			bytes, err := json.Marshal(single)
			if err != nil {
				utils.LogErrorf("Action triggered. Failed to marshal command %s: %s", key, err.Error())
				return err
			}
			// emit message
			b.emit(bytes)
		}
	} else {
		bytes, err := json.Marshal(payload.Commands)
		if err != nil {
			utils.LogErrorf("Action triggered. Failed to marshal payload: %s", err.Error())
			return err
		}

		// emit message
		b.emit(bytes)
	}

	// on success update device context with new values to avoid querying the device again
	// since dont know if action is succeed it we store it as a pending state
	for key, value := range payload.Commands {
		// Always resolve and set current state (used elsewhere)
		resolvedValue := b.resolveValue(key, value, ctx)
		ctx.SetCurrentState(key, resolvedValue)

		// Only set pending state for commands that could cause feedback loops
		if b.needsFeedbackPrevention(key, value) {
			ctx.SetPendingState(key, resolvedValue)
		}
	}

	return nil
}

func (b *MqttBaseAction) needsFeedbackPrevention(key string, value any) bool {
	if key == "state" && value == "TOGGLE" {
		return true
	}
	_, isBool := value.(bool)
	return isBool
}

func (b *MqttBaseAction) resolveValue(key string, value any, ctx AutomationContext) any {
	// Handle TOGGLE for binary states
	if valueStr, ok := value.(string); ok && valueStr == "TOGGLE" {
		// Get current state to determine what TOGGLE should become
		currentValue := ctx.GetCurrentState(key)

		// For binary states, toggle the boolean value
		if key == "state" {
			// If no current state exists, default to false (so TOGGLE becomes true)
			if currentValue == nil {
				return true
			}
			currentBool := b.toBool(currentValue)
			return !currentBool // Return the toggled boolean value
		}
	}

	return value
}

// toBool converts various representations to boolean (same as in trigger.go)
func (b *MqttBaseAction) toBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v == "ON" || v == "true"
	case int:
		return v != 0
	case float64:
		return v != 0
	default:
		return false
	}
}

func (b *MqttBaseAction) GetID() string {
	return b.Id
}

func (b *MqttBaseAction) GetType() ActionType {
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

func (b *MqttBaseAction) executeBaseWithDelay(delay *utils.TimeInterval, ctx AutomationContext) error {
	b.mut.Lock()
	defer b.mut.Unlock()

	if b.isPending {
		return nil
	}

	b.exit = make(chan bool, 1)
	go func() {

		timestamp := time.Now().Add(delay.Duration())
		diff := time.Until(timestamp).Milliseconds()

		duration := time.Duration(diff)
		ticker := *time.NewTicker(duration * time.Millisecond)
		b.isPending = true
		//utils.LogInfof("time constraint started Delay: %d ms", a.Delay)

		defer func() {
			close(b.exit)
			b.isPending = false
		}()

		select {
		case <-ticker.C:

			err := b.processAction(ctx)
			if err != nil {
				utils.LogErrorf("trigger action failed %s", err.Error())
				return
			}
			//utils.LogInfo("timer constraint finished")
			return

		case <-b.exit:

			//utils.LogInfo("timer constraint stopped")
			return
		}
	}()

	return nil
}

func (b *MqttBaseAction) executeBase(ctx AutomationContext) error {
	b.mut.Lock()
	defer b.mut.Unlock()

	if b.isPending {
		return nil
	}

	defer func() {
		b.isPending = false
	}()

	b.isPending = true

	err := b.processAction(ctx)
	if err != nil {
		return err
	}

	return nil
}

type MqttAction interface {
	Execute(ctx AutomationContext) error
	Stop()
	GetID() string
	GetType() ActionType
	Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error
}
