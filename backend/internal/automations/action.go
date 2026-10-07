package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/models/devices"
	"node-herder/utils"
	"reflect"
	"sync"
	"sync/atomic"
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

func (a *MqttTriggerAction) MarshalJSON() ([]byte, error) {
	a.configMu.RLock()
	exposes := make([]*MqttTriggerActionExpose, len(a.Exposes))
	if a.Exposes == nil {
		exposes = nil
	}
	for i, expose := range a.Exposes {
		if expose != nil {
			copy := *expose
			exposes[i] = &copy
		}
	}
	wire := struct {
		Id          string                     `json:"id"`
		Type        ActionType                 `json:"type"`
		Exposes     []*MqttTriggerActionExpose `json:"exposes"`
		Delay       *utils.TimeInterval        `json:"delay,omitempty"`
		PublishMode PublishMode                `json:"publishMode,omitempty"`
	}{a.Id, a.Type, exposes, a.Delay, a.PublishMode}
	a.configMu.RUnlock()
	return json.Marshal(wire)
}

func (a *MqttTriggerAction) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {
	a.configMu.RLock()
	id := a.Id
	exposes := make([]MqttTriggerActionExpose, len(a.Exposes))
	for i, property := range a.Exposes {
		if property == nil {
			a.configMu.RUnlock()
			return fmt.Errorf("nil expose in action %s", id)
		}
		exposes[i] = *property
	}
	mode := a.PublishMode
	invalidDelay := a.Delay != nil && a.Delay.Value == 0
	a.configMu.RUnlock()
	bridgeInfo, err := registrar.FindBridgeInfo(id)
	if err != nil {
		return err
	}
	data := make(map[string]any, len(exposes))
	for i := range exposes {
		property := &exposes[i]
		sanitizedData, err := bridgeInfo.SanitiseProperty(property.Name, property.Data)
		if err != nil {
			return fmt.Errorf("failed to sanitize data for action %s: %w", id, err)
		}
		property.Data = sanitizedData
		data[property.Name] = sanitizedData
	}

	if invalidDelay {
		return fmt.Errorf("delay must be greater than zero")
	}

	device, err := registrar.LookupById(id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %w ", id, err)
	}
	a.configMu.Lock()
	defer a.configMu.Unlock()
	for i := range exposes {
		a.Exposes[i].Data = exposes[i].Data
	}
	a.publishConfiguration(device, client, registrar, newTriggerOperation(data, mode))
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
	a.configMu.RLock()
	id := a.Id
	// Copy recipe only; never copy the base action's synchronization fields.
	recipe := &MqttStepAction{Property: a.Property, Data: a.Data}
	for _, step := range a.Steps {
		if step == nil {
			a.configMu.RUnlock()
			return fmt.Errorf("nil step in action %s", id)
		}
		copy := *step
		recipe.Steps = append(recipe.Steps, &copy)
	}
	a.configMu.RUnlock()
	device, err := registrar.LookupById(id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %w ", id, err)
	}
	expose, ok := device.GetExpose(recipe.Property)
	if !ok {
		return fmt.Errorf("configure action %s: property %s not found", id, recipe.Property)
	}
	operation := createStepOperation(expose, recipe, registrar)
	a.configMu.Lock()
	a.publishConfiguration(device, client, registrar, operation)
	a.configMu.Unlock()

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
	a.configMu.RLock()
	id, property := a.Id, a.Property
	a.configMu.RUnlock()
	device, err := registrar.LookupById(id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %w ", id, err)
	}

	expose, ok := device.GetExpose(property)
	if !ok {
		return fmt.Errorf("configure action %s: property %s not found", id, property)
	}
	operation := CreateRotateOperation(expose)
	a.configMu.Lock()
	a.publishConfiguration(device, client, registrar, operation)
	a.configMu.Unlock()
	return nil
}

type MqttBaseAction struct {
	// Do not copy after use. Direct recipe/Client edits require exclusive ownership.
	Id            string                              `json:"id"`
	Type          ActionType                          `json:"type"`
	Client        mqtt.MqttClient                     `json:"-"`
	configMu      sync.RWMutex                        `json:"-"`
	configuration atomic.Pointer[actionConfiguration] `json:"-"`
	mut           sync.RWMutex                        `json:"-"`
	exit          chan struct{}                       `json:"-"`
	isPending     bool                                `json:"-"`
	clock         utils.Clock                         `json:"-"`
}

// Immutable binding; operation-local mutable state retains existing Execute ownership.
type actionConfiguration struct {
	friendlyName string
	client       mqtt.MqttClient
	registrar    services.DeviceRegistrar
	operation    actionOperation
}

// Caller holds configMu; preparation and external lookups have already succeeded.
func (a *MqttBaseAction) publishConfiguration(device *devices.Device, client mqtt.MqttClient, registrar services.DeviceRegistrar, operation actionOperation) {
	if a.Id != device.Id {
		a.Id = device.Id
	}
	a.Client = client
	a.configuration.Store(&actionConfiguration{device.FriendlyName, client, registrar, operation})
}

func (a *actionConfiguration) emit(payload []byte) {

	msg := fmt.Sprintf("%s/set", a.friendlyName)
	a.client.Publish(msg, payload)

	utils.LogInfof("Action triggered. Message %s published in %s", string(payload), a.friendlyName)
}

func (b *MqttBaseAction) processAction(ctx AutomationContext) error {
	configuration := b.configuration.Load()
	if configuration == nil {
		return fmt.Errorf("action is not configured")
	}
	payload, err := configuration.operation.CreatePayload()
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
			configuration.emit(bytes)
		}
	} else {
		bytes, err := json.Marshal(payload.Commands)
		if err != nil {
			utils.LogErrorf("Action triggered. Failed to marshal payload: %s", err.Error())
			return err
		}

		// emit message
		configuration.emit(bytes)
	}

	// on success update device context with new values to avoid querying the device again
	for key, value := range payload.Commands {
		// Always resolve and set current state (used elsewhere)
		resolvedValue := b.resolveValue(key, value, ctx)
		ctx.SetCurrentState(key, resolvedValue)
	}

	return nil
}

func (b *MqttBaseAction) resolveValue(key string, value any, ctx AutomationContext) any {
	// Handle TOGGLE for binary states
	if valueStr, ok := value.(string); ok && valueStr == "TOGGLE" {
		// Get current state to determine what TOGGLE should become
		currentValue := ctx.GetCurrentState(key)

		// For binary states, toggle the boolean value
		if key == "state" {
			// If no current state exists, default to false
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
	b.configMu.RLock()
	defer b.configMu.RUnlock()
	return b.Id
}

func (b *MqttBaseAction) GetType() ActionType {
	return b.Type
}

// SetClock sets the clock that times delayed execution; nil means the real
// clock. Wiring only (tests inject a fake clock), not for use while executing.
func (b *MqttBaseAction) SetClock(clock utils.Clock) {
	b.configMu.Lock()
	defer b.configMu.Unlock()
	b.clock = clock
}

func (b *MqttBaseAction) delayClock() utils.Clock {
	b.configMu.RLock()
	defer b.configMu.RUnlock()
	if b.clock == nil {
		return utils.NewRealClock()
	}
	return b.clock
}

func (a *MqttBaseAction) Stop() {
	a.mut.Lock()
	defer a.mut.Unlock()
	if a.exit != nil {
		// Do not join delayed workers: a committed publish may call Stop.
		close(a.exit)
		a.exit = nil
		a.isPending = false
	}
}

func (b *MqttBaseAction) executeBaseWithDelay(delay *utils.TimeInterval, ctx AutomationContext) error {
	b.mut.Lock()
	defer b.mut.Unlock()

	if b.isPending {
		return nil
	}

	duration := delay.Duration()
	if duration <= 0 {
		return fmt.Errorf("delay must be greater than zero")
	}
	exit := make(chan struct{})
	b.exit = exit
	b.isPending = true
	fired := make(chan struct{})
	timer := b.delayClock().AfterFunc(duration, func() { close(fired) })
	go func() {
		defer timer.Stop()
		defer func() {
			b.mut.Lock()
			defer b.mut.Unlock()
			// A cancelled run must not clear a replacement's pending state.
			if b.exit == exit {
				b.exit = nil
				b.isPending = false
			}
		}()

		select {
		case <-fired:
			// Serialize cancellation versus commitment, not external callbacks.
			b.mut.Lock()
			current := b.exit == exit
			b.mut.Unlock()
			if !current {
				return
			}
			err := b.processAction(ctx)
			if err != nil {
				utils.LogErrorf("trigger action failed %s", err.Error())
				return
			}
			return

		case <-exit:
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
