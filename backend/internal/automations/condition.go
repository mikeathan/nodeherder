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
)

var conditionTypeRegistry = map[ConditionType]reflect.Type{
	ExposeConditionType: reflect.TypeOf(ExposeCondition{}),
}

// BaseCondition
type BaseCondition struct {
	EqualityOperator utils.EqualityOperator `json:"equality"`
	Type             ConditionType          `json:"type"`
}

// ExposeCondition
type ExposeCondition struct {
	BaseCondition
	Name      string     `json:"name"`
	Value     any        `json:"value"`
	TimeRange *TimeRange `json:"timeRange,omitempty"`
	handlers  []ConditionHandler
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

func NewExposeCondition(name string, value any, operation utils.EqualityOperator) *ExposeCondition {
	cond := &ExposeCondition{
		Name:      name,
		Value:     value,
		TimeRange: nil,
		handlers:  []ConditionHandler{},
		BaseCondition: BaseCondition{
			Type:             ExposeConditionType,
			EqualityOperator: operation,
		},
	}

	err := exposeHandlerInitializer[cond.Type](cond)
	if err != nil {
		utils.LogErrorf("error initialising expose condition %s", err.Error())
		return nil
	}

	return cond
}
func NewExposeConditionwithTimeRange(name string, value any, operation utils.EqualityOperator, timeRange *TimeRange, clock utils.Clock) *ExposeCondition {

	cond := &ExposeCondition{
		Name:      name,
		Value:     value,
		TimeRange: timeRange,
		handlers:  []ConditionHandler{},
		BaseCondition: BaseCondition{
			Type:             ExposeConditionType,
			EqualityOperator: operation,
		},
	}

	initialiser := newConditionHandlerInitialiser(WithClock(clock))
	err := initialiser[cond.Type](cond)
	if err != nil {
		utils.LogErrorf("initialising expose timeRange failed. Error: %s", err.Error())
		return nil
	}
	return cond
}

var exposeHandlerInitializer = newConditionHandlerInitialiser()

type ConditionHandlerOption func(*ConditionHandlerOptions)
type ConditionHandlerOptions struct {
	Clock utils.Clock
}

func WithClock(clock utils.Clock) func(*ConditionHandlerOptions) {
	return func(opts *ConditionHandlerOptions) {
		opts.Clock = clock
	}
}

func newConditionHandlerInitialiser(opts ...ConditionHandlerOption) map[ConditionType]func(Condition) error {
	options := &ConditionHandlerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return map[ConditionType]func(Condition) error{
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

type Condition interface {
	Evaluate(ctx AutomationContext) bool
	GetType() ConditionType
	HasValueChanged(name string, ctx AutomationContext) bool
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
