package automations_test

import (
	automation "node-herder/transport/automations"
	"testing"
	"time"
)

func TestTriggers(t *testing.T) {

	tt := automation.TriggerFromDuration(5 * time.Second)
	tt.Repeat = false
	trigger := new(automation.TimerTrigger)
	trigger.WithCondition(tt)
	trigger.WithAction("Action triggered!!!!!!!!!!!1")
	trigger.Process()

	// https://www.home-assistant.io/docs/automation/basics/
	// https://www.home-assistant.io/docs/automation/editor/
}
