package automations_test

import (
	"context"
	"node-herder/internal/automations"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	utils_test "node-herder/testing"
	"testing"
	"time"
)

func createBridgeInfoes() []*devices.BridgeInfo {

	dev1 := &devices.BridgeInfo{}
	dev1.IeeeAddress = "0x123456"
	dev1.Definition.Description = "Mocking human sensor"
	dev1.FriendlyName = "human sensor"
	dev1.Type = "EndDevice"
	dev1.PowerSource = "battery"
	dev1.Disabled = false
	dev1.InterviewCompleted = true

	e := devices.BridgeExpose{}
	e.Name = "presence"
	e.Property = "presence"
	e.Type = "binary"
	e.Description = "determines if presence has been detected"

	dev1.Definition.Exposes = append(dev1.Definition.Exposes, e)

	//
	dev2 := &devices.BridgeInfo{}
	dev2.IeeeAddress = "0x56789"
	dev2.Definition.Description = "Mocking Attic light"
	dev2.FriendlyName = "Attic light"
	dev2.Type = "EndDevice"
	dev2.PowerSource = "mains"
	dev2.Disabled = false
	dev2.InterviewCompleted = true

	e2 := devices.BridgeExpose{}
	e2.Type = "light"
	b2f1 := devices.BridgeInfoFeature{}
	b2f1.Description = "On/off state of this light"
	b2f1.Name = "state"
	b2f1.Property = "state"
	b2f1.Type = "binary"
	b2f1.ValueOff = "OFF"
	b2f1.ValueOn = "ON"

	b2f2 := devices.BridgeInfoFeature{}
	b2f2.Description = "Brightness of this light"
	b2f2.Name = "brightness"
	b2f2.Property = "brightness"
	b2f2.Type = "numeric"
	b2f2.ValueMax = 255
	b2f2.ValueMin = 0

	e2.Features = append(e2.Features, b2f1)
	e2.Features = append(e2.Features, b2f2)

	dev2.Definition.Exposes = append(dev2.Definition.Exposes, e2)

	return []*devices.BridgeInfo{dev1, dev2}
}

func TestExportAutomationsFromFile(t *testing.T) {

	mqtt := &mocks.MockMqttClient{}
	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	dev1 := createMockDevice("0x56789", "livingroom", "brightness", nil, 0.0, 255.0)
	dev2 := createMockDevice("0x56789", "humansensor", "left_click", nil, 0.0, 255.0)

	bridgeinfos := createBridgeInfoes()
	registrar.RegisterBridge(bridgeinfos, 60)
	registrar.Register("livingroom", dev1)
	registrar.Register("humansensor", dev2)

	storage := mocks.NewMockAutomationStorage([]*automations.Device{})

	ctx := context.Background()
	engine := automations.NewEngine(registrar, mqtt, ctx)
	engine.WithStorage(storage)
	engine.Initialize()

	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, 100*time.Millisecond)
	turnOffTrigger.Action.Id = "0x56789"
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(mqtt, 30.1)
	turnOnTrigger.Action.Id = "0x56789"

	// create device trigger
	inputDeviceTriggers := []*automations.Device{}
	deviceTrigger1 := automations.NewDevice("0x123456")
	deviceTrigger1.FriendlyName = "humansensor"
	deviceTrigger1.Description = "test human sensor automation"
	deviceTrigger1.Triggers = append(deviceTrigger1.Triggers, turnOffTrigger)
	deviceTrigger1.Triggers = append(deviceTrigger1.Triggers, turnOnTrigger)
	inputDeviceTriggers = append(inputDeviceTriggers, deviceTrigger1)

	for _, d := range inputDeviceTriggers {
		err := engine.Add(d)
		if err != nil {
			t.Fatalf("ERROR adding trigger %s", err.Error())
		}
	}

	outputDeviceTriggers := engine.GetAllTriggers()
	if len(outputDeviceTriggers) != len(inputDeviceTriggers) {
		t.Fatalf("ERROR size mismatch. want %v got %d", len(inputDeviceTriggers), len(outputDeviceTriggers))
	}
	for dIdx, outDeviceTrigger := range outputDeviceTriggers {
		inputDeviceTrigger := inputDeviceTriggers[dIdx]
		if outDeviceTrigger.Id != inputDeviceTrigger.Id {
			t.Fatalf("ERROR Id mismatch")
		}
		if outDeviceTrigger.FriendlyName != inputDeviceTrigger.FriendlyName {
			t.Fatalf("ERROR Name mismatch")
		}
		if outDeviceTrigger.Description != inputDeviceTrigger.Description {
			t.Fatalf("ERROR Description mismatch")
		}
		if outDeviceTrigger.Enabled != inputDeviceTrigger.Enabled {
			t.Fatalf("ERROR Description mismatch")
		}

		for tidx, outputTrigger := range outDeviceTrigger.Triggers {
			inputTrigger := inputDeviceTrigger.Triggers[tidx]

			if outputTrigger.Name != inputTrigger.Name {
				t.Fatalf("ERROR Trigger.Name mismatch")
			}
			if outputTrigger.Action.FriendlyName != inputTrigger.Action.FriendlyName {
				t.Fatalf("ERROR Action.Friendlyname mismatch")
			}

			if outputTrigger.Action.Property != inputTrigger.Action.Property {
				t.Fatalf("ERROR Action.Property mismatch")
			}

			if outputTrigger.Action.Delay != inputTrigger.Action.Delay {
				t.Fatalf("ERROR Action.Delay mismatch")
			}

			for cidx, outputCondition := range outputTrigger.Conditions {

				inputCondition := inputTrigger.Conditions[cidx]

				if outputCondition.EqualityOperator != inputCondition.EqualityOperator {
					t.Fatalf("ERROR Condition.EqualityOperator mismatch")
				}
				if outputCondition.Name != inputCondition.Name {
					t.Fatalf("ERROR Condition.Name mismatch")
				}

				if outputCondition.Value != inputCondition.Value {
					t.Fatalf("ERROR Condition.Value mismatch")
				}
			}
		}
	}

	for _, inputDeviceTrigger := range inputDeviceTriggers {
		err := engine.Delete(inputDeviceTrigger.Id)
		if err != nil {
			t.Fatalf("ERROR deleting file %v .Error %s", inputDeviceTrigger.Id, err.Error())
		}
	}

	inputDeviceTriggers = engine.GetAllTriggers()
	if len(inputDeviceTriggers) != 0 {
		t.Fatalf("ERROR triggers found. expecting empty")
	}
}
