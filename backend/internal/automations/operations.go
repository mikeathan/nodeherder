package automations

import (
	"errors"
	"math"
	"node-herder/models/devices"
	"sort"
)

func CreateTriggerOperation(action *MqttTriggerAction) actionOperation {
	actionData := map[string]any{}
	for _, expose := range action.Exposes {
		actionData[expose.Name] = expose.Data
	}

	return newTriggerOperation(actionData)
}

func CreateStepOperation(expose *devices.Entity, action *MqttStepAction) actionOperation {

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

func CreateRotateOperation(expose *devices.Entity) actionOperation {

	var keys []string
	for k := range expose.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var presets []any
	for _, k := range keys {
		presets = append(presets, expose.Values[k])
	}

	return newRotateOperation(expose.Name, presets)
}

type actionOperation interface {
	CreatePayload() (map[string]any, error)
}

type triggerOperation struct {
	data map[string]any
}

func newTriggerOperation(data map[string]any) actionOperation {
	return &triggerOperation{data: data}
}

func (t *triggerOperation) CreatePayload() (map[string]any, error) {
	return t.data, nil
}

type rotateOperation struct {
	property string
	position int
	size     int
	items    []any
}

func newRotateOperation(property string, items []any) actionOperation {
	return &rotateOperation{property: property, items: items, size: len(items)}
}

func (r *rotateOperation) CreatePayload() (map[string]any, error) {
	if r.position >= r.size {
		r.position = 0
	}

	v := r.items[r.position]
	r.position++

	payload := map[string]any{
		r.property: v,
	}
	return payload, nil
}

type stepOperation struct {
	action      *MqttStepAction
	expose      *devices.Entity
	limits      map[string]float64
	propertyMap map[string]float64
}

func newStepOperation(expose *devices.Entity, action *MqttStepAction, minLimit float64, maxLimit float64) actionOperation {

	limits := make(map[string]float64)
	limits["+"] = maxLimit
	limits["-"] = minLimit
	return &stepOperation{expose: expose, action: action, limits: limits, propertyMap: make(map[string]float64, len(action.Steps))}
}

func (r *stepOperation) CreatePayload() (map[string]any, error) {

	// example:
	// brightness = 10
	// action_time = 20
	// coeffiecient = 0.5

	// 	eg brightness = brightness + action_time * 0.5
	result := r.action.Data.(float64) // coefficient
	for i := len(r.action.Steps) - 1; i >= 0; i-- {
		step := r.action.Steps[i]
		if value, err := r.action.registrar.RetrieveEntityData(step.Id, step.Property); err == nil {
			r.propertyMap[step.Property] = 0

			if stepValue, ok := value.(float64); ok {
				result = numericOperations[step.Operator](stepValue, result, r.limits[step.Operator])

				// Cache value for equality check.  temp needs refactoring
				r.propertyMap[step.Property] = stepValue
			}
		}
	}

	if r.propertyMap[r.action.Property] == result {
		return nil, errors.New("same value, skipping")
	}

	payload := map[string]any{
		r.action.Property: result,
	}
	return payload, nil
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
