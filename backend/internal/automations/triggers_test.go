package automations_test

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type delayedLifecycleClient struct {
	mocks.MockMqttClient
	publish func(string, interface{})
}

type actionConfigRegistrar struct {
	services.DeviceRegistrar
	name       string
	fail       bool
	bridge     *devices.BridgeInfo
	brightness *devices.Entity
}

func (r *actionConfigRegistrar) FindBridgeInfo(string) (*devices.BridgeInfo, error) {
	return r.bridge, nil
}
func (r *actionConfigRegistrar) LookupById(id string) (*devices.Device, error) {
	if r.fail {
		return nil, fmt.Errorf("lookup failed")
	}
	d := devices.NewDevice(id)
	d.FriendlyName = r.name
	if r.brightness != nil {
		d.Exposes["brightness"] = r.brightness
	}
	return d, nil
}

func (r *actionConfigRegistrar) RetrieveEntityData(id, property string) (*devices.EntityData, error) {
	if r.brightness == nil || property != "brightness" {
		return nil, fmt.Errorf("missing entity")
	}
	return r.brightness.Data, nil
}

func TestStepActionConfigurationBinding(t *testing.T) {
	a := automations.NewStepAction()
	a.Id = "light"
	a.Property = "brightness"
	a.Data = float64(2)
	a.Steps = []*automations.Step{{Id: "light", Property: "brightness", Operator: "+"}}
	var calls atomic.Int32
	registrars := []*actionConfigRegistrar{}
	clients := []*delayedLifecycleClient{}
	for i, name := range []string{"old", "new"} {
		entity := devices.NewEntity("brightness")
		entity.Data.SetValue(float64(10 + 100*i))
		entity.Attributes["min"] = float64(0)
		entity.Attributes["max"] = float64(255)
		registrars = append(registrars, &actionConfigRegistrar{name: name, brightness: entity})
		want := float64(12 + 100*i)
		clients = append(clients, &delayedLifecycleClient{publish: func(topic string, payload interface{}) {
			calls.Add(1)
			var data map[string]float64
			if err := json.Unmarshal(payload.([]byte), &data); err != nil {
				t.Error(err)
				return
			}
			if topic != name+"/set" || data["brightness"] != want {
				t.Errorf("mixed step binding: %s %s, want %v", topic, payload, want)
			}
		}})
	}
	if err := a.Configure(registrars[0], clients[0]); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			if err := a.Configure(registrars[i%2], clients[i%2]); err != nil {
				t.Error(err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		ctx := automations.NewRunContext(automations.NewDeviceContext(), nil, false)
		for i := 0; i < 500; i++ {
			if err := a.Execute(ctx); err != nil {
				t.Error(err)
			}
		}
	}()
	wg.Wait()
	if calls.Load() != 500 {
		t.Fatalf("want 500 commands, got %d", calls.Load())
	}
	if err := a.Configure(&actionConfigRegistrar{fail: true}, clients[0]); err == nil {
		t.Fatal("expected failed reconfiguration")
	}
	if err := a.Execute(automations.NewRunContext(automations.NewDeviceContext(), nil, false)); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 501 {
		t.Fatal("failed step configuration lost working binding")
	}
}

func TestActionUnconfiguredExecutionReturnsError(t *testing.T) {
	for _, a := range []automations.MqttAction{automations.NewTriggerAction(), automations.NewStepAction(), automations.NewPresetCyclingAction()} {
		if err := a.Execute(automations.NewRunContext(automations.NewDeviceContext(), nil, false)); err == nil {
			t.Error("unconfigured action accepted")
		}
	}
}

type actionConfigJSONValue struct{ encode func() ([]byte, error) }

func (v actionConfigJSONValue) MarshalJSON() ([]byte, error) { return v.encode() }

func TestTriggerConfigurationJSONCompatibilityAndReentry(t *testing.T) {
	for _, exposes := range [][]*automations.MqttTriggerActionExpose{nil, {}, {nil}} {
		a := automations.NewTriggerAction()
		a.Id = "sensor"
		a.Exposes = exposes
		data, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			t.Fatal(err)
		}
		if len(fields) != 3 {
			t.Fatalf("unexpected empty trigger fields: %s", data)
		}
		var decoded automations.MqttTriggerAction
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.Id != "sensor" || decoded.Type != automations.TriggerAction || len(decoded.Exposes) != len(exposes) || (decoded.Exposes == nil) != (exposes == nil) {
			t.Fatalf("changed nil/empty wire: %s", data)
		}
	}
	a := automations.NewTriggerAction()
	a.Id = "sensor"
	a.Delay = utils.IntervalFromSeconds(1)
	a.PublishMode = automations.PublishSingle
	b := &devices.BridgeInfo{}
	b.Definition.Exposes = []devices.BridgeExpose{{Property: "one", Type: "numeric"}}
	reg := &actionConfigRegistrar{name: "sensor", bridge: b}
	client := &delayedLifecycleClient{publish: func(string, interface{}) {}}
	a.Exposes = []*automations.MqttTriggerActionExpose{{Name: "one", Data: actionConfigJSONValue{encode: func() ([]byte, error) {
		if err := a.Configure(reg, client); err != nil {
			return nil, err
		}
		return []byte("42"), nil
	}}}}
	if err := a.Configure(reg, client); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		data, err := json.Marshal(a)
		if err != nil {
			done <- err
			return
		}
		var decoded automations.MqttTriggerAction
		if err := json.Unmarshal(data, &decoded); err != nil {
			done <- err
			return
		}
		if decoded.Delay == nil || decoded.Delay.Duration() != time.Second || decoded.PublishMode != automations.PublishSingle || decoded.Exposes[0].Data != float64(42) {
			done <- fmt.Errorf("wire changed: %s", data)
			return
		}
		done <- nil
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("encoding reentry blocked on configuration lock")
	}
}

func actionConfigBridge(off any) *devices.BridgeInfo {
	b := &devices.BridgeInfo{}
	b.Definition.Exposes = []devices.BridgeExpose{{Property: "presence", Type: "binary", ValueOff: off, ValueOn: true}}
	return b
}

func TestActionConfigurationFailureRetainsWorkingCommand(t *testing.T) {
	a := automations.NewTriggerAction()
	a.Id = "sensor"
	a.Exposes = []*automations.MqttTriggerActionExpose{{Name: "presence", Data: false}}
	calls := 0
	client := &delayedLifecycleClient{publish: func(topic string, payload interface{}) {
		calls++
		if topic != "old/set" || string(payload.([]byte)) != `{"presence":false}` {
			t.Errorf("failed configuration changed command: %s %s", topic, payload)
		}
	}}
	if err := a.Configure(&actionConfigRegistrar{name: "old", bridge: actionConfigBridge(false)}, client); err != nil {
		t.Fatal(err)
	}
	if err := a.Configure(&actionConfigRegistrar{name: "new", fail: true, bridge: actionConfigBridge("OFF")}, client); err == nil {
		t.Fatal("expected lookup failure")
	}
	if a.Exposes[0].Data != false {
		t.Error("failed configuration changed persisted recipe")
	}
	invalid := &devices.BridgeInfo{}
	invalid.Definition.Exposes = []devices.BridgeExpose{{Property: "presence", Type: "enum", Values: []string{"other"}}}
	if err := a.Configure(&actionConfigRegistrar{name: "invalid", bridge: invalid}, client); err == nil {
		t.Fatal("expected sanitization failure")
	}
	if err := a.Execute(automations.NewRunContext(automations.NewDeviceContext(), nil, false)); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("want one command, got %d", calls)
	}
}

func TestActionConcurrentConfigurationAndExecution(t *testing.T) {
	a := automations.NewTriggerAction()
	a.Id = "sensor"
	a.Exposes = []*automations.MqttTriggerActionExpose{{Name: "presence", Data: false}}
	var calls atomic.Int32
	clients := []*delayedLifecycleClient{}
	registrars := []*actionConfigRegistrar{}
	for _, name := range []string{"old", "new"} {
		clients = append(clients, &delayedLifecycleClient{publish: func(topic string, payload interface{}) {
			calls.Add(1)
			if topic != name+"/set" || string(payload.([]byte)) != `{"presence":false}` {
				t.Errorf("mixed runtime configuration: %s %s", topic, payload)
			}
			data, err := json.Marshal(a)
			if err != nil {
				t.Error(err)
				return
			}
			var decoded automations.MqttTriggerAction
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Error(err)
				return
			}
			if decoded.Id != "sensor" || decoded.Type != automations.TriggerAction || len(decoded.Exposes) != 1 || decoded.Exposes[0].Data != false {
				t.Errorf("changed trigger wire shape: %s", data)
			}
		}})
		registrars = append(registrars, &actionConfigRegistrar{name: name, bridge: actionConfigBridge(false)})
	}
	if err := a.Configure(registrars[0], clients[0]); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			if err := a.Configure(registrars[i%2], clients[i%2]); err != nil {
				t.Error(err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		ctx := automations.NewRunContext(automations.NewDeviceContext(), nil, false)
		for i := 0; i < 500; i++ {
			if err := a.Execute(ctx); err != nil {
				t.Error(err)
			}
		}
	}()
	wg.Wait()
	if calls.Load() != 500 {
		t.Fatalf("lost executions: %d", calls.Load())
	}
}

func TestActionConfigurationReentryKeepsSinglePublishBinding(t *testing.T) {
	a := automations.NewTriggerAction()
	a.Id = "sensor"
	a.PublishMode = automations.PublishSingle
	a.Exposes = []*automations.MqttTriggerActionExpose{{Name: "one", Data: 1}, {Name: "two", Data: 2}}
	b := &devices.BridgeInfo{}
	b.Definition.Exposes = []devices.BridgeExpose{{Property: "one", Type: "numeric"}, {Property: "two", Type: "numeric"}}
	oldReg, newReg := &actionConfigRegistrar{name: "old", bridge: b}, &actionConfigRegistrar{name: "new", bridge: b}
	oldCount, newCount := 0, 0
	newClient := &delayedLifecycleClient{publish: func(topic string, payload interface{}) {
		newCount++
		if topic != "new/set" {
			t.Errorf("wrong new topic: %s", topic)
		}
	}}
	oldClient := &delayedLifecycleClient{publish: func(topic string, payload interface{}) {
		oldCount++
		if topic != "old/set" {
			t.Errorf("wrong old topic: %s", topic)
		}
		if oldCount == 1 {
			if err := a.Configure(newReg, newClient); err != nil {
				t.Error(err)
			}
		}
	}}
	if err := a.Configure(oldReg, oldClient); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- a.Execute(automations.NewRunContext(automations.NewDeviceContext(), nil, false)) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("configuration callback deadlocked")
	}
	if oldCount != 2 || newCount != 0 {
		t.Fatalf("split execution across bindings: %d/%d", oldCount, newCount)
	}
	if err := a.Execute(automations.NewRunContext(automations.NewDeviceContext(), nil, false)); err != nil {
		t.Fatal(err)
	}
	if newCount != 2 {
		t.Fatalf("replacement not used: %d", newCount)
	}
}

func (c *delayedLifecycleClient) Publish(topic string, payload interface{}) {
	c.publish(topic, payload)
}

func delayedLifecycleAction(t *testing.T, client mqtt.MqttClient, delay *utils.TimeInterval) *automations.MqttTriggerAction {
	t.Helper()
	device := utils_test.CreatePresenceDevice("sensor", "sensor", "presence", false)
	store := utils_test.CreateStoreFromDeviceRepo(repository.NewMemoryDeviceRepo())
	registrar := services.NewHubRegisterService(store, &mocks.MockEventHub{}, 30000)
	registrar.RegisterBridge(utils_test.CreateBridgeInfoList([]*devices.Device{device}))
	action := automations.NewTriggerAction()
	action.Id = device.Id
	action.Delay = delay
	action.Exposes = []*automations.MqttTriggerActionExpose{{Name: "presence", Data: false}}
	if err := action.Configure(registrar, client); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(action.Stop)
	return action
}

func waitDelayedLifecycle(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(3 * time.Second):
		t.Fatal("delayed action did not reach expected phase")
	}
}

func TestDelayedActionConcurrentExecuteStop(t *testing.T) {
	var publishes atomic.Int32
	client := &delayedLifecycleClient{publish: func(string, interface{}) { publishes.Add(1) }}
	action := delayedLifecycleAction(t, client, utils.IntervalFromHours(1))
	ctx := automations.NewRunContext(automations.NewDeviceContext(), nil, false)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				if err := action.Execute(ctx); err != nil {
					t.Error(err)
				}
				action.Stop()
				action.Stop()
			}
		}()
	}
	wg.Wait()
	action.Stop()
	if got := publishes.Load(); got != 0 {
		t.Fatalf("cancelled long delay published %d commands", got)
	}
}

func TestDelayedActionCancelRestartAndCallbackStop(t *testing.T) {
	var action *automations.MqttTriggerAction
	var count atomic.Int32
	published := make(chan struct{}, 10)
	client := &delayedLifecycleClient{publish: func(string, interface{}) {
		count.Add(1)
		action.Stop() // Re-entry from a committed publish must not deadlock.
		published <- struct{}{}
	}}
	action = delayedLifecycleAction(t, client, utils.IntervalFromMilliseconds(50))
	ctx := automations.NewRunContext(automations.NewDeviceContext(), nil, false)
	if err := action.Execute(ctx); err != nil {
		t.Fatal(err)
	}
	action.Stop() // Cancel before expiry, then restart without changing the delay.
	action.Stop()
	for i := 0; i < 20; i++ {
		if err := action.Execute(ctx); err != nil {
			t.Fatal(err)
		}
	}
	waitDelayedLifecycle(t, published)
	select {
	case <-published:
		t.Fatal("cancelled or duplicate run published")
	case <-time.After(100 * time.Millisecond):
	}
	if got := count.Load(); got != 1 {
		t.Fatalf("want one restart command, got %d", got)
	}
}

func TestDelayedActionRejectsInvalidDuration(t *testing.T) {
	for _, delay := range []*utils.TimeInterval{
		utils.IntervalFromMilliseconds(0), utils.IntervalFromSeconds(-1),
		{Value: 1, Unit: "unknown"},
	} {
		action := automations.NewTriggerAction()
		action.Delay = delay
		if err := action.Execute(automations.NewRunContext(automations.NewDeviceContext(), nil, false)); err == nil {
			t.Errorf("invalid delay accepted: %+v", delay)
		}
		action.Stop()
	}
}

func TestDelayedActionRestartDoesNotLoseReplacement(t *testing.T) {
	first, second, extra := make(chan struct{}, 1), make(chan struct{}, 1), make(chan struct{}, 10)
	releaseFirst, releaseSecond := make(chan struct{}), make(chan struct{})
	var count atomic.Int32
	client := &delayedLifecycleClient{publish: func(topic string, payload interface{}) {
		if topic != "sensor/set" || string(payload.([]byte)) != `{"presence":false}` {
			t.Errorf("changed command: %s %s", topic, payload)
		}
		switch count.Add(1) {
		case 1:
			first <- struct{}{}
			<-releaseFirst
		case 2:
			second <- struct{}{}
			<-releaseSecond
		default:
			extra <- struct{}{}
		}
	}}
	// Always release blocked callbacks even if an assertion fails.
	t.Cleanup(func() { close(releaseFirst); close(releaseSecond) })
	action := delayedLifecycleAction(t, client, utils.IntervalFromMilliseconds(10))
	ctx := automations.NewRunContext(automations.NewDeviceContext(), nil, false)
	if err := action.Execute(ctx); err != nil {
		t.Fatal(err)
	}
	waitDelayedLifecycle(t, first)
	action.Stop() // Already publishing; must not join that callback.
	if err := action.Execute(ctx); err != nil {
		t.Fatal(err)
	}
	waitDelayedLifecycle(t, second)
	// Let old completion run while the replacement is still committed/pending.
	releaseFirst <- struct{}{}
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	deadline := time.NewTimer(100 * time.Millisecond)
	defer deadline.Stop()
	for {
		select {
		case <-tick.C:
			if err := action.Execute(ctx); err != nil {
				t.Fatal(err)
			}
		case <-extra:
			t.Fatal("old completion cleared replacement; duplicate command published")
		case <-deadline.C:
			if got := count.Load(); got != 2 {
				t.Fatalf("want two commands, got %d", got)
			}
			return
		}
	}
}

func TestTriggerWithNoConditionsCallsAction(t *testing.T) {
	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	triggerId := "button switch"

	exposeNames := []string{"brightness", "state", "color_brightness"}
	entities := []*devices.Entity{}
	for _, e := range exposeNames {
		if e == "state" {
			e := utils_test.CreateEntity(e, "binary", false)
			entities = append(entities, e)

		} else {
			e := utils_test.CreateNumericEntity(e, nil)
			entities = append(entities, e)
		}
	}

	device := utils_test.CreateDeviceWithExposes(triggerId, "dial device", entities)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})
	registrar.RegisterBridge(deviceBridgeList)

	testTriggerData := map[string]string{}
	testTriggerData["buttonSwitch1"] = "brightness"
	testTriggerData["buttonSwitch2"] = "state"
	testTriggerData["buttonSwitch3"] = "color_brightness"

	switch1Trigger := createSwitchTriggerWithBindingAction(triggerId, registrar, "buttonSwitch1", "brightness", 120, mqtt)
	switch2Trigger := createSwitchTriggerWithBindingAction(triggerId, registrar, "buttonSwitch2", "state", true, mqtt)
	switch3Trigger := createSwitchTriggerWithBindingAction(triggerId, registrar, "buttonSwitch3", "color_brightness", 250, mqtt)

	// create device trigger
	deviceTrigger := automations.NewDevice(triggerId)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, switch1Trigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, switch2Trigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, switch3Trigger)

	testCases := []struct {
		triggeredEntity string
		value           any
		result          bool
	}{
		{triggeredEntity: "buttonSwitch1", value: 120.0, result: true},
		{triggeredEntity: "buttonSwitch2", value: true, result: true},
		{triggeredEntity: "door_state", value: true, result: false},
		{triggeredEntity: "co2", value: 70, result: false},
		{triggeredEntity: "quality_index", value: 30.1, result: false},
		{triggeredEntity: "buttonSwitch3", value: 250.0, result: true},
		{triggeredEntity: "temperature", value: 25.1, result: false},
		{triggeredEntity: "humdity", value: 35.1, result: false},
	}

	for _, testCase := range testCases {
		var data = map[string]any{
			"data1":                  false,
			"data2":                  90,
			testCase.triggeredEntity: testCase.value,
		}
		var messageHandler = func(id string, payload []byte) {
			if !testCase.result {
				t.Fatalf("invalid msg received")
			}

			wg.Done()
			var msg map[string]interface{}
			json.Unmarshal([]byte(payload), &msg)
			property := testTriggerData[testCase.triggeredEntity]
			if msg[property] != testCase.value {
				t.Fatalf("invalid msg: received want %s got %s", testCase.value, msg[testCase.triggeredEntity])
			}
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}
		device.Exposes = createExposures(data)

		deviceTrigger.EvaluateTrigger(automations.NewManualEvent(device), testCase.triggeredEntity)
		mqtt.WaitResponses()
	}

	wg.Wait()
}

func TestAutomationwithMultipleTriggerActions(t *testing.T) {
	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}
	triggerId := "Attic light"

	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	exposeNames := []string{"brightness", "color_temperature", "color_brightness"}
	entities := []*devices.Entity{}
	for _, e := range exposeNames {
		e := utils_test.CreateNumericEntity(e, nil)
		entities = append(entities, e)

	}
	device := utils_test.CreateDeviceWithExposes(triggerId, "sensor device", entities)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})
	registrar.RegisterBridge(deviceBridgeList)

	trigger := createTriggerwithMultipleActions(triggerId, registrar, mqtt, "button1", []string{"brightness", "color_temperature", "color_brightness"}, []any{120, 250, 200})
	trigger.Conditions = append(trigger.Conditions, utils_test.NewExposeCondition("button1", "pressed", "="))

	automation := automations.NewDevice(triggerId)
	automation.Triggers = append(automation.Triggers, trigger)

	testCases := []struct {
		triggeredEntity string
		value           any
		result          bool
	}{
		{triggeredEntity: "button1", value: "pressed", result: true},
		{triggeredEntity: "button1", value: "release", result: false},
		{triggeredEntity: "button1", value: "pressed", result: true},
	}

	for _, testCase := range testCases {
		var data = map[string]any{
			"some_data1":             false,
			"some_data2":             90,
			testCase.triggeredEntity: testCase.value,
		}
		var messageHandler = func(id string, payload []byte) {

			for range trigger.Actions {
				data := unpackJsonToMap(string(payload))
				if data == nil {
					t.Fatalf("error unpacking json")
				}
				// triggerAction, ok := action.(*automations.MqttTriggerAction)
				// if !ok {
				// 	t.Fatalf("invalid action type")
				// }

			}

			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)

		if testCase.result {
			wg.Add(1)
		}
		device.Exposes = createExposures(data)

		automation.Evaluate(automations.NewDeviceEvent(device, data))
		mqtt.WaitResponses()
	}

	wg.Wait()
}

func TestHandleMultipleSameValueTriggerWithDelay(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	triggerId := "human sensor"
	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	device := utils_test.CreatePresenceDevice(triggerId, "sensor device", "presence", false)

	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	registrar.RegisterBridge(deviceBridgeList)

	turnOffTrigger := createTriggerDelayTurnOffLight(triggerId, registrar, mqtt, utils.IntervalFromSeconds(3))
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(triggerId, registrar, mqtt, 30)

	// create device trigger
	deviceTrigger := automations.NewDevice(triggerId)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	testCases := []struct {
		presence bool
		lux      any
		result   bool
	}{
		{presence: false, lux: 30, result: true},
		{presence: false, lux: 15, result: false},
		{presence: false, lux: 6, result: false},
	}

	for _, testCase := range testCases {
		var data = map[string]any{
			"presence": testCase.presence,
			"lux":      testCase.lux,
		}

		var messageHandler = func(id string, payload []byte) {

			for _, action := range turnOnTrigger.Actions {

				data := unpackJsonToMap(string(payload))
				if data == nil {
					t.Fatalf("error unpacking json")
				}
				triggerAction, ok := action.(*automations.MqttTriggerAction)
				if !ok {
					t.Fatalf("invalid action type")
				}

				for _, expose := range triggerAction.Exposes {
					value, ok := data[expose.Name]
					if !ok {
						t.Fatalf("property not %s found in payload", expose.Name)
					}
					if value != testCase.presence {
						t.Fatalf("value mismatch: want %v got %v", testCase.presence, value)
					}
				}
			}

			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}

		device.Exposes = createExposures(data)
		deviceTrigger.EvaluateTrigger(automations.NewManualEvent(device), "presence")

		time.Sleep(500 * time.Millisecond)
	}

	wg.Wait()
}

func TestTurnOnAndOffLightFromPresence(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	id := "human sensor"
	device := utils_test.CreatePresenceDevice(id, "sensor device", "presence", false)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(deviceBridgeList)
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(id, registrar, mqtt, 30)
	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(id, registrar, mqtt, utils.IntervalFromMilliseconds(100))

	// create device trigger
	deviceTrigger := automations.NewDevice(id)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	testCases := []struct {
		presence   bool
		sleepdelay time.Duration
		lux        any
		result     bool
	}{
		{presence: false, sleepdelay: 200, lux: 30, result: true},
		{presence: false, sleepdelay: 100, lux: 30, result: false},
		{presence: false, sleepdelay: 100, lux: 30, result: false},
		{presence: true, sleepdelay: 100, lux: 30, result: true},
		{presence: false, sleepdelay: 200, lux: 30, result: true},
		{presence: true, sleepdelay: 100, lux: 30, result: true},
		{presence: false, sleepdelay: 200, lux: 30, result: true},
		{presence: true, sleepdelay: 100, lux: 30.1, result: false},
		{presence: true, sleepdelay: 100, lux: 29.9, result: true},
		{presence: false, sleepdelay: 200, lux: 30, result: true},
		{presence: true, sleepdelay: 100, lux: 30.1, result: false},
		{presence: true, sleepdelay: 100, lux: 29, result: true},
		{presence: true, sleepdelay: 100, lux: 7, result: false},
		{presence: true, sleepdelay: 100, lux: 15, result: false},
	}

	for _, testCase := range testCases {
		var data = map[string]any{
			"presence": testCase.presence,
			"lux":      testCase.lux,
		}

		var messageHandler = func(id string, payload []byte) {

			if !testCase.result {
				t.Fatalf("invalid msg received")
			}

			for _, action := range turnOnTrigger.Actions {

				data := unpackJsonToMap(string(payload))
				if data == nil {
					t.Fatalf("error unpacking json")
				}
				triggerAction, ok := action.(*automations.MqttTriggerAction)
				if !ok {
					t.Fatalf("invalid action type")
				}

				for _, expose := range triggerAction.Exposes {
					value, ok := data[expose.Name]
					if !ok {
						t.Fatalf("property not %s found in payload", expose.Name)
					}
					if value != testCase.presence {
						t.Fatalf("value mismatch: want %v got %v", testCase.presence, value)
					}
				}
			}

			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}

		device.Exposes = createExposures(data)
		deviceTrigger.Evaluate(automations.NewDeviceEvent(device, data))

		time.Sleep(testCase.sleepdelay * time.Millisecond)
	}

	wg.Wait()
}

func TestActionWithTimerRangeConditionLightFromPresence(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now().UTC()
	})

	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	id := "human sensor"
	device := utils_test.CreatePresenceDevice(id, "sensor device", "presence", false)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(deviceBridgeList)
	// create turn on trigger
	turnOnTrigger := createTriggerTurnOnLight(id, registrar, mqtt)

	// initialize turn on condition with timer
	onTimeRange := automations.NewTimeRange("11:00", "17:00")
	turnOnCondition := utils_test.NewExposeCondition("presence", true, "=")
	turnOnTimerCondition := utils_test.NewTimeCondition(onTimeRange, mockClock)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnTimerCondition)

	// create turn off trigger
	turnOffTrigger := createTriggerDelayTurnOffLight(id, registrar, mqtt, nil)

	// initialize turn off condition with timer
	offTimeRange := automations.NewTimeRange("09:00", "13:25")
	turnOffCondition := utils_test.NewExposeCondition("presence", false, "=")
	turnOffTimerCondition := utils_test.NewTimeCondition(offTimeRange, mockClock)
	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffCondition)
	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffTimerCondition)

	// create device trigger
	deviceTrigger := automations.NewDevice(id)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)

	testCases := []struct {
		presence   bool
		sleepdelay time.Duration
		timeNow    time.Time
		result     bool
	}{
		{presence: true, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(11, 0, 0), result: true},
		{presence: false, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(12, 15, 0), result: true},
		{presence: true, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(17, 1, 0), result: false},
		{presence: true, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(13, 24, 0), result: true},
		{presence: false, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(6, 25, 0), result: false},
		{presence: false, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(9, 1, 0), result: true},
		{presence: true, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(10, 59, 0), result: false},
		{presence: true, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(12, 25, 0), result: true},
		{presence: false, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(8, 59, 0), result: false},
		{presence: false, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(11, 25, 0), result: true},
	}

	for _, testCase := range testCases {

		loc, _ := time.LoadLocation("Europe/London")
		now := time.Date(testCase.timeNow.Year(), testCase.timeNow.Month(), testCase.timeNow.Day(),
			testCase.timeNow.Hour(), testCase.timeNow.Minute(), testCase.timeNow.Second(),
			testCase.timeNow.Nanosecond(), loc)
		mockClock.SetMockTime(now.UTC())

		var messageHandler = func(id string, payload []byte) {

			if !testCase.result {
				t.Fatalf("invalid msg received")
			}

			for _, action := range turnOnTrigger.Actions {

				data := unpackJsonToMap(string(payload))
				if data == nil {
					t.Fatalf("error unpacking json")
				}
				triggerAction, ok := action.(*automations.MqttTriggerAction)
				if !ok {
					t.Fatalf("invalid action type")
				}
				if !ok {
					t.Fatalf("invalid action type")
				}

				for _, expose := range triggerAction.Exposes {
					value, ok := data[expose.Name]
					if !ok {
						t.Fatalf("property not %s found in payload", expose.Name)
					}
					if value != testCase.presence {
						t.Fatalf("value mismatch: want %v got %v", testCase.presence, value)
					}
					if value != testCase.presence {
						t.Fatalf("value mismatch: want %v got %v", testCase.presence, value)
					}
				}
			}

			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}

		var payload = map[string]any{
			"presence": testCase.presence,
		}

		device.Exposes = createExposures(payload)
		deviceTrigger.Evaluate(automations.NewDeviceEvent(device, payload))
		time.Sleep(testCase.sleepdelay * time.Millisecond)
	}

	wg.Wait()
}

func TestManualTrigger_TurnsOnLight(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	id := "Light attic"
	device := utils_test.CreatePresenceDevice(id, "light device", "presence", false)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(deviceBridgeList)

	turnOnTrigger := createTriggerTurnOnLight(id, registrar, mqtt)

	deviceTrigger := automations.NewDevice(id)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	testCases := []struct {
		presence bool
		result   bool
	}{
		{presence: true, result: true},
		{presence: false, result: true},
		{presence: true, result: true},
		{presence: false, result: true},
	}
	for _, testCase := range testCases {

		publishCount := 0
		var messageHandler = func(id string, payload []byte) {
			publishCount++
			if testCase.result {
				wg.Done()
				for _, action := range turnOnTrigger.Actions {

					responseData := unpackJsonToMap(string(payload))
					if responseData == nil {
						t.Fatalf("error unpacking json")
					}
					triggerAction, ok := action.(*automations.MqttTriggerAction)
					if !ok {
						t.Fatalf("invalid action type")
					}
					for _, expose := range triggerAction.Exposes {
						value, ok := responseData[expose.Name]
						if !ok {
							t.Fatalf("property not %s found in payload", expose.Name)
						}
						if value != expose.Data {
							t.Fatalf("value mismatch: want %v got %v", expose.Data, value)
						}
					}
				}
			} else {
				t.Fatalf("unexpected message published when result should be false")
			}
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}

		var payload = map[string]any{
			"presence": testCase.presence,
		}
		device.Exposes = createExposures(payload)
		deviceTrigger.EvaluateTrigger(automations.NewManualEvent(device), turnOnTrigger.Name)
		wg.Wait()
	}
}

func TestManualTriggerWithScheduleTurnsOnLight(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now().UTC()
	})
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	id := "Light attic"
	device := utils_test.CreatePresenceDevice(id, "Attic Light", "state", false)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(deviceBridgeList)
	// create turn on trigger
	onTimeRange := automations.NewTimeRange("11:00", "17:00")

	turnOnTrigger := createDeviceTriggerWithScheduleTurnOnLight(id, onTimeRange, registrar, mqtt, mockClock)
	deviceTrigger := automations.NewDevice(id)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	testCases := []struct {
		timeNow time.Time
		result  bool
	}{
		{timeNow: utils_test.CreateTimeFrom(10, 59, 0), result: false},
		{timeNow: utils_test.CreateTimeFrom(11, 0, 0), result: true},
		{timeNow: utils_test.CreateTimeFrom(12, 15, 0), result: true},
		{timeNow: utils_test.CreateTimeFrom(13, 36, 0), result: true},
		{timeNow: utils_test.CreateTimeFrom(14, 7, 10), result: true},
		{timeNow: utils_test.CreateTimeFrom(17, 0, 0), result: true},
		{timeNow: utils_test.CreateTimeFrom(17, 0, 1), result: false},
		{timeNow: utils_test.CreateTimeFrom(18, 10, 30), result: false},
	}

	for _, testCase := range testCases {
		loc, _ := time.LoadLocation("Europe/London")
		now := time.Date(testCase.timeNow.Year(), testCase.timeNow.Month(), testCase.timeNow.Day(),
			testCase.timeNow.Hour(), testCase.timeNow.Minute(), testCase.timeNow.Second(),
			testCase.timeNow.Nanosecond(), loc)

		mockClock.SetMockTime(now)

		var data = map[string]any{
			"presence": true,
		}

		var messageHandler = func(id string, payload []byte) {
			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}

		device.Exposes = createExposures(data)
		deviceTrigger.EvaluateTrigger(automations.NewManualEvent(device), turnOnTrigger.Name)

		wg.Wait()
	}
}

func TestSwitch(t *testing.T) {

	// todo
}

func TestEqualityChecks(t *testing.T) {
	testCases := []struct {
		op     utils.EqualityOperator
		value1 any
		value2 any
		result bool
	}{
		{op: ">", value1: 1.1, value2: 1, result: true},
		{op: ">=", value1: 1, value2: 1, result: true},
		{op: "<", value1: 1, value2: 1.1, result: true},
		{op: "<=", value1: 1, value2: 1, result: true},
		{op: "=", value1: 2, value2: 2, result: true},

		{op: ">", value1: 1, value2: 1.1, result: false},
		{op: ">=", value1: 1, value2: 1.1, result: false},
		{op: "<", value1: 1.1, value2: 1, result: false},
		{op: "<=", value1: 1.1, value2: 1, result: false},
		{op: "=", value1: 1, value2: 2, result: false},
	}

	for _, testCase := range testCases {
		res := utils.EqualityOperators[testCase.op](testCase.value1, testCase.value2)
		if res != testCase.result {
			t.Fatalf("operation result mismatch: want %v got %v in  %v %s %v", testCase.result, res, testCase.value1, testCase.op, testCase.value2)
		}
	}
}

func unpackJsonToMap(value string) map[string]any {
	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		fmt.Print("ERROR unpacking ", value)
	}
	return payload
}

func createTriggerTurnOnLightWithPresenceOnAndLux(id string, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, lux any) *automations.DeviceTrigger {
	// action = turn off light
	turnOnAction := automations.NewTriggerAction()
	turnOnAction.Id = id
	turnOnAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "presence",
			Data: true,
		},
	}
	turnOnAction.Delay = nil
	turnOnAction.Configure(registrar, mqtt)

	// Turn on sensor trigger
	turnOnTrigger := automations.NewDeviceTrigger("presence")
	turnOnTrigger.Actions = []automations.MqttAction{turnOnAction}

	// condition = presence = off && lux <= 30
	turnOnCondition := utils_test.NewExposeCondition("presence", true, "=")
	luxCondition := utils_test.NewExposeCondition("lux", lux, "<=")

	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, luxCondition)

	return turnOnTrigger
}

func createDeviceTriggerWithScheduleTurnOnLight(id string, timeRange *automations.TimeRange, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, clock utils.Clock) *automations.DeviceTrigger {

	// Turn on sensor trigger
	turnOnTrigger := automations.NewDeviceTrigger(id)

	// action = turn off light
	turnOnAction := automations.NewTriggerAction()
	turnOnAction.Id = id
	turnOnAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "state",
			Data: true,
		},
	}
	turnOnAction.Delay = nil
	turnOnAction.Configure(registrar, mqtt)
	turnOnTrigger.Actions = []automations.MqttAction{turnOnAction}

	// schedule condition
	scheduleCondition := utils_test.NewTimeCondition(timeRange, clock)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, scheduleCondition)

	return turnOnTrigger
}

func createTriggerDelayTurnOffLightWithPresenceOff(id string, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, delay *utils.TimeInterval) *automations.DeviceTrigger {

	turnOffTrigger := createTriggerDelayTurnOffLight(id, registrar, mqtt, delay)

	turnOffCondition := utils_test.NewExposeCondition("presence", false, "=")

	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffCondition)

	return turnOffTrigger
}

func createTriggerTurnOnLight(id string, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient) *automations.DeviceTrigger {
	turnOnAction := automations.NewTriggerAction()
	turnOnAction.Id = id
	turnOnAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "presence",
			Data: true,
		},
	}
	turnOnAction.Delay = nil
	turnOnAction.Configure(registrar, mqtt)

	// Turn on sensor trigger
	turnOnTrigger := automations.NewDeviceTrigger("presence")
	turnOnTrigger.Actions = []automations.MqttAction{turnOnAction}

	return turnOnTrigger
}

func createSwitchTriggerWithBindingAction(id string, registrar services.DeviceRegistrar, triggerName string, actionProp string, actionData any, mqtt mqtt.MqttClient) *automations.DeviceTrigger {
	// action = turn off light
	brightnessAction := automations.NewTriggerAction()
	brightnessAction.Id = id
	brightnessAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: actionProp,
			Data: actionData,
		},
	}
	brightnessAction.Configure(registrar, mqtt)

	// Turn off sensor trigger
	button1Trigger := automations.NewDeviceTrigger(triggerName)
	button1Trigger.Actions = []automations.MqttAction{brightnessAction}

	return button1Trigger
}

func createTriggerwithMultipleActions(id string, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, triggerName string, actions []string, data []any) *automations.DeviceTrigger {
	brightnessAction := automations.NewTriggerAction()
	brightnessAction.Id = id
	for i, action := range actions {
		brightnessAction.Exposes = append(brightnessAction.Exposes, &automations.MqttTriggerActionExpose{
			Name: action,
			Data: data[i],
		})
	}

	err := brightnessAction.Configure(registrar, mqtt)
	if err != nil {
		fmt.Println("[ERROR] configuring action", err)
		return nil
	}

	trigger := automations.NewDeviceTrigger(triggerName)
	trigger.Actions = []automations.MqttAction{brightnessAction}
	trigger.Conditions = []automations.Condition{}

	return trigger
}

func createTriggerDelayTurnOffLight(id string, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, delay *utils.TimeInterval) *automations.DeviceTrigger {
	// action = turn off light
	turnOffAction := automations.NewTriggerAction()
	turnOffAction.Id = id
	turnOffAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "presence",
			Data: false,
		},
	}

	turnOffAction.Delay = delay
	err := turnOffAction.Configure(registrar, mqtt)
	if err != nil {
		fmt.Println("error configuring delay", err)
	}

	// Turn off sensor trigger
	turnOffTrigger := automations.NewDeviceTrigger("presence")
	turnOffTrigger.Actions = []automations.MqttAction{turnOffAction}

	return turnOffTrigger
}

func createExposures(data map[string]interface{}) map[string]*devices.Entity {
	var entities = make(map[string]*devices.Entity)
	for key, value := range data {

		newEntity := createEntity(key, "", value, "", nil)
		entities[key] = newEntity
	}
	return entities
}

func createEntity(name string, description string, data any, unit string, attributes map[string]any) *devices.Entity {
	if attributes == nil {
		attributes = make(map[string]any)
	}

	newEntity := devices.NewEntity(name)
	newEntity.Attributes = map[string]any{}
	newEntity.Values = map[string]any{}
	newEntity.Data.SetValue(data)
	newEntity.Unit = unit
	newEntity.Description = description
	newEntity.Attributes = attributes
	return newEntity
}
