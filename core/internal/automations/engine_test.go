package automations_test

import (
	"node-herder/internal/automations"
	"node-herder/mocks"
	"testing"
	"time"
)

func TestExportToFile(t *testing.T) {

	mqtt := &mocks.MockMqttClient{}
	repo := &mocks.NopRepository{}
	engine := automations.NewEngine(mqtt, repo)
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
