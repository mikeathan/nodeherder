package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/utils"
	"reflect"
)

type TriggerType string
type TriggerList []Trigger

type Trigger interface {
	GetType() TriggerType
	GetName() string
	Process(ctx AutomationContext)
	GetActions() []MqttAction
	GetConditions() []Condition
}

const (
	DeviceTriggerType TriggerType = "deviceTrigger"
	ManualTriggerType TriggerType = "manualTrigger"
)

var triggerTypeRegistry = map[TriggerType]reflect.Type{
	DeviceTriggerType: reflect.TypeOf(&DeviceTrigger{}),
	ManualTriggerType: reflect.TypeOf(&ManualTrigger{}),
}

type ManualTrigger struct {
	BaseTrigger
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

func (t *BaseTrigger) Process(ctx AutomationContext) {
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

		handlerInitialiser := newConditionHandlerInitialiser(WithClock(utils.NewRealClock()))
		if conditionInitialiser, ok := handlerInitialiser[condition.GetType()]; ok {
			err := conditionInitialiser(condition)
			if err != nil {
				return err
			}
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

func (t *DeviceTrigger) Process(ctx AutomationContext) {

	for _, c := range t.Conditions {

		isMatched := c.Evaluate(ctx)
		if !isMatched {
			// reset any pending actions eg if its a timer
			for _, action := range t.Actions {
				action.Stop()
			}
			return
		}

		if !c.HasValueChanged(t.Name, ctx) {
			return
		}
	}

	for _, action := range t.Actions {
		action.Execute(ctx)
	}
}


// ManualTrigger
func (t *ManualTrigger) Process(ctx AutomationContext) {
	for _, action := range t.Actions {
		action.Execute(ctx)
	}
}


// TriggerList
// Wrapper to control the unmarshalling of different Trigger types

func (tl *TriggerList) UnmarshalJSON(data []byte) error {
	var rawList []json.RawMessage
	if err := json.Unmarshal(data, &rawList); err != nil {
		return err
	}

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

		*tl = append(*tl, trigger.(Trigger))
	}
	return nil
}
