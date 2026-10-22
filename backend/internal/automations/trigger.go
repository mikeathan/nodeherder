package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/utils"
	"reflect"
)

type TriggerType string
type TriggerList []Trigger

type Trigger interface {
	// AffectedBy selects relevant device updates without evaluating conditions.
	AffectedBy(ctx AutomationContext, changes map[string]interface{}) bool
	GetType() TriggerType
	GetName() string
	Process(ctx AutomationContext) error
	GetActions() []MqttAction
	GetConditions() []Condition
}

const (
	DeviceTriggerType TriggerType = "deviceTrigger"
)

var triggerTypeRegistry = map[TriggerType]reflect.Type{
	DeviceTriggerType: reflect.TypeOf(&DeviceTrigger{}),
}

type DeviceTrigger struct {
	BaseTrigger
}

type BaseTrigger struct {
	Type       TriggerType  `json:"type"`
	Actions    []MqttAction `json:"actions"`
	Name       string       `json:"name"`
	Conditions []Condition  `json:"conditions"`
}

func (t *BaseTrigger) GetType() TriggerType {
	return t.Type
}

func (t *BaseTrigger) GetName() string {
	return t.Name
}

// AffectedBy accepts the primary input and, for state sources, changes to any
// declared condition dependency. A cached physical event must never be replayed
// because another property changed. Selection has no eligibility history.
func (t *BaseTrigger) AffectedBy(ctx AutomationContext, changes map[string]interface{}) bool {
	if _, primary := changes[t.Name]; primary {
		return true
	}
	source, exists := ctx.GetDevicePayload(t.Name)
	if !exists || source.IsEventValue(source.Data.Value()) {
		return false
	}
	return t.conditionsAffectedBy(changes)
}

func (t *BaseTrigger) conditionsAffectedBy(changes map[string]interface{}) bool {
	for property := range changes {
		for _, condition := range t.Conditions {
			if condition.DependsOn(property) {
				return true
			}
		}
	}
	return false
}

func (t *BaseTrigger) Process(ctx AutomationContext) error {
	return nil
}

func (t *BaseTrigger) GetActions() []MqttAction {
	return t.Actions
}

func (t *BaseTrigger) GetConditions() []Condition {
	return t.Conditions
}

func (t *BaseTrigger) UnmarshalJSON(data []byte) error {
	// Unmarshal into a temporary struct to get basic fields
	var temp struct {
		Name       string            `json:"name"`
		Type       TriggerType       `json:"type"`
		Conditions []json.RawMessage `json:"conditions"`
		Actions    []json.RawMessage `json:"actions"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	t.Name = temp.Name
	t.Type = temp.Type

	// Handle the Conditions using the type registry
	t.Conditions = make([]Condition, len(temp.Conditions))
	for i, rawCondition := range temp.Conditions {
		baseCondition := BaseCondition{}
		if err := json.Unmarshal(rawCondition, &baseCondition); err != nil {
			return fmt.Errorf("unmarshaling base condition: %w", err)
		}
		concreteType, ok := conditionTypeRegistry[baseCondition.Type]
		if !ok {
			return fmt.Errorf("unknown condition type: %s", baseCondition.Type)
		}

		condition := reflect.New(concreteType).Interface().(Condition)
		if err := json.Unmarshal(rawCondition, condition); err != nil {
			return fmt.Errorf("unmarshaling concrete condition: %w", err)
		}

		err := condition.InitHandlers(utils.NewRealClock())
		if err != nil {
			return err
		}

		t.Conditions[i] = condition
	}

	// Handle the Actions using the type registry
	t.Actions = make([]MqttAction, len(temp.Actions))
	for i, rawAction := range temp.Actions {
		baseAction := MqttBaseAction{}
		if err := json.Unmarshal(rawAction, &baseAction); err != nil {
			return fmt.Errorf("unmarshaling base action: %w", err)
		}

		concreteType, ok := actionTypeRegistry[baseAction.Type]
		if !ok {
			return fmt.Errorf("unknown action type: %s", baseAction.Type)
		}

		action := reflect.New(concreteType).Interface().(MqttAction)
		if err := json.Unmarshal(rawAction, action); err != nil {
			return fmt.Errorf("unmarshaling concrete action: %w", err)
		}
		t.Actions[i] = action
	}

	return nil
}

// DeviceTrigger
func NewDeviceTrigger(name string) *DeviceTrigger {
	return &DeviceTrigger{
		BaseTrigger: BaseTrigger{
			Actions:    []MqttAction{},
			Type:       DeviceTriggerType,
			Conditions: []Condition{},
			Name:       name,
		},
	}
}

// Process is the single execution path for device-selected and manual runs.
func (t *DeviceTrigger) Process(ctx AutomationContext) error {
	if !t.conditionsPermitExecution(ctx) {
		return nil
	}
	return t.executeActions(ctx)
}

func (t *DeviceTrigger) conditionsPermitExecution(ctx AutomationContext) bool {
	// Retain feedback protection and the existing condition/value-change policy.
	if len(t.Conditions) == 0 && !ctx.IsManualTrigger() {
		utils.LogInfof("Blocking device-triggered automation without conditions: %s", t.Name)
		return false
	}
	for _, condition := range t.Conditions {
		if !condition.Evaluate(ctx) {
			t.stopActions()
			return false
		}
		if !condition.HasValueChanged(t.Name, ctx) {
			return false
		}
	}
	return true
}

func (t *DeviceTrigger) stopActions() {
	for _, action := range t.Actions {
		action.Stop()
	}
}

func (t *DeviceTrigger) executeActions(ctx AutomationContext) error {
	var failures error
	for i, action := range t.Actions {
		if err := action.Execute(ctx); err != nil {
			failures = errors.Join(failures, fmt.Errorf("trigger %s action %d: %w", t.Name, i, err))
		}
	}
	return failures
}

// TriggerList
// Wrapper to control the unmarshalling of different Trigger types
func (tl *TriggerList) UnmarshalJSON(data []byte) error {
	var rawList []json.RawMessage
	if err := json.Unmarshal(data, &rawList); err != nil {
		return err
	}

	decoded := make(TriggerList, 0, len(rawList))
	for _, raw := range rawList {
		var peek struct {
			Type TriggerType `json:"type"`
		}
		if err := json.Unmarshal(raw, &peek); err != nil {
			return fmt.Errorf("unmarshal type: %w", err)
		}

		concreteType, ok := triggerTypeRegistry[peek.Type]
		if !ok {
			return fmt.Errorf("unknown trigger type: %s", peek.Type)
		}

		trigger := reflect.New(concreteType.Elem()).Interface()

		if err := json.Unmarshal(raw, trigger); err != nil {
			return fmt.Errorf("unmarshal %s: %w", peek.Type, err)
		}

		decoded = append(decoded, trigger.(Trigger))
	}
	*tl = decoded
	return nil
}
