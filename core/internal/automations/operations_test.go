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
	turnOnAction.FriendlyName = "attic light"
	turnOnAction.Property = "brightness"
	turnOnAction.Data = 1.0
	step := automations.Step{}
	step.Property = "brightness"
	step.Operator = "+"

	turnOnAction.Steps = append(turnOnAction.Steps, step)
	turnOnAction.Client = mqtt
	operationAction := automations.CreateStepOperation(rotation, turnOnAction)
	max := rotation.Attributes["max"].(float64)

	ctx.SetCurrent("brightness", 0.0)

	for i := 0; i < 255; i++ {
		nextValue, er := operationAction.Next(ctx)
		if er != nil {
			t.Fatalf("error %v", er.Error())
		}

		got := nextValue.(float64)
		if got > max {
			t.Fatalf("max limit invalid operation value: want %v got %v", max, got)
		}

		want := ctx.GetCurrent("brightness").(float64) + turnOnAction.Data.(float64)
		want = math.Min(want, max)

		if got != want {
			t.Fatalf("invalid operation value: want %v got %v", want, got)
		}
		ctx.SetCurrent("brightness", got)
	}

	val, er := operationAction.Next(ctx)
	if er == nil {
		t.Fatalf("expected error got %v", val)
	}
}

func TestOperationDecreaseValue(t *testing.T) {

	ctx := automations.NewDeviceContext()

	rotation := createEntity("light", "some description", 43.0, "", nil)
	rotation.Attributes["max"] = 255.0
	rotation.Attributes["min"] = 0.0

	mqtt := &mocks.MockMqttClient{}
	turnOnAction := &automations.MqttAction{}
	turnOnAction.FriendlyName = "attic light"
	turnOnAction.Property = "brightness"
	turnOnAction.Data = 1.0
	step := automations.Step{}
	step.Property = "brightness"
	step.Operator = "-"

	turnOnAction.Steps = append(turnOnAction.Steps, step)

	turnOnAction.Client = mqtt
	operationAction := automations.CreateStepOperation(rotation, turnOnAction)
	min := rotation.Attributes["min"].(float64)

	ctx.SetCurrent("brightness", 255.0)

	for i := 0; i < 255; i++ {
		nextValue, er := operationAction.Next(ctx)
		if er != nil {
			t.Fatalf("error %v", er.Error())
		}

		got := nextValue.(float64)
		if got < min {
			t.Fatalf("min limit invalid operation value: want %v got %v", min, got)
		}

		want := ctx.GetCurrent("brightness").(float64) - turnOnAction.Data.(float64)
		want = math.Max(want, min)

		if got != want {
			t.Fatalf("invalid operation value: want %v got %v", want, got)
		}

		ctx.SetCurrent("brightness", got)
	}

	val, er := operationAction.Next(ctx)
	if er == nil {
		t.Fatalf("expected error got %v", val)
	}
}

func TestOperationMultiStepIncreaseValue(t *testing.T) {

	ctx := automations.NewDeviceContext()
	rotation := createEntity("light", "some description", 51.0, "", nil)
	rotation.Attributes["max"] = 255.0
	rotation.Attributes["min"] = 0.0

	mqtt := &mocks.MockMqttClient{}
	turnOnAction := &automations.MqttAction{}
	turnOnAction.FriendlyName = "attic light"
	turnOnAction.Property = "brightness"
	turnOnAction.Data = 0.5
	step := automations.Step{}
	step.Property = "brightness"
	step.Operator = "+"

	step2 := automations.Step{}
	step2.Property = "action_time"
	step2.Operator = "*"

	turnOnAction.Steps = append(turnOnAction.Steps, step)
	turnOnAction.Steps = append(turnOnAction.Steps, step2)

	turnOnAction.Client = mqtt
	operationAction := automations.CreateStepOperation(rotation, turnOnAction)
	max := rotation.Attributes["max"].(float64)

	ctx.SetCurrent("brightness", 0.0)
	action_times := []float64{10.0, 20.0, 10.0, 40.0, 80.0, 20.0, 100.0, 50.0, 80.0, 20.0, 20.0, 30.0, 30.0}
	// brightness = brightness + action_time * 0.5
	for i := 0; i < len((action_times)); i++ {
		ctx.SetCurrent("action_time", action_times[i]) // set new action_time value to simulate a new device update

		nextValue, er := operationAction.Next(ctx)
		if er != nil {
			t.Fatalf("error %v iteration %d action_time %v", er.Error(), i, action_times[i])
		}

		got := nextValue.(float64)
		if got > max {
			t.Fatalf("max limit invalid operation value: want %v got %v", max, got)
		}

		// brightness = brightness + action_time * 0.5
		want := ctx.GetCurrent("brightness").(float64) + (action_times[i] * turnOnAction.Data.(float64))
		want = math.Min(want, max)

		if got != want {
			t.Fatalf("invalid operation value: want %v got %v", want, got)
		}

		ctx.SetCurrent("brightness", got) // set current updated brightness value
	}

	val, er := operationAction.Next(ctx)
	if er == nil {
		t.Fatalf("expected error got %v", val)
	}
}

func TestOperationMultiStepDecreaseValue(t *testing.T) {

	ctx := automations.NewDeviceContext()
	rotation := createEntity("light", "some description", 51.0, "", nil)
	rotation.Attributes["max"] = 255.0
	rotation.Attributes["min"] = 0.0

	mqtt := &mocks.MockMqttClient{}
	turnOnAction := &automations.MqttAction{}
	turnOnAction.FriendlyName = "attic light"
	turnOnAction.Property = "brightness"
	turnOnAction.Data = 0.5
	step := automations.Step{}
	step.Property = "brightness"
	step.Operator = "-"

	step2 := automations.Step{}
	step2.Property = "action_time"
	step2.Operator = "*"

	turnOnAction.Steps = append(turnOnAction.Steps, step)
	turnOnAction.Steps = append(turnOnAction.Steps, step2)

	turnOnAction.Client = mqtt
	operationAction := automations.CreateStepOperation(rotation, turnOnAction)
	max := rotation.Attributes["max"].(float64)

	ctx.SetCurrent("brightness", 255.0)
	action_times := []float64{10.0, 20.0, 10.0, 40.0, 80.0, 20.0, 100.0, 50.0, 80.0, 20.0, 20.0, 30.0, 30.0}
	// brightness = brightness - action_time * 0.5

	for i := 0; i < len((action_times)); i++ {
		ctx.SetCurrent("action_time", action_times[i]) // set new action_time value to simulate a new device update

		nextValue, er := operationAction.Next(ctx)
		if er != nil {
			t.Fatalf("error %v iteration %d action_time %v", er.Error(), i, action_times[i])
		}

		got := nextValue.(float64)
		if got > max {
			t.Fatalf("max limit invalid operation value: want %v got %v", max, got)
		}

		// brightness = brightness - action_time * 0.5
		want := ctx.GetCurrent("brightness").(float64) - (action_times[i] * turnOnAction.Data.(float64))
		want = math.Min(want, max)

		if got != want {
			t.Fatalf("invalid operation value: want %v got %v", want, got)
		}

		ctx.SetCurrent("brightness", got) // set current updated brightness value
	}

	val, er := operationAction.Next(ctx)
	if er == nil {
		t.Fatalf("expected error got %v", val)
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
	turnOnAction.FriendlyName = "attic light"
	turnOnAction.Property = "brightness"
	turnOnAction.Data = 0
	turnOnAction.Delay = 0
	turnOnAction.Client = mqtt
	operationAction := automations.CreateRotateOperation(light, turnOnAction)

	for i := 0; i < 15; i++ {
		nextValue, er := operationAction.Next(ctx)
		if er != nil {
			t.Fatalf("error %v", er.Error())
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
