package automations_test

import (
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/settings"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"os"
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

const (
	presenceID   = "0xa4c13894070052fc"
	presenceName = "Living room presence sensor"
	presenceLamp = "0x001788010ba4328e"
)

var presenceBridge = []byte(`[
 {"ieee_address":"` + presenceID + `","friendly_name":"` + presenceName + `","type":"Router","interview_completed":true,
  "definition":{"exposes":[
   {"type":"binary","name":"presence","property":"presence","access":5,"value_on":true,"value_off":false},
   {"type":"numeric","name":"illuminance","property":"illuminance","access":5}]}},
 {"ieee_address":"` + presenceLamp + `","friendly_name":"Living Room Light","type":"Router","interview_completed":true,
  "definition":{"exposes":[{"type":"light","features":[
   {"type":"binary","name":"state","property":"state","access":7,"value_on":"ON","value_off":"OFF","value_toggle":"TOGGLE"}]}]}}
]`)

// commandRecorder is an mqtt.MqttClient that records published commands.
type commandRecorder struct {
	commands chan string
}

func (c *commandRecorder) Connect() error                        { return nil }
func (c *commandRecorder) AddTopic(string) error                 { return nil }
func (c *commandRecorder) RemoveTopic(string) error              { return nil }
func (c *commandRecorder) Disconnect()                           {}
func (c *commandRecorder) OnMessageHandler(func(string, []byte)) {}
func (c *commandRecorder) Publish(topic string, payload interface{}) {
	c.commands <- topic + " " + string(payload.([]byte))
}

type presenceRig struct {
	automation automations.Automation
	sensor     *devices.Device
	clock      *mocks.MockClock
	commands   chan string
	lifetime   *services.DeviceLifetimeService
}

// newPresenceRig loads the real presence automation and binds every action to
// a fake clock.
func newPresenceRig(t *testing.T) *presenceRig {
	t.Helper()
	data, err := os.ReadFile("../../configs/automations/" + presenceID + ".json")
	if err != nil {
		t.Fatal(err)
	}
	serializer := automations.NewAutomationSerialiser()
	automation, err := serializer.Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}

	bridgeList, err := devices.LoadBridgeDevices(presenceBridge)
	if err != nil {
		t.Fatal(err)
	}
	registrar := services.NewHubRegisterService(utils_test.CreateStore(), &mocks.MockEventHub{}, 30000)
	registrar.RegisterBridge(bridgeList)

	recorder := &commandRecorder{commands: make(chan string, 16)}
	if err := automation.Configure(registrar, recorder); err != nil {
		t.Fatal(err)
	}
	clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 10, 7, 20, 0, 0, 0, time.UTC) })
	for _, trigger := range automation.GetTriggers() {
		for _, action := range trigger.GetActions() {
			action.(interface{ SetClock(utils.Clock) }).SetClock(clock)
		}
	}
	t.Cleanup(func() {
		for _, trigger := range automation.GetTriggers() {
			for _, action := range trigger.GetActions() {
				action.Stop()
			}
		}
	})

	sensor, err := registrar.LookupById(presenceID)
	if err != nil {
		t.Fatal(err)
	}
	r := &presenceRig{automation: automation, sensor: sensor, clock: clock, commands: recorder.commands}
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, settings.NewAppConfig(), &sync.RWMutex{})
	r.lifetime = services.NewDeviceLifetimeService(sensor, &devices.DeviceRequestEvents{
		OnDeviceAutomationTriggered: func(d *devices.Device, delta map[string]interface{}) {
			r.automation.Evaluate(automations.NewDeviceEvent(d, delta))
		},
		OnDeviceUpdated: func(*devices.Device, *devices.UpdatePackage) {},
	}, cache, mocks.NewMockAutomationDeviceQuerierWithValues(map[string]bool{presenceID: true}), clock)
	t.Cleanup(func() {
		cfg := settings.NewDeviceConfig(presenceID)
		cfg.Disabled = true
		r.lifetime.OnConfigUpdated(cfg)
	})
	return r
}

// report traverses the real acceptance/deduplication/sampling path.
func (r *presenceRig) report(payload map[string]any) { r.lifetime.Update(payload) }

func (r *presenceRig) expectCommand(t *testing.T, want string) {
	t.Helper()
	select {
	case got := <-r.commands:
		if got != want {
			t.Fatalf("published %q, want %q", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("expected %q, nothing published", want)
	}
}

// expectNothing asserts no command; delayed commands publish on their own
// goroutine, so absence is checked over a short bounded window.
func (r *presenceRig) expectNothing(t *testing.T) {
	t.Helper()
	select {
	case got := <-r.commands:
		t.Fatalf("unexpected command %q", got)
	case <-time.After(100 * time.Millisecond):
	}
}

const (
	lampOff = "Living Room Light/set {\"state\":\"OFF\"}"
	lampOn  = "Living Room Light/set {\"state\":\"ON\"}"
)

func TestPresenceConditionReportOrdering(t *testing.T) {
	// Only a presence (primary) change starts actions. Illuminance-only changes are
	// evaluated (and can cancel pending actions) but never command the light; this is
	// what stops turning the light off from re-lighting it. See spec 068.
	for _, tc := range []struct {
		name    string
		reports []map[string]any
		wantOn  int
	}{
		{"dark before presence", []map[string]any{{"presence": false, "illuminance": 100.0}, {"illuminance": 10.0}, {"presence": true}}, 1},
		{"dark with presence", []map[string]any{{"presence": false, "illuminance": 100.0}, {"presence": true, "illuminance": 10.0}}, 1},
		{"dark after presence (illuminance only)", []map[string]any{{"presence": false, "illuminance": 100.0}, {"presence": true}, {"illuminance": 10.0}}, 0},
		{"repeated full report (presence unchanged)", []map[string]any{{"presence": true, "illuminance": 100.0}, {"presence": true, "illuminance": 10.0}, {"presence": true, "illuminance": 10.0}}, 0},
		{"illuminance-only readings do not re-request", []map[string]any{{"presence": true, "illuminance": 45.0}, {"illuminance": 30.0}, {"illuminance": 0.0}, {"linkquality": 90.0}}, 1},
		{"illuminance-only readings do not recover on their own", []map[string]any{{"presence": true, "illuminance": 30.0}, {"illuminance": 100.0}, {"illuminance": 45.0}}, 1},
		{"empty room remains off", []map[string]any{{"presence": false, "illuminance": 100.0}, {"illuminance": 0.0}}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newPresenceRig(t)
			for _, report := range tc.reports {
				r.report(report)
			}
			if len(r.commands) != tc.wantOn {
				t.Fatalf("ON commands=%d, want %d", len(r.commands), tc.wantOn)
			}
			for i := 0; i < tc.wantOn; i++ {
				r.expectCommand(t, lampOn)
			}
		})
	}
}

// countingCondition wraps a real recipe condition to record how often selection
// evaluates it, so a test can prove a dependency-only report was evaluated.
type countingCondition struct {
	automations.Condition
	evaluations int
}

func (c *countingCondition) Evaluate(ctx automations.AutomationContext) bool {
	c.evaluations++
	return c.Condition.Evaluate(ctx)
}

// AC-02/AC-06: a condition-dependency change is evaluated but is not a request to
// command a device. Turning the light off darkens the room; the resulting
// illuminance-only update must not re-light it. Regression for spec 068.
func TestIlluminanceOnlyChangeDoesNotStartActions(t *testing.T) {
	r := newPresenceRig(t)
	var probe *countingCondition
	for _, trigger := range r.automation.GetTriggers() {
		deviceTrigger, ok := trigger.(*automations.DeviceTrigger)
		if !ok {
			continue
		}
		for i, condition := range deviceTrigger.Conditions {
			if condition.DependsOn("illuminance") {
				probe = &countingCondition{Condition: condition}
				deviceTrigger.Conditions[i] = probe
			}
		}
	}
	if probe == nil {
		t.Fatal("presence recipe has no illuminance condition to observe")
	}

	r.report(map[string]any{"presence": true, "illuminance": 100.0})
	r.expectNothing(t)
	before := probe.evaluations

	// Same presence, only illuminance changed (the room darkened because the light
	// was switched off). This must be evaluated but must not publish state ON.
	r.report(map[string]any{"illuminance": 39.0})
	r.expectNothing(t)
	if probe.evaluations <= before {
		t.Fatal("dependency-only change was not evaluated")
	}
}

// A late dark reading after the delayed OFF must not, by itself, relight the lamp;
// a fresh presence transition in the dark does. The darkening report fails against
// pre-fix code, which published state ON for a dependency-only change.
func TestPresenceReturnAfterDelayedOffWithLateDarkReading(t *testing.T) {
	r := newPresenceRig(t)
	r.report(map[string]any{"presence": false, "illuminance": 100.0})
	r.clock.Advance(5 * time.Minute)
	r.expectCommand(t, lampOff)

	// Someone returns while the room is still bright: primary change, condition fails.
	r.report(map[string]any{"presence": true, "illuminance": 100.0})
	r.expectNothing(t)

	// Only illuminance changed as the room darkened. Evaluated, never started.
	r.report(map[string]any{"illuminance": 10.0})
	r.expectNothing(t)

	// A fresh presence transition in the dark starts ON.
	r.report(map[string]any{"presence": false})
	r.report(map[string]any{"presence": true})
	r.expectCommand(t, lampOn)

	// Repeated report is deduplicated.
	r.report(map[string]any{"presence": true, "illuminance": 10.0})
	r.expectNothing(t)
}

func TestPresenceDependencyDoesNotRestartOrRetainOffDelay(t *testing.T) {
	r := newPresenceRig(t)
	r.report(map[string]any{"presence": false, "illuminance": 100.0})
	r.clock.Advance(3 * time.Minute)
	// Illuminance-only changes must not restart the pending OFF delay.
	r.report(map[string]any{"presence": false, "illuminance": 10.0})
	r.clock.Advance(2 * time.Minute)
	r.expectCommand(t, lampOff)

	// Illuminance-only darkening must not start ON.
	r.report(map[string]any{"presence": false, "illuminance": 0.0})
	r.expectNothing(t)

	// A primary presence change in the dark starts ON; nothing is retained afterwards.
	r.report(map[string]any{"presence": true})
	r.expectCommand(t, lampOn)

	// The lamp going off darkens an occupied room; this dependency-only report must
	// not relight it. Pre-fix it published a second ON.
	r.report(map[string]any{"illuminance": 39.0})
	r.expectNothing(t)

	r.clock.Advance(10 * time.Minute)
	r.expectNothing(t)
}

func TestConditionChangesNeverReplayCachedPhysicalEvents(t *testing.T) {
	device := utils_test.CreateDeviceWithExposes("button", "button", []*devices.Entity{
		utils_test.CreateEntity("action", "enum", "press"),
		utils_test.CreateEntity("illuminance", "numeric", 100.0),
	})
	trigger := automations.NewDeviceTrigger("action")
	trigger.Conditions = []automations.Condition{
		utils_test.NewExposeCondition("action", "press", utils.Equals),
		utils_test.NewExposeCondition("illuminance", 45.0, utils.LessThanEqual),
	}
	action := &recordingAction{}
	trigger.Actions = []automations.MqttAction{action}
	automation := automations.NewDevice("button")
	automation.Triggers = automations.TriggerList{trigger}
	automation.Evaluate(automations.NewDeviceEvent(device, map[string]interface{}{"action": "press"}))
	device.Exposes["illuminance"].Data.SetValue(10.0)
	automation.Evaluate(automations.NewDeviceEvent(device, map[string]interface{}{"illuminance": 10.0}))
	if action.deviceRuns.Load() != 0 {
		t.Fatal("condition update replayed a cached button event")
	}
	for i := 0; i < 2; i++ {
		automation.Evaluate(automations.NewDeviceEvent(device, map[string]interface{}{"action": "press"}))
	}
	if action.deviceRuns.Load() != 2 {
		t.Fatal("identical new physical events must both execute")
	}
}

func TestManualAndDeviceRunsRemainIndependent(t *testing.T) {
	r := newPresenceRig(t)
	r.report(map[string]any{"presence": true, "illuminance": 100.0})
	// Manual evaluation may overlap a device run after fresh state is accepted.
	r.sensor.Exposes["illuminance"].Data.SetValue(10.0)
	if err := r.automation.EvaluateTrigger(automations.NewManualEvent(r.sensor), "presence"); err != nil {
		t.Fatal(err)
	}
	r.expectCommand(t, lampOn)
	// A device run selected by its primary property remains independent of the manual
	// run above; dependency-only device updates no longer start actions.
	r.automation.Evaluate(automations.NewDeviceEvent(r.sensor, map[string]interface{}{"presence": true, "illuminance": 10.0}))
	r.expectCommand(t, lampOn)
}

func TestActionFailureReportingFollowsSelection(t *testing.T) {
	device := utils_test.CreateDeviceWithExposes("sensor", "sensor", []*devices.Entity{
		utils_test.CreateEntity("presence", "binary", true),
		utils_test.CreateEntity("illuminance", "numeric", 100.0),
	})
	trigger := automations.NewDeviceTrigger("presence")
	trigger.Conditions = []automations.Condition{utils_test.NewExposeCondition("illuminance", 45.0, utils.LessThanEqual)}
	trigger.Actions = []automations.MqttAction{automations.NewTriggerAction()} // deliberately unconfigured
	automation := automations.NewDevice("sensor")
	automation.Triggers = automations.TriggerList{trigger}
	// Primary selection whose condition is not met: no action, no error.
	if !automation.Evaluate(automations.NewDeviceEvent(device, map[string]interface{}{"presence": true})) {
		t.Fatal("unmatched condition is not an action error")
	}
	// Dependency-only selection whose condition is met: evaluated, never started.
	device.Exposes["illuminance"].Data.SetValue(10.0)
	if !automation.Evaluate(automations.NewDeviceEvent(device, map[string]interface{}{"illuminance": 10.0})) {
		t.Fatal("dependency-only selection started an action")
	}
	// Primary selection whose condition is met: the action runs and its failure is reported.
	if automation.Evaluate(automations.NewDeviceEvent(device, map[string]interface{}{"presence": true, "illuminance": 10.0})) {
		t.Fatal("primary action failure was concealed")
	}
	// A previous failure must not suppress the next primary selection (spec 010 AC-04).
	if automation.Evaluate(automations.NewDeviceEvent(device, map[string]interface{}{"presence": true, "illuminance": 10.0})) {
		t.Fatal("a previous failure suppressed the next primary run")
	}
}

// A condition-dependency change cancels a pending delayed action but never schedules
// a new one; only a primary selection starts actions.
func TestConditionDependencyCancelsButDoesNotStartDelayedAction(t *testing.T) {
	r := newPresenceRig(t)
	for _, trigger := range r.automation.GetTriggers() {
		for _, action := range trigger.GetActions() {
			action.(*automations.MqttTriggerAction).Delay = utils.IntervalFromSeconds(5)
		}
	}
	r.report(map[string]any{"presence": true, "illuminance": 10.0})
	r.report(map[string]any{"illuminance": 100.0})
	r.clock.Advance(10 * time.Second)
	r.expectNothing(t)
	// Condition passes again, but only illuminance changed: no action is scheduled.
	r.report(map[string]any{"illuminance": 10.0})
	r.clock.Advance(5 * time.Second)
	r.expectNothing(t)
}
