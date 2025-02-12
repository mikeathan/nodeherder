package automations

import (
	"fmt"
	"node-herder/utils"
	"time"
)

const (
	ExposeConditionType = "expose"
)

type ConditionHandlerOption func(*ConditionHandlerOptions)
type ConditionHandlerOptions struct {
	Clock utils.Clock
}

func WithClock(clock utils.Clock) func(*ConditionHandlerOptions) {
	return func(opts *ConditionHandlerOptions) {
		opts.Clock = clock
	}
}

func NewConditionHandlerInitialiser(opts ...ConditionHandlerOption) map[string]func(Condition) error {
	options := &ConditionHandlerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return map[string]func(Condition) error{
		ExposeConditionType: func(condition Condition) error {
			return exposeConditionInitialiser(condition, options)
		},
	}
}

func exposeConditionInitialiser(condition Condition, opts *ConditionHandlerOptions) error {
	ec := condition.(*ExposeCondition)
	if ec.TimeRange != nil {

		handler, err := NewTimeRangeHandler(ec.TimeRange, opts.Clock)
		if err != nil {
			return err
		}
		ec.handlers = append(ec.handlers, handler)
	}

	eh, err := NewExposeHandler(ec)
	if err != nil {
		return err
	}

	ec.handlers = append(ec.handlers, eh)
	return nil
}

// Expose Handler
type ExposeHandler struct {
	Name      string
	Operation string
	Value     any
}

func NewExposeHandler(cond *ExposeCondition) (*ExposeHandler, error) {
	if _, ok := EqualityOperators[cond.EqualityOperator]; ok {
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

func (e *ExposeHandler) Evaluate(ctx *DeviceContext) bool {
	expose, ok := ctx.GetPayload(e.Name)
	if !ok {
		utils.LogDebugf("sensor %s not found in payload", e.Name)
		return false
	}

	if EqualityOperators[e.Operation](expose.Data, e.Value) {
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

func (t *TimeRangeHandler) Evaluate(ctx *DeviceContext) bool {
	now := t.clock.Now()
	return now.After(t.startTime) && now.Before(t.endTime)
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

	return &TimeRangeHandler{
		startTime: st,
		endTime:   et,
		clock:     clock,
	}, nil
}

type ConditionHandler interface {
	Evaluate(ctx *DeviceContext) bool
}

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

// ExposeCondition
type ExposeCondition struct {
	BaseCondition
	Name      string     `json:"name"`
	Value     any        `json:"value"`
	TimeRange *TimeRange `json:"timeRange,omitempty"`
	handlers  []ConditionHandler
}

func (e *ExposeCondition) HasValueChanged(name string, ctx *DeviceContext) bool {
	return ctx.GetCurrent(name) != e.Value
}

func (e *ExposeCondition) Evaluate(ctx *DeviceContext) bool {
	for _, handler := range e.handlers {
		if !handler.Evaluate(ctx) {
			return false
		}
	}
	return true
}

func (e *ExposeCondition) GetType() string {
	return ExposeConditionType
}

func NewExposeCondition(name string, value any, operation string) *ExposeCondition {
	return &ExposeCondition{
		Name:      name,
		Value:     value,
		TimeRange: nil,
		handlers:  []ConditionHandler{},
		BaseCondition: BaseCondition{
			Type:             ExposeConditionType,
			EqualityOperator: operation,
		},
	}
}
func NewExposeConditionwithTimeRange(name string, value any, operation string, timeRange *TimeRange) *ExposeCondition {
	return &ExposeCondition{
		Name:      name,
		Value:     value,
		TimeRange: timeRange,
		handlers:  []ConditionHandler{},
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
