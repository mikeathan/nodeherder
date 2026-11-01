package automations_test

import (
	"context"
	"testing"

	"node-herder/internal/automations"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/repository"
	utils_test "node-herder/testing"
)

type captureProviderCommands struct {
	id, name string
	value    any
}

func (c *captureProviderCommands) Command(_ context.Context, id, name string, value any) error {
	c.id, c.name, c.value = id, name, value
	return nil
}

func TestProviderAutomationUsesSharedCommandBoundary(t *testing.T) {
	store := utils_test.CreateStoreFromDeviceRepo(repository.NewMemoryDeviceRepo())
	device := devices.NewDevice("provider-id")
	device.FriendlyName, device.Provider = "Provider", "esphome"
	device.SetAvailability(devices.OnlineAvailability)
	expose := devices.NewEntity("switch")
	expose.Type, expose.AccessMode = bridge.BinaryDataType, bridge.ReadWriteAccessMode
	device.Exposes[expose.Name] = expose
	if err := store.StoreDevice(device.FriendlyName, device); err != nil {
		t.Fatal(err)
	}
	registrar := services.NewHubRegisterService(store, &mocks.MockEventHub{}, 0)
	capture := &captureProviderCommands{}
	registrar.SetProviderCommands(capture)
	action := automations.NewTriggerAction()
	action.Id = device.Id
	action.Exposes = []*automations.MqttTriggerActionExpose{{Name: "switch", Data: true}}
	if err := action.Configure(registrar, &mocks.MockMqttClient{}); err != nil {
		t.Fatal(err)
	}
	state := automations.NewDeviceContext()
	if err := action.Execute(automations.NewRunContext(state, device.Exposes, false)); err != nil {
		t.Fatal(err)
	}
	if capture.id != device.Id || capture.name != "switch" || capture.value != true {
		t.Fatalf("command: %+v", capture)
	}
	if state.GetCurrentState("switch") != nil {
		t.Fatal("submitted command became confirmed state")
	}
	device.SetAvailability(devices.OfflineAvailability)
	if err := action.Execute(automations.NewRunContext(state, device.Exposes, false)); err == nil {
		t.Fatal("offline command accepted")
	}
}

func TestProviderButtonAutomationUsesNull(t *testing.T) {
	store := utils_test.CreateStoreFromDeviceRepo(repository.NewMemoryDeviceRepo())
	device := devices.NewDevice("button-provider")
	device.FriendlyName, device.Provider = "Buttons", "esphome"
	device.SetAvailability(devices.OnlineAvailability)
	entity := devices.NewEntity("restart")
	entity.Type, entity.AccessMode = "button", bridge.WriteAccessMode
	device.Exposes[entity.Name] = entity
	if err := store.StoreDevice(device.FriendlyName, device); err != nil {
		t.Fatal(err)
	}
	registrar := services.NewHubRegisterService(store, &mocks.MockEventHub{}, 0)
	capture := &captureProviderCommands{}
	registrar.SetProviderCommands(capture)
	action := automations.NewTriggerAction()
	action.Id = device.Id
	action.Exposes = []*automations.MqttTriggerActionExpose{{Name: "restart", Data: "press"}}
	if err := action.Configure(registrar, &mocks.MockMqttClient{}); err == nil {
		t.Fatal("accepted button value")
	}
	action.Exposes[0].Data = nil
	if err := action.Configure(registrar, &mocks.MockMqttClient{}); err != nil {
		t.Fatal(err)
	}
	if err := action.Execute(automations.NewRunContext(automations.NewDeviceContext(), device.Exposes, false)); err != nil {
		t.Fatal(err)
	}
	if capture.id != device.Id || capture.name != "restart" || capture.value != nil {
		t.Fatalf("command: %+v", capture)
	}
}
