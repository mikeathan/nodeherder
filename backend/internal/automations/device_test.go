package automations_test

import (
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/models/devices"
	utils_test "node-herder/testing"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// parkingTrigger blocks the first Process call until released, so a test can
// interleave another evaluation at an exact point.
type parkingTrigger struct {
	automations.BaseTrigger
	once    sync.Once
	parked  chan struct{}
	release chan struct{}
}

func newParkingTrigger(name string) *parkingTrigger {
	t := &parkingTrigger{parked: make(chan struct{}), release: make(chan struct{})}
	t.Name = name
	t.Type = automations.DeviceTriggerType
	return t
}

func (p *parkingTrigger) Process(ctx automations.AutomationContext) error {
	first := false
	p.once.Do(func() { first = true })
	if first {
		close(p.parked)
		<-p.release
	}
	return nil
}

type recordingAction struct {
	automations.MqttBaseAction
	manualRuns atomic.Int32
	deviceRuns atomic.Int32
}

func (a *recordingAction) Execute(ctx automations.AutomationContext) error {
	if ctx.IsManualTrigger() {
		a.manualRuns.Add(1)
	} else {
		a.deviceRuns.Add(1)
	}
	return nil
}

func (a *recordingAction) Configure(services.DeviceRegistrar, mqtt.MqttClient) error { return nil }

func newIsolationAutomation() (*automations.Device, *parkingTrigger, *recordingAction) {
	parking := newParkingTrigger("state")
	action := &recordingAction{}
	unconditional := automations.NewDeviceTrigger("state")
	unconditional.Actions = []automations.MqttAction{action}

	automation := automations.NewDevice("0x70ac08fffefafeca")
	automation.Enabled = true
	automation.Triggers = automations.TriggerList{parking, unconditional}
	return automation, parking, action
}

func isolationDevice() *devices.Device {
	return utils_test.CreateDeviceWithExposes("0x70ac08fffefafeca", "Attic room Light", []*devices.Entity{
		utils_test.CreateEntity("state", "binary", "OFF"),
		utils_test.CreateEntity("linkquality", "numeric", 80.0),
	})
}

func isolationEvent(payload map[string]interface{}) *automations.DeviceEvent {
	return automations.NewDeviceEvent(isolationDevice(), payload)
}

func manualEvent() *automations.DeviceEvent {
	return automations.NewManualEvent(isolationDevice())
}

func awaitParked(t *testing.T, p *parkingTrigger) {
	t.Helper()
	select {
	case <-p.parked:
	case <-time.After(5 * time.Second):
		t.Fatal("trigger was not reached")
	}
}

// AC-08: a device message evaluated while a manual trigger is in flight must not
// stop the manual run's unconditional trigger.
func TestManualTriggerRunsDespiteConcurrentDeviceMessage(t *testing.T) {
	automation, parking, action := newIsolationAutomation()

	manualErr := make(chan error, 1)
	go func() {
		manualErr <- automation.EvaluateTrigger(manualEvent(), "state")
	}()
	awaitParked(t, parking)

	// A device message for another property arrives mid-run.
	automation.Evaluate(isolationEvent(map[string]interface{}{"linkquality": 81.0}))
	close(parking.release)

	if err := <-manualErr; err != nil {
		t.Fatal(err)
	}
	if got := action.manualRuns.Load(); got != 1 {
		t.Fatalf("manual run executed the unconditional trigger %d times, want 1", got)
	}
	if got := action.deviceRuns.Load(); got != 0 {
		t.Fatalf("unconditional trigger ran %d times as a device run, want 0", got)
	}
}

// AC-09: a device message must never run the unconditional trigger, even when a
// manual trigger fires while that device run is in flight.
func TestDeviceMessageNeverRunsUnconditionalTrigger(t *testing.T) {
	automation, parking, action := newIsolationAutomation()

	deviceDone := make(chan struct{})
	go func() {
		defer close(deviceDone)
		automation.Evaluate(isolationEvent(map[string]interface{}{"state": "ON"}))
	}()
	awaitParked(t, parking)

	if err := automation.EvaluateTrigger(manualEvent(), "state"); err != nil {
		t.Fatal(err)
	}
	close(parking.release)
	<-deviceDone

	if got := action.manualRuns.Load(); got != 1 {
		t.Fatalf("manual trigger ran the action %d times, want 1", got)
	}
	if got := action.deviceRuns.Load(); got != 0 {
		t.Fatalf("device message ran the unconditional trigger %d times, want 0", got)
	}
}
