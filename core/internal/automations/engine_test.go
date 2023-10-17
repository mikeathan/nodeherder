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
	build mock data for mock storaget


	
	storage := NewMockStorage(nil)
	mqtt := &mocks.MockMqttClient{}
	repo := &mocks.NopRepository{}
	engine := automations.NewEngine(mqtt, repo)
	engine.WithStorage(storage)
	engine.Initialize()

	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, 100*time.Millisecond)
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(mqtt, 30.1)

	// create device trigger
	deviceTrigger := automations.NewDevice("human sensor")
	deviceTrigger.Description = "test human sensor automation"
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	err := deviceTrigger.Save("temp1", true)
	if err != nil {
		t.Fatalf("ERROR saving trigger %s", err.Error())
	}

	newTrigger, err := automations.LoadTrigger("temp1")
	if err != nil {
		t.Fatalf("ERROR laoding trigger from file %s", err.Error())
	}
	if newTrigger.Id != deviceTrigger.Id {
		t.Fatalf("ERROR Id mismatch")
	}
	if newTrigger.FriendlyName != deviceTrigger.FriendlyName {
		t.Fatalf("ERROR Name mismatch")
	}
	if newTrigger.Description != deviceTrigger.Description {
		t.Fatalf("ERROR Description mismatch")
	}
	if newTrigger.Enabled != deviceTrigger.Enabled {
		t.Fatalf("ERROR Description mismatch")
	}

	for tidx, trigger := range deviceTrigger.Triggers {
		newTrigger := newTrigger.Triggers[tidx]

		if trigger.Name != newTrigger.Name {
			t.Fatalf("ERROR Trigger.Name mismatch")
		}
		if trigger.Action.FriendlyName != newTrigger.Action.FriendlyName {
			t.Fatalf("ERROR Action.Friendlyname mismatch")
		}

		if trigger.Action.Property != newTrigger.Action.Property {
			t.Fatalf("ERROR Action.Property mismatch")
		}

		if trigger.Action.Type != newTrigger.Action.Type {
			t.Fatalf("ERROR Action.Type mismatch")
		}
		if trigger.Action.Delay != newTrigger.Action.Delay {
			t.Fatalf("ERROR Action.Delay mismatch")
		}

		for cidx, condition := range trigger.Conditions {

			newCondition := newTrigger.Conditions[cidx]

			if condition.EqualityOperator != newCondition.EqualityOperator {
				t.Fatalf("ERROR Condition.EqualityOperator mismatch")
			}
			if condition.Name != newCondition.Name {
				t.Fatalf("ERROR Condition.Name mismatch")
			}

			if condition.Value != newCondition.Value {
				t.Fatalf("ERROR Condition.Value mismatch")
			}
		}

	}

	err = automations.DeleteTrigger("temp1")

	if err != nil {
		t.Fatalf("ERROR deleting file%s", err.Error())
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
