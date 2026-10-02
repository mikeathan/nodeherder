package automations_test

import (
	"encoding/json"
	"fmt"
	"math"
	"node-herder/internal/automations"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestOperationIncreaseValue(t *testing.T) {

	stepValue := 17.0
	expectedCalls := 255 / int(stepValue)

	responses := 0

	mqtt := &mocks.MockMqttClient{}
	t.Cleanup(mqtt.WaitResponses)
	currentActionTimeIndex := 0

	// create mock action
	action := createMockLivingRoomStepAction("+", stepValue)
	action.Client = mqtt

	repo := createMockLivingRoomButtonDevices(0.0, 0.0)

	store := utils_test.CreateStoreFromDeviceRepo(repo)

	//store dd in map for easy access
	var dd map[string]*devices.Device = make(map[string]*devices.Device)
	dev1, _ := repo.FindDevice("x1234")
	dev2, _ := repo.FindDevice("x5678")

	dd["livingroom"] = dev1
	dd["button"] = dev2
	max := dd["livingroom"].Exposes["brightness"].Attributes["max"].(float64)

	var messageHandler = func(id string, payload []byte) {
		responses++
		if responses > expectedCalls {
			t.Errorf("unexpected command after reaching limit")
			return
		}

		//name := strings.Replace(id, "/set", "", -1)
		data := make(map[string]interface{})
		err := json.Unmarshal(payload, &data)
		if err != nil {
			t.Errorf("invalid payload")
			return
		}

		// 	eg. brightness = brightness + action_time * 0.5
		got := data["brightness"].(float64)
		if got > max {
			t.Errorf("max limit invalid operation value: want %v got %v", max, got)
		}

		var prevValue float64 = 0
		// need to get the previous value of brightness before update
		brightnessExpose := dd["livingroom"].Exposes["brightness"]
		prevValue = brightnessExpose.Data.Value().(float64)

		want := prevValue + action.Data.(float64)
		want = math.Min(want, max)

		if got != want {
			t.Errorf("invalid operation value: want %v got %v", want, got)
			return
		}

		brightnessExpose.Data.SetValue(got)
	}

	mqtt.OnMessageHandler(messageHandler)
	eventHub := &mocks.MockEventHub{}

	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	// create trigger automation
	ctx := automations.NewDeviceContext()
	err := action.Configure(registrar, mqtt)

	if err != nil {
		t.Fatalf("error configuring action %v ", err.Error())
	}
	for i := 0; i < expectedCalls; i++ {

		// action_time property of button is not really used for calculation,
		// is just a triggering device so we can publish the payload
		ctx.SetDevicePayload(map[string]*devices.Entity{
			"action_time": dd["button"].Exposes["action_time"],
		})

		if err := action.Execute(ctx); err != nil {
			t.Fatal(err)
		}

		mqtt.WaitResponses()
		currentActionTimeIndex++
	}

	// we are expecting to have reached the max value of the 'brightness' property
	// so next payload event shoud not publish new mqqt message. if it does it should error in the handler

	action.Execute(ctx)
	mqtt.WaitResponses()

	if responses != expectedCalls {
		t.Fatalf("unexpected command count: %d", responses)
	}
}

func TestOperationDecreaseValue(t *testing.T) {

	stepValue := 17.0
	expectedCalls := 255 / int(stepValue)

	responses := 0

	mqtt := &mocks.MockMqttClient{}
	t.Cleanup(mqtt.WaitResponses)
	currentActionTimeIndex := 0

	// create mock action
	action := createMockLivingRoomStepAction("-", stepValue)
	action.Client = mqtt

	repo := createMockLivingRoomButtonDevices(255.0, 0.0)
	store := utils_test.CreateStoreFromDeviceRepo(repo)

	//store dd in map for easy access
	var dd map[string]*devices.Device = make(map[string]*devices.Device)
	dev1, _ := repo.FindDevice("x1234")
	dev2, _ := repo.FindDevice("x5678")

	dd["livingroom"] = dev1
	dd["button"] = dev2
	max := dd["livingroom"].Exposes["brightness"].Attributes["max"].(float64)

	var messageHandler = func(id string, payload []byte) {
		responses++
		if responses > expectedCalls {
			t.Errorf("unexpected command after reaching limit")
			return
		}

		//name := strings.Replace(id, "/set", "", -1)
		data := make(map[string]interface{})
		err := json.Unmarshal(payload, &data)
		if err != nil {
			t.Errorf("invalid payload")
			return
		}

		// 	eg. brightness = brightness + action_time * 0.5
		got := data["brightness"].(float64)
		if got > max {
			t.Errorf("max limit invalid operation value: want %v got %v", max, got)
		}

		var prevValue float64 = 0
		// need to get the previous value of brightness before update
		prevValue = dd["livingroom"].Exposes["brightness"].Data.Value().(float64)

		want := prevValue - action.Data.(float64)
		want = math.Min(want, max)
		if got != want {
			t.Errorf("invalid operation value: want %v got %v", want, got)
		}

		dd["livingroom"].Exposes["brightness"].Data.SetValue(got)
	}

	mqtt.OnMessageHandler(messageHandler)
	eventHub := &mocks.MockEventHub{}
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	// create trigger automation
	ctx := automations.NewDeviceContext()
	err := action.Configure(registrar, mqtt)

	if err != nil {
		t.Fatalf("error configuring action %v ", err.Error())
	}
	for i := 0; i < expectedCalls; i++ {

		// action_time property of button is not really used for calculation,
		// is just a triggering device so we can publish the payload

		ctx.SetDevicePayload(map[string]*devices.Entity{
			"action_time": dd["button"].Exposes["action_time"],
		})

		if err := action.Execute(ctx); err != nil {
			t.Fatal(err)
		}

		mqtt.WaitResponses()
		currentActionTimeIndex++
	}

	// we are expecting to have reached the max value of the 'brightness' property
	// so next payload event shoud not publish new mqqt message. if it does it should error in the handler

	action.Execute(ctx)
	mqtt.WaitResponses()
	if responses != expectedCalls {
		t.Fatalf("unexpected command count: %d", responses)
	}
}

func TestOperationMultiStepIncreaseValue(t *testing.T) {
	action_times := []float64{10.0, 20.0, 10.0, 40.0, 80.0, 20.0, 100.0, 50.0, 80.0, 20.0, 20.0, 30.0, 30.0}

	responses := 0

	mqtt := &mocks.MockMqttClient{}
	t.Cleanup(mqtt.WaitResponses)
	currentActionTimeIndex := 0

	// create mock action
	action := createMockLivingRoomButtonStepAction("+", 0.5)
	action.Client = mqtt

	repo := createMockLivingRoomButtonDevices(0.0, 0.0)
	store := utils_test.CreateStoreFromDeviceRepo(repo)

	//store dd in map for easy access
	var dd map[string]*devices.Device = make(map[string]*devices.Device)
	dev1, _ := repo.FindDevice("x1234")
	dev2, _ := repo.FindDevice("x5678")

	dd["livingroom"] = dev1
	dd["button"] = dev2

	max := dd["livingroom"].Exposes["brightness"].Attributes["max"].(float64)

	var messageHandler = func(id string, payload []byte) {
		responses++
		if responses > len(action_times) {
			t.Errorf("unexpected command after reaching limit")
			return
		}

		name := strings.Replace(id, "/set", "", -1)
		data := make(map[string]interface{})
		err := json.Unmarshal(payload, &data)
		if err != nil {
			t.Errorf("invalid payload")
			return
		}

		// 	eg. brightness = brightness + action_time * 0.5
		got := data["brightness"].(float64)
		if got > max {
			t.Errorf("max limit invalid operation value: want %v got %v", max, got)
		}

		var prevValue float64 = 0
		// need to get the previous value of brightness before update
		prevValue = dd[name].Exposes["brightness"].Data.Value().(float64)

		want := prevValue + (action_times[currentActionTimeIndex] * action.Data.(float64))
		want = math.Min(want, max)
		if got != want {
			t.Errorf("invalid operation value: want %v got %v", want, got)
		}

		dd[name].Exposes["brightness"].Data.SetValue(got)
	}

	mqtt.OnMessageHandler(messageHandler)
	eventHub := &mocks.MockEventHub{}

	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	// create trigger automation
	ctx := automations.NewDeviceContext()
	err := action.Configure(registrar, mqtt)

	if err != nil {
		t.Fatalf("error configuring action %v ", err.Error())
	}
	for _, actionTime := range action_times {
		// update both device and payload as they are used
		dd["button"].Exposes["action_time"].Data.SetValue(actionTime)
		ctx.SetDevicePayload(map[string]*devices.Entity{
			"action_time": dd["button"].Exposes["action_time"],
		})
		if err := action.Execute(ctx); err != nil {
			t.Fatal(err)
		}

		mqtt.WaitResponses()
		currentActionTimeIndex++
	}

	// we are expecting to have reached the max value of the 'brightness' property
	// so next payload event shoud not publish new mqqt message. if it does it should error in the handler
	dd["button"].Exposes["action_time"].Data.SetValue(float64(30))
	ctx.SetDevicePayload(map[string]*devices.Entity{
		"action_time": dd["button"].Exposes["action_time"],
	})
	action.Execute(ctx)
	mqtt.WaitResponses()
	if responses != len(action_times) {
		t.Fatalf("unexpected command count: %d", responses)
	}
}

func TestOperationMultiStepDecreaseValue(t *testing.T) {

	action_times := []float64{10.0, 20.0, 10.0, 40.0, 80.0, 20.0, 100.0, 50.0, 80.0, 20.0, 20.0, 30.0, 30.0}
	// brightness = brightness - action_time * 0.5

	responses := 0

	mqtt := &mocks.MockMqttClient{}
	t.Cleanup(mqtt.WaitResponses)
	currentActionTimeIndex := 0

	// create mock action
	action := createMockLivingRoomButtonStepAction("-", 0.5)
	action.Client = mqtt

	repo := createMockLivingRoomButtonDevices(255.0, 0.0)
	store := utils_test.CreateStoreFromDeviceRepo(repo)

	//store dd in map for easy access
	var dd map[string]*devices.Device = make(map[string]*devices.Device)
	dev1, _ := repo.FindDevice("x1234")
	dev2, _ := repo.FindDevice("x5678")

	dd["livingroom"] = dev1
	dd["button"] = dev2

	max := dd["livingroom"].Exposes["brightness"].Attributes["max"].(float64)

	var messageHandler = func(id string, payload []byte) {
		responses++
		if responses > len(action_times) {
			t.Errorf("unexpected command after reaching limit")
			return
		}
		name := strings.Replace(id, "/set", "", -1)
		data := make(map[string]interface{})
		err := json.Unmarshal(payload, &data)
		if err != nil {
			t.Errorf("invalid payload")
			return
		}

		// 	eg. brightness = brightness - action_time * 0.5
		got := data["brightness"].(float64)
		if got > max {
			t.Errorf("max limit invalid operation value: want %v got %v", max, got)
		}

		var prevValue float64 = 0
		// need to get the previous value of brightness before update
		prevValue = dd[name].Exposes["brightness"].Data.Value().(float64)

		want := prevValue - (action_times[currentActionTimeIndex] * action.Data.(float64))
		want = math.Min(want, max)
		if got != want {
			t.Errorf("invalid operation value: want %v got %v", want, got)
		}

		dd[name].Exposes["brightness"].Data.SetValue(got)
	}

	mqtt.OnMessageHandler(messageHandler)
	eventHub := &mocks.MockEventHub{}

	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	// create trigger automation
	ctx := automations.NewDeviceContext()
	err := action.Configure(registrar, mqtt)

	if err != nil {
		t.Fatalf("error configuring action %v ", err.Error())
	}
	for _, actionTime := range action_times {
		// update both device and payload as they are used
		dd["button"].Exposes["action_time"].Data.SetValue(actionTime)
		ctx.SetDevicePayload(map[string]*devices.Entity{
			"action_time": dd["button"].Exposes["action_time"],
		})
		if err := action.Execute(ctx); err != nil {
			t.Fatal(err)
		}

		mqtt.WaitResponses()
		currentActionTimeIndex++
	}

	// we are expecting to have reached the max value of the 'brightness' property
	// so next payload event shoud not publish new mqqt message. if it does it should error in the handler
	dd["button"].Exposes["action_time"].Data.SetValue(float64(30))
	ctx.SetDevicePayload(map[string]*devices.Entity{
		"action_time": dd["button"].Exposes["action_time"],
	})

	action.Execute(ctx)
	mqtt.WaitResponses()
	if responses != len(action_times) {
		t.Fatalf("unexpected command count: %d", responses)
	}
}

func TestOperationCycleValue(t *testing.T) {

	exposeName := "light"
	light := createEntity(exposeName, "some description", 0.0, "", nil)
	light.Values["cold"] = 255.0
	light.Values["hot"] = 123.0
	light.Values["colder"] = 89.0
	light.Values["hotter"] = 67.0
	light.Values["natural"] = 43.0

	var keys []string
	for k := range light.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var pos int = 0
	var presets []any
	for _, k := range keys {
		presets = append(presets, light.Values[k])

	}

	mqtt := &mocks.MockMqttClient{}
	turnOnAction := automations.NewTriggerAction()
	turnOnAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "brightness",
			Data: 0,
		},
	}

	turnOnAction.Client = mqtt

	operationAction := automations.CreateRotateOperation(light)

	for i := 0; i < 15; i++ {
		payload, er := operationAction.CreatePayload()
		if er != nil {
			t.Fatalf("error %v", er.Error())
		}

		want := presets[pos].(float64)
		got := payload.Commands[exposeName].(float64)
		if got != want {
			t.Fatalf("invalid operation value: want %v got %v", want, got)
		}

		pos++
		if pos > len(presets)-1 {
			pos = 0
		}
	}
}

func createMockDevice(id string, name string, property string, data any, min float64, max float64) *devices.Device {
	device1 := &devices.Device{}

	device1.Id = id
	device1.FriendlyName = name
	device1.ConnectionType = "mqtt"
	device1.Description = fmt.Sprintf("Test device %s description", id)
	device1.PowerSource = "mains"
	device1.LastSeen = time.Now().Format(time.RFC3339)
	device1.Exposes = make(map[string]*devices.Entity)

	ent1 := devices.NewEntity(property)

	ent1.Description = fmt.Sprintf("%s readings", property)
	ent1.Unit = "test"
	ent1.Data.SetValue(data)

	device1.Exposes[property] = ent1
	device1.Exposes[property].Attributes = make(map[string]any)
	device1.Exposes[property].Attributes["min"] = min
	device1.Exposes[property].Attributes["max"] = max
	return device1
}

func createMockLivingRoomButtonStepAction(operation string, stepValue float64) *automations.MqttStepAction {
	action := automations.NewStepAction()
	action.Id = "x1234"
	action.Property = "brightness"
	action.Type = automations.StepAction
	action.Data = stepValue
	brightnessStep := &automations.Step{}
	brightnessStep.Property = "brightness"
	brightnessStep.Operator = operation
	brightnessStep.Id = "x1234"

	actionTimeStep := &automations.Step{}
	actionTimeStep.Property = "action_time"
	actionTimeStep.Operator = "*"
	actionTimeStep.Id = "x5678"

	action.Steps = append(action.Steps, brightnessStep)
	action.Steps = append(action.Steps, actionTimeStep)

	return action
}

func createMockLivingRoomStepAction(operation string, stepValue float64) *automations.MqttStepAction {
	action := automations.NewStepAction()
	action.Id = "x1234"
	action.Property = "brightness"
	action.Type = automations.StepAction
	action.Data = stepValue
	brightnessStep := &automations.Step{}
	brightnessStep.Property = "brightness"
	brightnessStep.Operator = operation
	brightnessStep.Id = "x1234"

	action.Steps = append(action.Steps, brightnessStep)
	return action
}

func createMockLivingRoomButtonDevices(brightnessValue float64, actionTimeValue float64) devices.Repository {
	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	repo := repository.NewMemoryDeviceRepo()

	testDevices := []struct {
		id       string
		name     string
		property string
		data     any
	}{
		{id: "x1234", name: "livingroom", property: "brightness", data: brightnessValue},
		{id: "x5678", name: "button", property: "action_time", data: actionTimeValue},
	}

	// initialize mock devices
	for _, d := range testDevices {
		newDevice := createMockDevice(d.id, d.name, d.property, d.data, 0.0, 255.0)

		devices[d.name] = newDevice
		repo.Store(d.id, newDevice)
	}

	return repo
}
