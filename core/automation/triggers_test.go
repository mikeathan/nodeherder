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
}
