package automations_test

import (
	"errors"
	"node-herder/internal/automations"
	"node-herder/mocks"
	"node-herder/utils/storage"
	"sort"
	"testing"
	"time"
)

func TestExportToFile(t *testing.T) {

	t.Skip("delete - it needs Bridgeinfo mocking which we currently dont have")
	mqtt := &mocks.MockMqttClient{}
	repo := &mocks.NopRepository{}

	//automationData := mockData(mqtt)
	storage := NewMockStorage([]*automations.Device{})

	engine := automations.NewEngine(mqtt, repo)
	engine.WithStorage(storage)
	engine.Initialize()

	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, 100*time.Millisecond)
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(mqtt, 30.1)

	// create device trigger
	inputDeviceTriggers := []*automations.Device{}
	deviceTrigger1 := automations.NewDevice("human sensor")
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

			if outputTrigger.Action.Type != inputTrigger.Action.Type {
				t.Fatalf("ERROR Action.Type mismatch")
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

type MockStorage[T any] struct {
	cache    map[string]*automations.Device
	mockData []*automations.Device
}

func NewMockStorage[T automations.Device](mockData []*automations.Device) storage.Storage[automations.Device] {
	d := new(MockStorage[automations.Device])
	d.cache = make(map[string]*automations.Device)
	d.mockData = mockData
	return d
}

func (d *MockStorage[T]) Initialize() ([]*automations.Device, error) {

	d.ClearCache()

	// initialize with mock data
	for _, mockItem := range d.mockData {
		d.Store(mockItem.Id, mockItem)
	}

	return d.LoadAll(), nil
}

func (d *MockStorage[T]) LoadAll() []*automations.Device {

	keys := make([]string, 0, len(d.cache))
	values := make([]*automations.Device, 0, len(d.cache))

	for k, _ := range d.cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		values = append(values, d.cache[k])
	}

	return values
}

func (d *MockStorage[T]) Delete(name string) error {

	d.deleteFromCache(name)
	return nil
}

func (d *MockStorage[T]) ClearCache() {

	for k := range d.cache {
		delete(d.cache, k)
	}
}

func (d *MockStorage[T]) Store(name string, item *automations.Device) error {

	d.addToCache(name, item)
	return nil
}

func (d *MockStorage[T]) Load(name string) (*automations.Device, error) {

	item := d.loadFromCache(name)
	if item != nil {
		return item, nil
	}

	for _, mockItem := range d.mockData {
		if mockItem.Id == name {
			return mockItem, nil
		}
	}

	return nil, errors.New("item not found")
}

func (d *MockStorage[T]) addToCache(name string, item *automations.Device) {
	d.cache[name] = item
}

func (d *MockStorage[T]) loadFromCache(name string) *automations.Device {
	if item, ok := d.cache[name]; ok {
		return item
	}

	return nil
}

func (d *MockStorage[T]) deleteFromCache(name string) {
	delete(d.cache, name)
}
