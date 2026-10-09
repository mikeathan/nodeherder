// T026: automation behaviour when the hub's sampled outputs fail, and when an
// automation is saved, driven through the real hub with the dial automation.
package scenarios

import (
	"errors"
	"node-herder/testing/hubharness"
	"reflect"
	"testing"
)

// AC-03, AC-11: with every broadcast and device/metrics write failing, the dial
// automation publishes exactly the golden commands: failures neither replay an
// event nor suppress the remaining outputs (each broadcast is still attempted,
// which is what the rig waits on).
func TestDialCommandsUnchangedWhenOutputsFail(t *testing.T) {
	r := newDialRig(t, dialGoldenSeed)
	r.h.EventHub.FailBroadcasts(errors.New("fixture broadcast failed"))
	r.h.Store.FailWrites(errors.New("fixture store failed"))

	r.runDialGolden(t)

	if r.h.Store.FailedWrites() == 0 {
		t.Fatal("no device or metrics write was attempted while writes were failing")
	}
}

// AC-03: saving an automation from the UI does not execute it; the next genuine
// device event still does. This catches a command published by the save itself
// or routed through the hub's lanes; it does not wait for timers the save might
// start.
func TestSavingAutomationDoesNotExecuteIt(t *testing.T) {
	r := newDialRig(t, dialGoldenSeed)

	if err := r.h.EventHub.SaveAutomation(hubharness.AutomationConfigPayload(t, dialID)); err != nil {
		t.Fatal(err)
	}
	// The save runs synchronously; anything it queued on the hub's lanes is
	// processed before the dial event, so all such commands are recorded once
	// the dial event has broadcast.
	r.h.Deliver(t, dialName, dialEvent("dial_rotate_right_slow", 20))
	r.awaitDial(t, 1)

	want := []map[string]any{{"brightness": 110.0}}
	if got := r.drainPublished(); !reflect.DeepEqual(got, want) {
		t.Fatalf("published %v, want only the dial event's command %v", got, want)
	}
}
