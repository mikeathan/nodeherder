package automations_test

import (
	"math"
	"node-herder/internal/automations"
	"node-herder/mocks"
	"sort"
	"testing"
)

func TestOperationIncreaseValue(t *testing.T) {

	ctx := automations.NewDeviceContext()
	rotation := createEntity("light", "some description", 51.0, "", nil)
	rotation.Attributes["max"] = 255.0
	rotation.Attributes["min"] = 0.0

	mqtt := &mocks.MockMqttClient{}
	turnOnAction := &automations.MqttAction{}
	turnOnAction.FriendlyName = "button_rotation_slow"
	turnOnAction.Property = "state"
	turnOnAction.Data = 10.0
	turnOnAction.Delay = 0
	turnOnAction.Operation = 1 // increase step
	turnOnAction.Client = mqtt
	operationAction := automations.OperationTypes[turnOnAction.Operation].Create(rotation, turnOnAction)
	max := rotation.Attributes["max"].(float64)
	for i := 0; i < 150; i++ {
		nextValue, er := operationAction.Next(ctx)
		if er != nil { // we are expecting value is same error
			t.Fatalf("error %v", er.Error())
		}

		rotationValue := rotation.Data.(float64)
		stepValue := turnOnAction.Data.(float64)

		got := nextValue.(float64)
		if got > max {
			t.Fatalf("max limit invalid operation value: want %v got %v", max, got)
		}

		want := rotationValue + stepValue
		want = math.Min(want, max)

		if got != want {
			t.Fatalf("invalid operation value: want %v got %v", want, got)
		}

		rotation.Data = got
	}
}

func TestOperationDecreaseValue(t *testing.T) {

	ctx := automations.NewDeviceContext()

	rotation := createEntity("light", "some description", 43.0, "", nil)
	rotation.Attributes["max"] = 255.0
	rotation.Attributes["min"] = 0.0

	mqtt := &mocks.MockMqttClient{}
	turnOnAction := &automations.MqttAction{}
	turnOnAction.FriendlyName = "button_rotation_slow"
	turnOnAction.Property = "state"
	turnOnAction.Data = 10.0
	turnOnAction.Delay = 0
	turnOnAction.Operation = 2 // decrease step
	turnOnAction.Client = mqtt
	operationAction := automations.OperationTypes[turnOnAction.Operation].Create(rotation, turnOnAction)
	min := rotation.Attributes["min"].(float64)
	for i := 0; i < 150; i++ {
		nextValue, er := operationAction.Next(ctx)
		if er != nil { // we are expecting value is same error
			t.Fatalf("error %v", er.Error())
		}

		rotationValue := rotation.Data.(float64)
		stepValue := turnOnAction.Data.(float64)

		got := nextValue.(float64)
		if got < min {
			t.Fatalf("min limit invalid operation value: want %v got %v", min, got)
		}

		want := rotationValue - stepValue
		want = math.Max(want, min)

		if got != want {
			t.Fatalf("invalid operation value: want %v got %v", want, got)
		}

		rotation.Data = got
	}
}

func TestOperationCycleValue(t *testing.T) {

	ctx := automations.NewDeviceContext()

	light := createEntity("light", "some description", 0.0, "", nil)
	light.Presets["cold"] = 255.0
	light.Presets["hot"] = 123.0
	light.Presets["colder"] = 89.0
	light.Presets["hotter"] = 67.0
	light.Presets["natural"] = 43.0

	var keys []string
	for k := range light.Presets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var pos int = 0
	var presets []any
	for _, k := range keys {
		presets = append(presets, light.Presets[k])

	}

	mqtt := &mocks.MockMqttClient{}
	turnOnAction := &automations.MqttAction{}
	turnOnAction.FriendlyName = "button_rotation_slow"
	turnOnAction.Property = "state"
	turnOnAction.Data = 0
	turnOnAction.Delay = 0
	turnOnAction.Operation = 3 // decrease step
	turnOnAction.Client = mqtt
	operationAction := automations.OperationTypes[turnOnAction.Operation].Create(light, turnOnAction)

	for i := 0; i < 15; i++ {
		nextValue, er := operationAction.Next(ctx)
		if er != nil { // we are expecting value is same error
			continue
		}

		want := presets[pos].(float64)
		got := nextValue.(float64)
		if got != want {
			t.Fatalf("invalid operation value: want %v got %v", want, got)
		}

		pos++
		if pos > len(presets)-1 {
			pos = 0
		}
	}
}
