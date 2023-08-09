package automation_test

import (
	"node-herder/automation"
	"testing"
)

func TestTriggers(t *testing.T) {

	// payload coming in
	// if name matches our list of triggers
	// then
	// find trigger sensor forthat device
	// and process value

	// BUT that doesn work with timer
	// input is not payload but is timer
	// .....
	// maybe timer does not go through same flow
	// we need automation processor
	//......

	trigger := new(automation.TimerTrigger)
	trigger.WithCondition("some condition")
	trigger.WithAction("some action")
	trigger.Process()

	// https://www.home-assistant.io/docs/automation/basics/
	// https://www.home-assistant.io/docs/automation/editor/
	// trigger if condition is met
	// then perform action
}
