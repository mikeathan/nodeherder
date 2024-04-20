package automations

import (
	"errors"
	"fmt"
	"math"
	"node-herder/models/devices"
	"sort"
)

func CreateStepOperation(expose *devices.Entity, action *MqttAction) actionOperation {

	var minLimit float64 = 0
	var maxLimit float64 = 255
	if val, ok := expose.Attributes["min"]; ok {
		minLimit = val.(float64)
	}
	if val, ok := expose.Attributes["max"]; ok {
		maxLimit = val.(float64)
	}

	return newStepOperation(expose, action, minLimit, maxLimit)
}

func CreateRotateOperation(expose *devices.Entity, action *MqttAction) actionOperation {

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
	action *MqttAction
	expose *devices.Entity
	limits map[string]float64
}

func newStepOperation(expose *devices.Entity, action *MqttAction, minLimit float64, maxLimit float64) actionOperation {

	limits := make(map[string]float64)
	limits["+"] = maxLimit
	limits["-"] = minLimit
	return &stepOperation{expose: expose, action: action, limits: limits}
}

func (r *stepOperation) Next(ctx *DeviceContext) (any, error) {

	// [brightness] = value
	// [action_time] = value

	// actionId := r.action.Id
	// actionProperty := r.action.Property
	// if sourceValue, err := r.action.registrar.RetrieveEntityData(actionId, actionProperty); err == nil {
	// 	sourceValue = sourceValue.(float64)
	// 	fmt.Println("[DEBUG] property: ", r.action.Property, " sourcevValue ", sourceValue)
	// }

	result := r.action.Data.(float64)
	for i := len(r.action.Steps) - 1; i >= 0; i-- {
		step := r.action.Steps[i]
		if value, err := r.action.registrar.RetrieveEntityData(step.Id, step.Property); err == nil {
			stepValue, ok := value.(float64)
			if !ok {
				stepValue = 0.0
				fmt.Println("[DEBUG] default to zero")
			}

			result = numericOperations[step.Operator](stepValue, result, r.limits[step.Operator])
			fmt.Println("[DEBUG] step:", step.Id, ", ", step.Property, ",", step.Operator, " data: ", stepValue, " result: ", result)

		}
	}
	fmt.Println("[DEBUG] result ", result)

	// ??
	if sourceValue == result {
		return nil, errors.New("same value, skipping")
	}

	// brightness = 10
	// action_time = 20
	// coeffiecient = 0.5

	// 	eg brightness = brightness + action_time * 0.5

	// NOTE;
	// for multi step operation
	// 	eg brightness = brightness + action_time * 0.5
	// for single step operation
	//  eg brightness = brightness + 0.5
	// var newValue = r.action.Data.(float64)
	// for i := len(r.action.Steps) - 1; i >= 0; i-- {
	// 	step := r.action.Steps[i]

	// 	t := ctx.Payload[step.Property] // get data for the trigger device
	// 	// we need data for the action device too

	// 	fmt.Println(t.Data)
	// 	if propValue, ok := ctx.GetCurrent(step.Property).(float64); ok {
	// 		fmt.Println("[DEBUG]  ", step.Property, " current value ", propValue)
	// 		newValue = numericOperations[step.Operator](propValue, newValue, r.limits[step.Operator])
	// 	}
	// }
	// fmt.Println("[DEBUG] newvalue ", newValue)

	// if sourceValue == newValue {
	// 	return nil, errors.New("same value, skipping")
	// }

	return result, nil
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
