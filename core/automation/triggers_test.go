package automation_test

import (
	"node-herder/automation"
	"testing"
	"time"
)

func TestTriggers(t *testing.T) {

	tt := automation.FromDuration(5 * time.Second)
	tt.Repeat = true
	trigger := new(automation.TimerTrigger)
	trigger.WithCondition(tt)
	trigger.WithAction("Action triggered!!!!!!!!!!!1")
	trigger.Process()

	// https://www.home-assistant.io/docs/automation/basics/
	// https://www.home-assistant.io/docs/automation/editor/
}
