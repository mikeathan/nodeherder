package automation_test

import (
	"node-herder/automation"
	"testing"
)

func TestTriggers(t *testing.T) {

	trigger := new(automation.TimerTrigger)
	err := trigger.Trigger()

	if err != nil {
		t.Fatalf("error %v", err.Error())
	}

	// https://www.home-assistant.io/docs/automation/basics/
	// https://www.home-assistant.io/docs/automation/editor/
	// trigger if condition is met
	// then perform action
}
