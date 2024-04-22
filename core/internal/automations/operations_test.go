package automations_test

import (
	"encoding/json"
	"fmt"
	"math"
	"node-herder/internal/automations"
	"node-herder/internal/services"
	"node-herder/internal/ws"
	"node-herder/mocks"
	"node-herder/models/devices"
	repository "node-herder/repository/devices"
	"sort"
	"strings"
	"testing"
	"time"
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

func newMockHubRegisterService(mockBroadcastEvent func(eventName string, data interface{}) error) ws.EventHub {
	return &mocks.MockEventHub{MockBroadcastEvent: mockBroadcastEvent}
}

func TestOperationMultiStepIncreaseValue(t *testing.T) {

	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	repo := repository.NewMemoryDeviceRepo()

	testDevices := []struct {
		id       string
		name     string
		property string
		data     any
	}{
		{id: "x1234", name: "livingroom", property: "brightness", data: nil},
		{id: "x5678", name: "button", property: "action_time", data: nil},
	}

	// initialize mock devices
	for _, d := range testDevices {
		newDevice := createMockDevice(d.id, d.name, d.property, d.data)
		devices[d.name] = newDevice
		repo.Store(d.id, newDevice)
	}

	var messageHandler = func(id string, payload []byte) {
		//"livingroom light/set"
		//"{\"brightness\":1}"
		name := strings.Replace(id, "/set", "", -1)
		data := make(map[string]interface{})
		err := json.Unmarshal(payload, &data)
		if err != nil {
			t.Fatalf("invalid payload")

		}
		fmt.Println("mqtt", id, payload)
		devices[name].Exposes["brightness"].Data = data["brightness"]
	}

	mqtt := &mocks.MockMqttClient{}
	mqtt.OnMessageHandler(messageHandler)
	eventHub := &mocks.MockEventHub{}

	registrar := services.NewHubRegisterService(repo, eventHub, 30000)

	// create trigger automation
	ctx := automations.NewDeviceContext()
	// rotation := createEntity("livingroom light", "livingroom light description", 51.0, "", nil)
	// rotation.Attributes["max"] = 255.0
	// rotation.Attributes["min"] = 0.0

	action := &automations.MqttAction{}
	action.FriendlyName = "livingroom"
	action.Id = "x1234"
	action.Property = "brightness"
	action.Type = automations.StepAction
	action.Data = float64(1)
	brightnessStep := automations.Step{}
	brightnessStep.Property = "brightness"
	brightnessStep.Operator = "+"
	brightnessStep.Id = "x1234"

	actionTimeStep := automations.Step{}
	actionTimeStep.Property = "action_time"
	actionTimeStep.Operator = "*"
	actionTimeStep.Id = "x5678"

	action.Steps = append(action.Steps, brightnessStep)
	action.Steps = append(action.Steps, actionTimeStep)

	action.Client = mqtt
	err := action.Configure(registrar)

	if err != nil {
		t.Fatalf("error configuring action %v ", err.Error())
	}

	devices["button"].Exposes["action_time"].Data = 10
	ctx.Payload["action_time"] = devices["button"].Exposes["action_time"]
	action.Execute("action_time", ctx)

	// operationAction := automations.CreateStepOperation(rotation, action)
	// max := rotation.Attributes["max"].(float64)

	// //ctx.SetCurrent("brightness", 0.0)
	// action_times := []float64{10.0}
	// // brightness = brightness + action_time * 0.5
	// for i := 0; i < len((action_times)); i++ {
	// 	//ctx.SetCurrent("action_time", action_times[i]) // set new action_time value to simulate a new device update

	// 	nextValue, er := operationAction.Next(ctx)
	// 	if er != nil {
	// 		t.Fatalf("error %v iteration %d action_time %v", er.Error(), i, action_times[i])
	// 	}
	// 	got := nextValue.(float64)
	// 	if got > max {
	// 		t.Fatalf("max limit invalid operation value: want %v got %v", max, got)
	// 	}

	// 	// brightness = brightness + action_time * 0.5
	// 	want := ctx.GetCurrent("brightness").(float64) + (action_times[i] * action.Data.(float64))
	// 	want = math.Min(want, max)

	// 	if got != want {
	// 		t.Fatalf("invalid operation value: want %v got %v", want, got)
	// 	}

	// 	ctx.SetCurrent("brightness", got) // set current updated brightness value
	// }

	// val, er := operationAction.Next(ctx)
	// if er == nil {
	// 	t.Fatalf("expected error got %v", val)
	// }
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

func createMockDevice(id string, name string, property string, data any) *devices.Device {
	device1 := &devices.Device{}
	device1.Id = id
	device1.FriendlyName = name
	device1.ConnectionType = "mqtt"
	device1.Description = fmt.Sprintf("Test device %s description", id)
	device1.PowerSource = "mains"
	device1.Properties = map[string]any{}
	device1.Properties["last_seen"] = time.Now().Format(time.RFC3339)
	device1.Properties["link_quality"] = 45.0
	device1.Exposes = make(map[string]*devices.Entity)

	ent1 := &devices.Entity{}
	ent1.Description = fmt.Sprintf("%s readings", property)
	ent1.Name = property
	ent1.Unit = "test"
	ent1.Data = data

	device1.Exposes[property] = ent1

	return device1
}
