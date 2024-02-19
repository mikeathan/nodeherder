package automations

import (
	"errors"
	"math"
	"node-herder/models/devices"
	"sort"
)

type operationFactory interface {
	Create(expose *devices.Entity, action *MqttAction) actionOperation
}

type stepOperationFactory struct {
	op    string
	limit string
}

func (s *stepOperationFactory) Create(expose *devices.Entity, action *MqttAction) actionOperation {

	var limit float64
	if val, ok := expose.Attributes[s.limit]; ok {
		limit = val.(float64)
	}

	return newStepOperation(expose, action, s.op, limit)
}

type rotateOperationFactory struct {
}

func (r *rotateOperationFactory) Create(expose *devices.Entity, action *MqttAction) actionOperation {

	var keys []string
	for k := range expose.Presets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var presets []any
	for _, k := range keys {
		presets = append(presets, expose.Presets[k])
	}

	return newRotateOperation(presets)
}

var OperationTypes = map[int]operationFactory{
	1: &stepOperationFactory{op: "+", limit: "max"},
	2: &stepOperationFactory{op: "-", limit: "min"},
	3: &rotateOperationFactory{},
}

func toFloat(value any) float32 {
	switch v := value.(type) {
	case int:
		return float32(v)
	case float64:
		return float32(v)
	case float32:
		return float32(v)
	default:
		return float32(0)
	}
}

type actionOperation interface {
	Next(ctx *DeviceContext) (any, error)
}

type rotateOperation struct {
	position int
	size     int
	items    []any
}

func newRotateOperation(items []any) actionOperation {
	return &rotateOperation{items: items, size: len(items)}
}

func (r *rotateOperation) Next(ctx *DeviceContext) (any, error) {
	if r.position >= r.size {
		r.position = 0
	}

	v := r.items[r.position]
	r.position++

	return v, nil
}

type stepOperation struct {
	limit    float64
	stepType string
	action   *MqttAction
	expose   *devices.Entity
}

func newStepOperation(expose *devices.Entity, action *MqttAction, stepType string, limit float64) actionOperation {
	return &stepOperation{expose: expose, stepType: stepType, action: action, limit: limit}
}

func (r *stepOperation) Next(ctx *DeviceContext) (any, error) {

	var sourceValue float64
	var newValue float64 = 0
	var ok bool

	if sourceValue, ok = ctx.GetCurrent(r.action.Property).(float64); !ok {
		sourceValue = 0.0
	}

	for i := len(r.action.Steps) - 1; i >= 0; i-- {
		step := r.action.Steps[i]
		if newValue, ok = ctx.GetCurrent(step.Property).(float64); ok {
			newValue = numericOperations[r.stepType](newValue, r.action.Data.(float64), r.limit)
		}
	}

	if sourceValue == newValue {
		return nil, errors.New("same value, skipping")
	}

	return newValue, nil

}

var numericOperations = map[string]func(float64, float64, float64) float64{
	"+": func(v1 float64, v2 float64, limit float64) float64 {

		newValue := v1 + v2
		if limit != 0 {
			newValue = math.Min(newValue, limit)
		}
		return newValue
	},
	"-": func(v1 float64, v2 float64, limit float64) float64 {

		newValue := v1 - v2
		newValue = math.Max(newValue, limit)

		return newValue
	}, "*": func(v1 float64, v2 float64, limit float64) float64 {
		return v1 * v2
	},
}

var EqualityOperators = map[string]func(any, any) bool{
	"=": func(v1 any, v2 any) bool {
		return v1 == v2
	},
	">=": func(v1 any, v2 any) bool {
		return toFloat(v1) >= toFloat(v2)
	},
	"<=": func(v1 any, v2 any) bool {
		return toFloat(v1) <= toFloat(v2)
	},
	">": func(v1 any, v2 any) bool {
		return toFloat(v1) > toFloat(v2)
	},
	"<": func(v1 any, v2 any) bool {
		return toFloat(v1) < toFloat(v2)
	},
}
