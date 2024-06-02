package automations

import (
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
	Name       string       `json:"name"`
	Conditions []*Condition `json:"conditions"`
	Action     *MqttAction  `json:"action"`
}

func (t *Trigger) process(ctx *DeviceContext) {

	currValue := ctx.GetCurrent(t.Name)
	for _, c := range t.Conditions {

		isMatched := c.Evaluate(ctx.Payload)
		if !isMatched {
			t.Action.Stop()
			return
		}

		// avoid calling action again for current trigger if value hasnt changed
		if t.Name == c.Name && currValue == c.Value {
			return
		}
	}

	t.Action.Execute(t.Name, ctx)
}
