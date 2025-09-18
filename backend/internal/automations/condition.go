package automations

import (
	"fmt"
	"node-herder/utils"
	"reflect"
	"time"
)

type ConditionType string

const (
	ExposeConditionType ConditionType = "expose"
	TimeConditionType   ConditionType = "time"
)

var conditionTypeRegistry = map[ConditionType]reflect.Type{
	ExposeConditionType: reflect.TypeOf(ExposeCondition{}),
	TimeConditionType:   reflect.TypeOf(TimeCondition{}),
}

type Condition interface {
	Evaluate(ctx AutomationContext) bool
	GetType() ConditionType
	HasValueChanged(name string, ctx AutomationContext) bool
	InitHandlers(clock utils.Clock) error
}

// BaseCondition
type BaseCondition struct {
	EqualityOperator utils.EqualityOperator `json:"equality"`
	Type             ConditionType          `json:"type"`
	TimeRange        *TimeRange             `json:"timeRange,omitempty"`
	handlers         []ConditionHandler
}

func (e *BaseCondition) Evaluate(ctx AutomationContext) bool {
	return true
}

func (e *BaseCondition) GetType() ConditionType {
	return e.Type
}

func (e *BaseCondition) HasValueChanged(name string, ctx AutomationContext) bool {
	return true
}

func (b *BaseCondition) InitHandlers(clock utils.Clock) error {

	if b.TimeRange != nil {
		trHandler, err := NewTimeRangeHandler(b.TimeRange, clock)
		if err != nil {
			return err
		}
		b.handlers = append(b.handlers, trHandler)
	}
	return nil
}

func NewBaseCondition(conditionType ConditionType, operator utils.EqualityOperator, timeRange *TimeRange) *BaseCondition {
	return &BaseCondition{
		Type:             conditionType,
		EqualityOperator: operator,
		TimeRange:        timeRange,
		handlers:         []ConditionHandler{},
	}
}

// ExposeCondition
type ExposeCondition struct {
	BaseCondition
	Name  string `json:"name"`
	Value any    `json:"value"`
}

func (e *ExposeCondition) HasValueChanged(name string, ctx AutomationContext) bool {
	return ctx.GetCurrent(name) != e.Value
}

func (e *ExposeCondition) Evaluate(ctx AutomationContext) bool {
	for _, handler := range e.handlers {
		if !handler.Evaluate(ctx) {
			return false
		}
	}
	return true
}

func (e *ExposeCondition) GetType() ConditionType {
	return ExposeConditionType
}

func (e *ExposeCondition) InitHandlers(clock utils.Clock) error {
	if err := e.BaseCondition.InitHandlers(clock); err != nil {
		return err
	}

	eh := &ExposeHandler{
		Name:      e.Name,
		Value:     e.Value,
		Operation: e.EqualityOperator,
	}
	e.handlers = append(e.handlers, eh)
	return nil
}

// Expose Handler
type ExposeHandler struct {
	Name      string
	Operation utils.EqualityOperator
	Value     any
}

func NewExposeHandler(cond *ExposeCondition) (*ExposeHandler, error) {
	if _, ok := utils.EqualityOperators[cond.EqualityOperator]; !ok {
		return nil, fmt.Errorf("invalid operation %s", cond.EqualityOperator)
	}

	if cond.Value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	if cond.Name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	return &ExposeHandler{
		Name:      cond.Name,
		Value:     cond.Value,
		Operation: cond.EqualityOperator,
	}, nil
}

func (e *ExposeHandler) Evaluate(ctx AutomationContext) bool {
	expose, ok := ctx.GetPayload(e.Name)
	if !ok {
		utils.LogDebugf("sensor %s not found in payload", e.Name)
		return false
	}

	if utils.EqualityOperators[e.Operation](expose.Data, e.Value) {
		return true
	}

	return false
}

// TimeRange Handler
type TimeRangeHandler struct {
	startTime time.Time
	endTime   time.Time
	clock     utils.Clock
}

func (t *TimeRangeHandler) Evaluate(ctx AutomationContext) bool {
	return t.clock.IsInRange(t.startTime, t.endTime)
}

func NewTimeRangeHandler(timeRange *TimeRange, clock utils.Clock) (*TimeRangeHandler, error) {
	st, err := ConvertStringToTime(timeRange.StartAt)
	if err != nil {
		return nil, fmt.Errorf("invalid TimeRangeHandler.StartAt format %s", err.Error())
	}

	et, err := ConvertStringToTime(timeRange.EndAt)
	if err != nil {
		return nil, fmt.Errorf("invalid TimeRangeHandler.EndAt format %s", err.Error())
	}

	if st.After(et) {
		return nil, fmt.Errorf("startAt cannot be after endAt")
	}
	return &TimeRangeHandler{
		startTime: st,
		endTime:   et,
		clock:     clock,
	}, nil
}

type ConditionHandler interface {
	Evaluate(ctx AutomationContext) bool
}

type TimeRange struct {
	StartAt string `json:"startAt"`
	EndAt   string `json:"endAt"`
}

func NewTimeRange(startAt string, endAt string) *TimeRange {
	return &TimeRange{
		StartAt: startAt,
		EndAt:   endAt,
	}
}

// TimeCondition
type TimeCondition struct {
	BaseCondition
}

func (e *TimeCondition) HasValueChanged(name string, ctx AutomationContext) bool {
	return true
}

func (e *TimeCondition) Evaluate(ctx AutomationContext) bool {
	if len(e.handlers) == 0 {
		utils.LogError("no handlers found for time condition")
		return false
	}

	for _, handler := range e.handlers {
		if !handler.Evaluate(ctx) {
			return false
		}
	}
	return true
}

func (e *TimeCondition) GetType() ConditionType {
	return TimeConditionType
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
