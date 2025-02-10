package automations

import (
	"node-herder/utils"
)

const (
	ExposeConditionType = "expose"
	TimeConditionType   = "time"
)

var conditionHandlerInitialiser = map[string]func(Condition) error{
	TimeConditionType: timeConditionInitialiser,
}

// func timeConditionInitialiser(condition Condition) error {
// 	tc := condition.(*TimeCondition)
// 	t, err := ConvertStringToTime(tc.Value)
// 	if err != nil {
// 		return err
// 	}
// 	tc.timeAt = t
// 	tc.clock = utils.NewRealClock()
// 	return nil
// }

type Condition interface {
	Evaluate(ctx *DeviceContext) bool
	GetType() string
	HasValueChanged(name string, ctx *DeviceContext) bool
}

// BaseCondition
type BaseCondition struct {
	EqualityOperator string `json:"equality"`
	Type             string `json:"type"`
}

// ExposeCondition
type ExposeCondition struct {
	BaseCondition
	Name      string          `json:"name"`
	Value     any             `json:"value"`
	schedules []*TimeSchedule `json:"schedules"`
}

func (e *ExposeCondition) HasValueChanged(name string, ctx *DeviceContext) bool {
	return ctx.GetCurrent(name) != e.Value
}

func (e *ExposeCondition) Evaluate(ctx *DeviceContext) bool {
	expose, ok := ctx.GetPayload(e.Name)
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

func NewExposeCondition(name string, value any, operation string) *ExposeCondition {
	return &ExposeCondition{
		Name:      name,
		Value:     value,
		schedules: []*TimeSchedule{},
		BaseCondition: BaseCondition{
			Type:             ExposeConditionType,
			EqualityOperator: operation,
		},
	}
}

// TimeCondition
// type TimeCondition struct {
// 	BaseCondition
// 	Value  string `json:"value"`
// 	timeAt time.Time
// 	clock  utils.Clock
// }

// func (t *TimeCondition) Evaluate(ctx *DeviceContext) bool {

// 	result, _ := t.clock.CompareWithNow(t.timeAt, t.EqualityOperator)
// 	return result
// }

// func (t *TimeCondition) GetType() string {
// 	return TimeConditionType
// }

// func (e *TimeCondition) HasValueChanged(name string, ctx *DeviceContext) bool {
// 	return true
// }

// func NewTimeCondition(value string, operation string, clock utils.Clock) (*TimeCondition, error) {
// 	t, err := ConvertStringToTime(value)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &TimeCondition{
// 		Value:  value,
// 		timeAt: t,
// 		clock:  clock,
// 		BaseCondition: BaseCondition{
// 			Type:             TimeConditionType,
// 			EqualityOperator: operation,
// 		},
// 	}, nil
// }
