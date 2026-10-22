// Package hubharness is the single test harness that runs a real HubController
// over an in-memory paho client, so messages travel the production path:
// paho callback → hub ingestion → DeviceProcessor → Update → broadcast/store.
// It is test support only, built on exported APIs, and is shared by the hub
// controller tests and the automation scenario tests.
package hubharness

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/controllers"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/store"
	utils_test "node-herder/testing"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

// Timeout bounds every wait the harness performs.
const Timeout = 5 * time.Second

// EventHub records broadcasts and exposes UI command handlers so a test can
// drive the same paths as the UI, optionally through a real transport.
type EventHub struct {
	ws.EventHub
	mu               sync.Mutex
	setValue         func(p interface{}) error
	saveAutomation   func(p interface{}) error
	deleteAutomation func(p interface{}) (interface{}, error)
	broadcast        func(eventName string, data interface{})
	broadcastErr     error
}

func (h *EventHub) OnSaveAutomation(action func(p interface{}) error) {
	h.mu.Lock()
	h.saveAutomation = action
	h.mu.Unlock()
	h.EventHub.OnSaveAutomation(action)
}

// SaveAutomation invokes the registered UI "save automation" handler.
func (h *EventHub) SaveAutomation(p interface{}) error {
	h.mu.Lock()
	action := h.saveAutomation
	h.mu.Unlock()
	return action(p)
}

func (h *EventHub) OnDeleteAutomation(action func(p interface{}) (interface{}, error)) {
	h.mu.Lock()
	h.deleteAutomation = action
	h.mu.Unlock()
	h.EventHub.OnDeleteAutomation(action)
}

// DeleteAutomation invokes the registered UI "delete automation" handler.
func (h *EventHub) DeleteAutomation(p interface{}) error {
	h.mu.Lock()
	action := h.deleteAutomation
	h.mu.Unlock()
	_, err := action(p)
	return err
}

// FailBroadcasts makes every later broadcast return err (nil restores success).
// The observer still sees each broadcast: a failed send was still attempted.
func (h *EventHub) FailBroadcasts(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.broadcastErr = err
}

func (h *EventHub) OnDeviceSetValue(action func(p interface{}) error) {
	h.mu.Lock()
	h.setValue = action
	h.mu.Unlock()
	h.EventHub.OnDeviceSetValue(action)
}

// SetValue invokes the registered UI "set value" handler.
func (h *EventHub) SetValue(p interface{}) error {
	h.mu.Lock()
	action := h.setValue
	h.mu.Unlock()
	return action(p)
}

// OnBroadcast sets (or, with nil, clears) the observer of every broadcast.
func (h *EventHub) OnBroadcast(fn func(eventName string, data interface{})) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.broadcast = fn
}

func (h *EventHub) Broadcast(eventName string, data interface{}) error {
	h.mu.Lock()
	fn, err := h.broadcast, h.broadcastErr
	h.mu.Unlock()
	if fn != nil {
		fn(eventName, data)
	}
	if err != nil {
		return err
	}
	return h.EventHub.Broadcast(eventName, data)
}

// Store is the hub's real test store with injectable device/metrics write failures.
type Store struct {
	store.AppStore
	mu       sync.Mutex
	writeErr error
	attempts int
}

// FailWrites makes every later StoreDevice and StoreMetrics call return err
// without writing (nil restores normal writes).
func (s *Store) FailWrites(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writeErr = err
}

// FailedWrites reports how many writes were attempted while failing.
func (s *Store) FailedWrites() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.attempts
}

func (s *Store) failWrite() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writeErr != nil {
		s.attempts++
	}
	return s.writeErr
}

func (s *Store) StoreDevice(friendlyName string, device *devices.Device) error {
	if err := s.failWrite(); err != nil {
		return err
	}
	return s.AppStore.StoreDevice(friendlyName, device)
}

func (s *Store) StoreMetrics(friendlyName string, payload map[string]any) error {
	if err := s.failWrite(); err != nil {
		return err
	}
	return s.AppStore.StoreMetrics(friendlyName, payload)
}

// Harness wires the real MqttService (over an in-memory paho client) into a real
// HubController.
type Harness struct {
	Hub      *controllers.HubController
	EventHub *EventHub
	Paho     *mocks.FakePahoClient
	Store    *Store
}

// New starts a hub that has been told about devs by a bridge/devices message.
func New(tb testing.TB, devs []*devices.Device, automationList ...automations.Automation) *Harness {
	tb.Helper()
	bridgeList, err := json.Marshal(utils_test.CreateBridgeInfoList(devs))
	if err != nil {
		tb.Fatal(err)
	}
	return NewFromBridge(tb, bridgeList, len(devs), automationList...)
}

// NewFromBridge starts a hub from a raw bridge/devices payload announcing
// deviceCount devices.
func NewFromBridge(tb testing.TB, bridgeList []byte, deviceCount int, automationList ...automations.Automation) *Harness {
	tb.Helper()
	transport := mocks.NewMockEventHub()
	transport.SetMockBroadcastEvent(func(string, interface{}) error { return nil })
	return newFromBridge(tb, bridgeList, deviceCount, transport, automationList)
}

// NewFromBridgeWithEventHub uses a caller-owned transport, for browser integration
// tests that need real WebSocket delivery as well as in-memory MQTT.
func NewFromBridgeWithEventHub(tb testing.TB, bridgeList []byte, deviceCount int, eventHub ws.EventHub, options ...controllers.HubControllerOption) *Harness {
	tb.Helper()
	return newFromBridge(tb, bridgeList, deviceCount, eventHub, nil, options...)
}

func newFromBridge(tb testing.TB, bridgeList []byte, deviceCount int, transport ws.EventHub, automationList []automations.Automation, options ...controllers.HubControllerOption) *Harness {
	tb.Helper()

	eventHub := &EventHub{EventHub: transport}
	appStore := &Store{AppStore: utils_test.CreateStore()}

	var paho *mocks.FakePahoClient
	client := mqtt.NewMqttClient(mqtt.WithClientFactory(func(options *mqttlib.ClientOptions) mqttlib.Client {
		paho = mocks.NewFakePahoClient(options)
		return paho
	}))

	hub := controllers.RegisterHubController(eventHub, appStore, client, options...)
	if hub == nil {
		tb.Fatal("hub controller failed to start")
	}
	hub.WithAutomationStorage(mocks.NewMockAutomationStorage[automations.Automation](automationList))
	tb.Cleanup(func() { paho.Disconnect(250) })

	h := &Harness{Hub: hub, EventHub: eventHub, Paho: paho, Store: appStore}

	// OnConnect subscribes the bridge topics on its own goroutine; let it finish
	// before bridge/devices adds device topics.
	h.AwaitSubscriptions(tb, 6)

	paho.Deliver("zigbee2mqtt/bridge/devices", bridgeList)
	h.AwaitSubscriptions(tb, deviceCount)
	return h
}

// AwaitSubscriptions waits for n topic subscriptions to be made.
func (h *Harness) AwaitSubscriptions(tb testing.TB, n int) {
	tb.Helper()
	timeout := time.After(Timeout)
	for i := 0; i < n; i++ {
		select {
		case <-h.Paho.Subscriptions():
		case <-timeout:
			tb.Fatalf("timed out waiting for %d subscriptions (got %d)", n, i)
		}
	}
}

// Seed delivers an initial payload per device, one device at a time, and waits
// until each device has been created (deviceAdded broadcast) before the next.
func (h *Harness) Seed(tb testing.TB, payloads map[string]map[string]any) {
	tb.Helper()
	added := make(chan string, len(payloads))
	h.EventHub.OnBroadcast(func(eventName string, data interface{}) {
		if eventName == ws.DeviceAdded {
			added <- data.(*devices.Device).FriendlyName
		}
	})
	for name, payload := range payloads {
		h.Deliver(tb, name, payload)
		select {
		case <-added:
		case <-time.After(Timeout):
			tb.Fatalf("timed out waiting for %s to be created", name)
		}
	}
	h.EventHub.OnBroadcast(nil)
}

// Deliver publishes payload on the device's Zigbee2MQTT topic, as the broker would.
func (h *Harness) Deliver(tb testing.TB, friendlyName string, payload map[string]any) {
	tb.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		tb.Fatal(err)
	}
	h.Paho.Deliver("zigbee2mqtt/"+friendlyName, data)
}

// LoadAutomationConfig loads backend/configs/automations/<id>.json, whatever the
// calling test's working directory.
func LoadAutomationConfig(tb testing.TB, id string) automations.Automation {
	tb.Helper()
	serializer := automations.NewAutomationSerialiser()
	automation, err := serializer.Unmarshal(readAutomationConfig(tb, id))
	if err != nil {
		tb.Fatal(err)
	}
	return automation
}

// AutomationConfigPayload returns the same configuration as the decoded JSON
// object a UI "save automation" message carries.
func AutomationConfigPayload(tb testing.TB, id string) map[string]any {
	tb.Helper()
	var payload map[string]any
	if err := json.Unmarshal(readAutomationConfig(tb, id), &payload); err != nil {
		tb.Fatal(err)
	}
	return payload
}

func readAutomationConfig(tb testing.TB, id string) []byte {
	tb.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("cannot locate the hubharness source directory")
	}
	path := filepath.Join(filepath.Dir(thisFile), "..", "..", "configs", "automations", fmt.Sprintf("%s.json", id))
	data, err := os.ReadFile(path)
	if err != nil {
		tb.Fatal(err)
	}
	return data
}
