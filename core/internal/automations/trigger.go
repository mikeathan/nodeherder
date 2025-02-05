package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/models/devices"
	"node-herder/utils"
	"reflect"
)

type Condition struct {
	Name             string `json:"name"`
	Value            any    `json:"value"`
	EqualityOperator string `json:"equality"`
}

func (c *Condition) Evaluate(exposes map[string]*devices.Entity) bool {

	expose, ok := exposes[c.Name]
	if !ok {
		utils.LogDebugf("sensor %s not found in payload", c.Name)
		return false
	}

	if EqualityOperators[c.EqualityOperator](expose.Data, c.Value) {
		return true
	}

	return false
}

type Trigger struct {
	Conditions []*Condition `json:"conditions"`
	Actions    []MqttAction `json:"actions"`
	Name       string       `json:"name"`
}

func NewTrigger(name string) *Trigger {
	return &Trigger{
		Conditions: []*Condition{},
		Actions:    []MqttAction{},
	}
}

func (t *Trigger) UnmarshalJSON(data []byte) error {
	// Unmarshal into a temporary struct to get basic fields
	var temp struct {
		Name       string            `json:"name"`
		Conditions []*Condition      `json:"conditions"`
		Actions    []json.RawMessage `json:"actions"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	t.Name = temp.Name
	t.Conditions = temp.Conditions

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

 TODO:
// automation
// Trigger 1:
// Condition -> door open
//  Action  -> start alarm (low volume, short duration, melody a)

// Trigger 2:
// Condition -> door open AND (after 3 am) and (before 6 am)
// 	Action  -> start alarm (high volume, long duration, melody b)

// Need new condition type:
// timer before 
// timer after 

func (t *Trigger) process(ctx *DeviceContext) {

	currValue := ctx.GetCurrent(t.Name)
	for _, c := range t.Conditions {

		isMatched := c.Evaluate(ctx.Payload)
		if !isMatched {
			// reset any pending actions eg if its one a timer
			for _, action := range t.Actions {
				action.Stop()
			}
			return
		}

		// avoid calling action again for current trigger if value hasnt changed
		if t.Name == c.Name && currValue == c.Value {
			return
		}
	}

	for _, action := range t.Actions {
		action.Execute(ctx)
	}
}
