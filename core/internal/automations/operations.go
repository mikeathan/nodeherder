package automations

import (
	"errors"
	"math"
)

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
	Next(value float64) (any, error)
}

type rotateOperation struct {
	position int
	size     int
	items    []any
}

func newRotateOperation(items []any) actionOperation {
	return &rotateOperation{items: items, size: len(items)}
}

func (r *rotateOperation) Next(value float64) (any, error) {
	if r.position >= r.size {
		r.position = 0
	}

	v := r.items[r.position]
	r.position++

	return v, nil
}

type stepOperation struct {
	limit     float64
	stepType  int
	stepValue float64
}

func newStepOperation(stepType int, stepValue float64, limit float64) actionOperation {
	return &stepOperation{stepType: stepType, stepValue: stepValue, limit: limit}
}

func (r *stepOperation) Next(value float64) (any, error) {
	op := stepsOperators[r.stepType]

	newValue := numericOperations[op.Operator](value, r.stepValue, r.limit)
	if value == newValue {
		return nil, errors.New("same value, skipping")
	}

	return newValue, nil
}

type numericOperator struct {
	Operator string
	Limit    string
}

var stepsOperators = map[int]*numericOperator{
	1: {Operator: "+", Limit: "max"},
	2: {Operator: "-", Limit: "min"},
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
