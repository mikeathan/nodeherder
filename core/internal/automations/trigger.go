package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/utils"
	"reflect"
)

var typeRegistry = map[string]reflect.Type{
	TriggerAction:       reflect.TypeOf(MqttTriggerAction{}),
	StepAction:          reflect.TypeOf(MqttStepAction{}),
	PresetCyclingAction: reflect.TypeOf(MqttPresetCyclingAction{}),
	ExposeConditionType: reflect.TypeOf(ExposeCondition{}),
}

type Trigger struct {
	Conditions []Condition  `json:"conditions"`
	Actions    []MqttAction `json:"actions"`
	Name       string       `json:"name"`
}

func NewTrigger(name string) *Trigger {
	return &Trigger{
		Conditions: []Condition{},
		Actions:    []MqttAction{},
	}
}

func (t *Trigger) UnmarshalJSON(data []byte) error {
	// Unmarshal into a temporary struct to get basic fields
	var temp struct {
		Name       string            `json:"name"`
		Conditions []json.RawMessage `json:"conditions"`
		Actions    []json.RawMessage `json:"actions"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	t.Name = temp.Name

	// Handle the Conditions using the type registry
	t.Conditions = make([]Condition, len(temp.Conditions))
	for i, rawCondition := range temp.Conditions {
		baseCondition := BaseCondition{}
		if err := json.Unmarshal(rawCondition, &baseCondition); err != nil {
			return fmt.Errorf("unmarshaling base condition: %w", err)
		}
		concreteType, ok := typeRegistry[baseCondition.Type]
		if !ok {
			return fmt.Errorf("unknown condition type: %s", baseCondition.Type)
		}

		condition := reflect.New(concreteType).Interface().(Condition)
		if err := json.Unmarshal(rawCondition, condition); err != nil {
			return fmt.Errorf("unmarshaling concrete condition: %w", err)
		}
		
		handlerInitialiser := NewConditionHandlerInitialiser(WithClock(utils.NewRealClock()))
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

		concreteType, ok := typeRegistry[baseAction.Type]
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

// presence  == true
// light >= 14
// turn on light

// presence == false
// start timer for 5 min
// turn off light

func (t *Trigger) process(ctx *DeviceContext) {

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
