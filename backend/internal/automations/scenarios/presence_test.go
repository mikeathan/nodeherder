// Scenario test of the Living room presence sensor automation using
// backend/configs/automations/0xa4c13894070052fc.json, including its delayed
// (presence off) actions on a fake clock.
package scenarios

import (
	"node-herder/internal/automations"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"os"
	"testing"
	"time"
)

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
}

// newPresenceRig loads the real presence automation and binds every action to
// a fake clock.
func newPresenceRig(t *testing.T) *presenceRig {
	t.Helper()
	data, err := os.ReadFile("../../../configs/automations/" + presenceID + ".json")
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
	return &presenceRig{automation: automation, sensor: sensor, clock: clock, commands: recorder.commands}
}

// report applies a sensor message the way Update does (store, then evaluate
// with the changed keys).
func (r *presenceRig) report(payload map[string]any) {
	for name, value := range payload {
		r.sensor.Exposes[name].Data.SetValue(value)
	}
	r.automation.Evaluate(automations.NewDeviceEvent(r.sensor, payload))
}

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

// AC-11: presence=false publishes OFF after the configured 5 minutes.
func TestPresenceOffAfterDelay(t *testing.T) {
	r := newPresenceRig(t)
	r.report(map[string]any{"presence": false, "illuminance": 30.0})
	r.expectNothing(t)

	r.clock.Advance(5*time.Minute - time.Second)
	r.expectNothing(t)
	r.clock.Advance(time.Second)
	r.expectCommand(t, lampOff)
}

// AC-11: presence=true before expiry cancels the pending OFF.
func TestPresenceReturnCancelsPendingOff(t *testing.T) {
	r := newPresenceRig(t)
	r.report(map[string]any{"presence": false, "illuminance": 100.0})
	r.clock.Advance(2 * time.Minute)

	r.report(map[string]any{"presence": true, "illuminance": 100.0}) // too bright: no ON
	r.clock.Advance(10 * time.Minute)
	r.expectNothing(t)

	// The cancelled delay no longer blocks a new one, which runs a full 5 minutes.
	r.report(map[string]any{"presence": false, "illuminance": 100.0})
	r.clock.Advance(5*time.Minute - time.Second)
	r.expectNothing(t)
	r.clock.Advance(time.Second)
	r.expectCommand(t, lampOff)
}

// AC-11 (current behaviour): a repeated presence=false does not restart the delay.
func TestRepeatedPresenceOffDoesNotRestartDelay(t *testing.T) {
	r := newPresenceRig(t)
	r.report(map[string]any{"presence": false, "illuminance": 30.0})
	r.clock.Advance(3 * time.Minute)
	r.report(map[string]any{"presence": false, "illuminance": 30.0})

	r.clock.Advance(2 * time.Minute)
	r.expectCommand(t, lampOff)
	r.expectNothing(t)
}

// AC-12: presence=true turns the lamp on only when illuminance <= 45.
func TestPresenceOnIlluminanceGate(t *testing.T) {
	tests := []struct {
		name        string
		illuminance float64
		want        string
	}{
		{"dark", 30, lampOn},
		{"at threshold", 45, lampOn},
		{"bright", 46, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newPresenceRig(t)
			r.report(map[string]any{"presence": true, "illuminance": tt.illuminance})
			if tt.want == "" {
				r.expectNothing(t)
				return
			}
			r.expectCommand(t, tt.want)
		})
	}
}
