package automations

import (
	"node-herder/utils"
	"time"
)

const (
	ExposeConditionType = "ExposeCondition"
	TimeConditionType   = "TimeCondition"
)

type Condition interface {
	Evaluate(ctx *DeviceContext) bool
	GetType() string
}

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

// ExposeCondition
type ExposeCondition struct {
	BaseCondition
	Name  string `json:"name"`
	Value any    `json:"value"`
}

func (e *ExposeCondition) Evaluate(ctx *DeviceContext) bool {
	expose, ok := ctx.Payload[e.Name]
	if !ok {
		utils.LogDebugf("sensor %s not found in payload", e.Name)
		return false
	}

	if EqualityOperators[e.EqualityOperator](expose.Data, e.Value) {
		return true
	}

	return false
}

func (e *ExposeCondition) GetType() string {
	return ExposeConditionType
}

func NewExposeCondition(name string, value any) *ExposeCondition {
	return &ExposeCondition{
		Name:  name,
		Value: value,
	}
}

type TimeCondition struct {
	BaseCondition
	Value string `json:"value"`
	time  time.Time
}

// TimeCondition
func (t *TimeCondition) Evaluate(ctx *DeviceContext) bool {
	return false
}

func (t *TimeCondition) GetType() string {
	return TimeConditionType
}

func NewTimeCondition(value string) (*TimeCondition, error) {
	t, err := ConvertStringToTime(value)
	if err != nil {
		return nil, err
	}
	return &TimeCondition{
		Value: value,
		time:  t,
	}, nil
}

// BaseCondition
type BaseCondition struct {
	EqualityOperator string `json:"equality"`
	Type             string `json:"type"`
}

// type Condition struct {
// 	Name             string `json:"name"`
// 	Value            any    `json:"value"`
// 	EqualityOperator string `json:"equality"`
// }

// func (c *Condition) Evaluate(exposes map[string]*devices.Entity) bool {

// 	expose, ok := exposes[c.Name]
// 	if !ok {
// 		utils.LogDebugf("sensor %s not found in payload", c.Name)
// 		return false
// 	}

// 	if EqualityOperators[c.EqualityOperator](expose.Data, c.Value) {
// 		return true
// 	}

// 	return false
// }
