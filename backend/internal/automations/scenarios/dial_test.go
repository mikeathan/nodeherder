// Scenario test of the Living room switch dial automation using
// backend/configs/automations/0x001788010d7d9d3f.json, driven through the hub.
package scenarios

import (
	"encoding/json"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/testing/hubharness"
	"node-herder/utils"
	"reflect"
	"testing"
	"time"
)

const (
	dialID    = "0x001788010d7d9d3f"
	dialName  = "Living room switch dial"
	lightID   = "0x001788010ba4328e"
	lightName = "Living Room Light"
)

// livingRoomBridge is a bridge/devices excerpt for the dial and the light it
// controls, shaped like Zigbee2MQTT output (exposes with features).
var livingRoomBridge = []byte(`[
 {"ieee_address":"` + dialID + `","friendly_name":"` + dialName + `","type":"EndDevice","power_source":"Battery","interview_completed":true,
  "definition":{"description":"Hue tap dial switch","exposes":[
   {"type":"enum","name":"action","property":"action","access":1,"values":["button_1_press","button_1_press_release","button_2_press","button_2_press_release","dial_rotate_left_slow","dial_rotate_left_fast","dial_rotate_right_slow","dial_rotate_right_fast"]},
   {"type":"numeric","name":"action_time","property":"action_time","access":1},
   {"type":"numeric","name":"linkquality","property":"linkquality","access":1,"value_min":0,"value_max":255}]}},
 {"ieee_address":"` + lightID + `","friendly_name":"` + lightName + `","type":"Router","power_source":"Mains (single phase)","interview_completed":true,
  "definition":{"description":"Hue white ambiance","exposes":[
   {"type":"light","features":[
    {"type":"binary","name":"state","property":"state","access":7,"value_on":"ON","value_off":"OFF","value_toggle":"TOGGLE"},
    {"type":"numeric","name":"brightness","property":"brightness","access":7,"value_min":0,"value_max":254},
    {"type":"numeric","name":"color_temp","property":"color_temp","access":7,"value_min":153,"value_max":500,
     "presets":[{"name":"coolest","value":153},{"name":"cool","value":250},{"name":"neutral","value":370},{"name":"warm","value":454},{"name":"warmest","value":500}]}]},
   {"type":"numeric","name":"linkquality","property":"linkquality","access":1,"value_min":0,"value_max":255}]}}
]`)

// dialRig drives the real dial automation through the hub. The light does not
// answer by itself; tests decide when a command is confirmed.
type dialRig struct {
	h          *hubharness.Harness
	published  chan map[string]any
	dialEvents chan any
	lightAcks  chan map[string]any
}

func newDialRig(t *testing.T, lightSeed map[string]any) *dialRig {
	t.Helper()
	h := hubharness.NewFromBridge(t, livingRoomBridge, 2, hubharness.LoadAutomationConfig(t, dialID))

	// Light confirmations must always be broadcast so the test can wait on them.
	lightCfg := settings.NewDeviceConfig(lightID)
	lightCfg.DebounceOverrides["brightness"] = utils.IntervalFromMilliseconds(0)
	lightCfg.DebounceOverrides["color_temp"] = utils.IntervalFromMilliseconds(0)
	if err := h.Store.AppConfig().SetDeviceConfigOverrides(lightCfg); err != nil {
		t.Fatal(err)
	}
	h.Seed(t, map[string]map[string]any{dialName: {"linkquality": 100.0}})
	h.Seed(t, map[string]map[string]any{lightName: lightSeed})

	r := &dialRig{
		h:          h,
		published:  make(chan map[string]any, 256),
		dialEvents: make(chan any, 256),
		lightAcks:  make(chan map[string]any, 256),
	}
	h.Paho.OnPublish(func(topic string, payload []byte) {
		if topic != "zigbee2mqtt/"+lightName+"/set" {
			return
		}
		var cmd map[string]any
		if err := json.Unmarshal(payload, &cmd); err != nil {
			t.Errorf("invalid command %s: %v", payload, err)
			return
		}
		r.published <- cmd
	})
	h.EventHub.OnBroadcast(func(eventName string, data interface{}) {
		p, ok := data.(*devices.UpdatePackage)
		if !ok || eventName != ws.DeviceUpdated {
			return
		}
		switch p.Id {
		case dialID:
			if action, has := p.Data["action"]; has {
				r.dialEvents <- action
			}
		case lightID:
			r.lightAcks <- p.Data
		}
	})
	return r
}

func dialEvent(action string, actionTime float64) map[string]any {
	return map[string]any{"action": action, "action_time": actionTime, "linkquality": 100.0}
}

// awaitDial waits for n dial broadcasts. Update broadcasts after the automation
// ran, so every command those events published has been recorded by then.
func (r *dialRig) awaitDial(t *testing.T, n int) {
	t.Helper()
	timeout := time.After(hubharness.Timeout)
	for i := 0; i < n; i++ {
		select {
		case <-r.dialEvents:
		case <-timeout:
			t.Fatalf("timed out waiting for dial event %d of %d", i+1, n)
		}
	}
}

func (r *dialRig) drainPublished() []map[string]any {
	var cmds []map[string]any
	for {
		select {
		case cmd := <-r.published:
			cmds = append(cmds, cmd)
		default:
			return cmds
		}
	}
}

// confirm plays the light: it applies the command and reports the new state,
// which is what the next step reads (Q-02: step from last confirmed state).
func (r *dialRig) confirm(t *testing.T, cmd map[string]any, lightState *string) {
	t.Helper()
	reply := map[string]any{}
	for k, v := range cmd {
		reply[k] = v
	}
	if reply["state"] == "TOGGLE" {
		*lightState = map[string]string{"ON": "OFF", "OFF": "ON"}[*lightState]
		reply["state"] = *lightState
	}
	r.h.Deliver(t, lightName, reply)
	select {
	case <-r.lightAcks:
	case <-time.After(hubharness.Timeout):
		t.Fatalf("timed out waiting for light to confirm %v", cmd)
	}
}

// dialGoldenCases are the golden commands for non-overlapping dial events using
// the real automation 0x001788010d7d9d3f, starting from a light seeded with
// dialGoldenSeed. Expected values follow the current step maths: brightness ±
// action_time × coefficient (slow 0.5, fast 1), clamped to the light's 0..254,
// skipped when unchanged.
var dialGoldenSeed = map[string]any{"state": "OFF", "brightness": 100.0, "color_temp": 370.0}

var dialGoldenCases = []struct {
	name  string
	event map[string]any
	want  []map[string]any
}{
	{"right slow", dialEvent("dial_rotate_right_slow", 20), []map[string]any{{"brightness": 110.0}}},
	{"right fast", dialEvent("dial_rotate_right_fast", 30), []map[string]any{{"brightness": 140.0}}},
	{"left slow", dialEvent("dial_rotate_left_slow", 40), []map[string]any{{"brightness": 120.0}}},
	{"left fast", dialEvent("dial_rotate_left_fast", 50), []map[string]any{{"brightness": 70.0}}},
	{"left fast clamps at min", dialEvent("dial_rotate_left_fast", 100), []map[string]any{{"brightness": 0.0}}},
	{"left slow at min is skipped", dialEvent("dial_rotate_left_slow", 10), nil},
	{"right fast clamps at max", dialEvent("dial_rotate_right_fast", 300), []map[string]any{{"brightness": 254.0}}},
	{"right slow at max is skipped", dialEvent("dial_rotate_right_slow", 20), nil},
	{"repeated event 1", dialEvent("dial_rotate_left_slow", 20), []map[string]any{{"brightness": 244.0}}},
	{"repeated event 2", dialEvent("dial_rotate_left_slow", 20), []map[string]any{{"brightness": 234.0}}},
	{"button 1 toggles", map[string]any{"action": "button_1_press_release"}, []map[string]any{{"state": "TOGGLE"}}},
	{"button 1 toggles again", map[string]any{"action": "button_1_press_release"}, []map[string]any{{"state": "TOGGLE"}}},
	{"button 1 press is ignored", map[string]any{"action": "button_1_press"}, nil},
	// presets cycle in sorted-name order: cool, coolest, neutral, warm, warmest, then wrap
	{"preset 1", map[string]any{"action": "button_2_press_release"}, []map[string]any{{"color_temp": 250.0}}},
	{"preset 2", map[string]any{"action": "button_2_press_release"}, []map[string]any{{"color_temp": 153.0}}},
	{"preset 3", map[string]any{"action": "button_2_press_release"}, []map[string]any{{"color_temp": 370.0}}},
	{"preset 4", map[string]any{"action": "button_2_press_release"}, []map[string]any{{"color_temp": 454.0}}},
	{"preset 5", map[string]any{"action": "button_2_press_release"}, []map[string]any{{"color_temp": 500.0}}},
	{"preset wraps", map[string]any{"action": "button_2_press_release"}, []map[string]any{{"color_temp": 250.0}}},
}

// runDialGolden delivers each golden event, checks exactly the commands it
// published, and confirms them as the light would.
func (r *dialRig) runDialGolden(t *testing.T) {
	t.Helper()
	lightState := "OFF"
	for _, c := range dialGoldenCases {
		r.h.Deliver(t, dialName, c.event)
		r.awaitDial(t, 1)
		got := r.drainPublished()
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("%s: published %v, want %v", c.name, got, c.want)
		}
		for _, cmd := range got {
			r.confirm(t, cmd, &lightState)
		}
	}
}

// AC-05, AC-06, AC-07: golden commands for non-overlapping dial events.
func TestDialGoldenSequence(t *testing.T) {
	newDialRig(t, dialGoldenSeed).runDialGolden(t)
}

// AC-04: in a burst of alternating-direction rotations (no confirmations in
// between, so the confirmed brightness stays 127), every evaluation must use
// its own action and action_time.
func TestDialBurstEachEventUsesItsOwnValues(t *testing.T) {
	const events = 40
	const base = 127.0
	r := newDialRig(t, map[string]any{"state": "ON", "brightness": base, "color_temp": 370.0})

	kinds := []struct {
		action string
		delta  func(actionTime float64) float64
	}{
		{"dial_rotate_right_slow", func(t float64) float64 { return t * 0.5 }},
		{"dial_rotate_left_fast", func(t float64) float64 { return -t }},
		{"dial_rotate_right_fast", func(t float64) float64 { return t }},
		{"dial_rotate_left_slow", func(t float64) float64 { return -t * 0.5 }},
	}

	want := make([]map[string]any, 0, events)
	for i := 0; i < events; i++ {
		kind := kinds[i%len(kinds)]
		actionTime := float64(5 + (i*7)%60)
		r.h.Deliver(t, dialName, dialEvent(kind.action, actionTime))
		want = append(want, map[string]any{"brightness": base + kind.delta(actionTime)})
	}
	r.awaitDial(t, events)

	got := r.drainPublished()
	if !reflect.DeepEqual(got, want) {
		mismatches := 0
		for i := range want {
			if i >= len(got) || !reflect.DeepEqual(got[i], want[i]) {
				mismatches++
			}
		}
		t.Fatalf("burst: %d published for %d events, %d mismatched positions\n got: %v\nwant: %v", len(got), events, mismatches, got, want)
	}
}
