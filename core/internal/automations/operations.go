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
	return newStepOperation(expose, s.op, action.Data.(float64), limit)
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
	Next() (any, error)
}

type rotateOperation struct {
	position int
	size     int
	items    []any
}

func newRotateOperation(items []any) actionOperation {
	return &rotateOperation{items: items, size: len(items)}
}

func (r *rotateOperation) Next() (any, error) {
	if r.position >= r.size {
		r.position = 0
	}

	v := r.items[r.position]
	r.position++

	return v, nil
}

type stepOperation struct {
	limit     float64
	stepType  string
	stepValue float64
	expose    *devices.Entity
}

func newStepOperation(expose *devices.Entity, stepType string, stepValue float64, limit float64) actionOperation {
	return &stepOperation{expose: expose, stepType: stepType, stepValue: stepValue, limit: limit}
}

func (r *stepOperation) Next() (any, error) {

	value := r.stepValue
	if r.expose.Data != nil {
		value = r.expose.Data.(float64)
	}

	newValue := numericOperations[r.stepType](value, r.stepValue, r.limit)
	if value == newValue {
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
