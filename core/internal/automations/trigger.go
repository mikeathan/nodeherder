package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/models/devices"
	"node-herder/utils"
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
	Conditions []*Condition     `json:"conditions"`
	Actions    []MqttActionTest `json:"actions"`
}

func NewTrigger(name string) *Trigger {
	return &Trigger{
		Conditions: []*Condition{},
		Actions:    []MqttActionTest{},
	}
}

func (t *Trigger) UnmarshalJSON(data []byte) error {
	// Define a temporary struct to avoid infinite recursion
	type Alias Trigger
	aux := &Alias{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	t.Name = aux.Name
	t.Conditions = aux.Conditions

	// Unmarshal actions using the custom logic
	var rawActions []json.RawMessage
	if err := json.Unmarshal(data, &struct {
		Actions *[]json.RawMessage `json:"actions"`
	}{Actions: &rawActions}); err != nil {
		return err
	}

	t.Actions = make([]MqttActionTest, len(rawActions))
	for i, rawAction := range rawActions {
		action, err := UnmarshalAction(rawAction)
		if err != nil {
			return fmt.Errorf("unmarshaling action %d: %w", i, err)
		}
		t.Actions[i] = action
	}

	return nil
}

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

		we are going to move this logic in actio.execute ? 
		or come up with sth else
		// avoid calling action again for current trigger if value hasnt changed
		if t.Name == c.Name && currValue == c.Value {
			return
		}
	}

	for _, action := range t.Actions {
		action.Execute(ctx)
	}
}
