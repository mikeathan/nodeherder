package controllers

import (
	"errors"
	"node-herder/internal/automations"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/store"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type failureEventHub struct {
	*mocks.MockEventHub
	calls       *[]string
	err         error
	lastEvent   string
	lastPayload interface{}
}

func (h *failureEventHub) Broadcast(event string, payload interface{}) error {
	*h.calls = append(*h.calls, "broadcast")
	h.lastEvent, h.lastPayload = event, payload
	return h.err
}

func (h *failureEventHub) EmitDevice(name string) error {
	*h.calls = append(*h.calls, "emit:"+name)
	return h.err
}

type failureStore struct {
	store.AppStore
	calls *[]string
	err   error
}

func (s *failureStore) StoreDevice(string, *devices.Device) error {
	*s.calls = append(*s.calls, "store")
	return s.err
}

func (s *failureStore) RemoveDeviceById(string) error {
	*s.calls = append(*s.calls, "remove")
	return s.err
}

func (s *failureStore) StoreMetrics(string, map[string]any) error {
	*s.calls = append(*s.calls, "metrics")
	return s.err
}

type disabledFailureEngine struct{ automations.Engine }

func (disabledFailureEngine) IsAutomationEnabled(string) bool { return false }

func TestDeviceOutputFailuresPreserveIndependentEffects(t *testing.T) {
	broadcastErr, storeErr := errors.New("broadcast unavailable"), errors.New("store unavailable")
	for _, tc := range []struct {
		name                   string
		broadcastErr, storeErr error
	}{
		{"success", nil, nil},
		{"broadcast fails", broadcastErr, nil},
		{"store fails", nil, storeErr},
		{"both fail", broadcastErr, storeErr},
	} {
		for _, added := range []bool{false, true} {
			t.Run(tc.name+map[bool]string{false: "/updated", true: "/added"}[added], func(t *testing.T) {
				var calls []string
				h := &HubController{
					eventHub: &failureEventHub{calls: &calls, err: tc.broadcastErr},
					store:    &failureStore{calls: &calls, err: tc.storeErr},
				}
				device := devices.NewDevice("fixture")
				var err error
				if added {
					err = h.handleDeviceAdded(device)
				} else {
					err = h.handleDeviceUpdated(device, devices.NewUpdatePackage(device.Id))
				}
				if !reflect.DeepEqual(calls, []string{"broadcast", "store"}) {
					t.Fatalf("effects = %v, want broadcast then store, once each", calls)
				}
				for _, want := range []error{tc.broadcastErr, tc.storeErr} {
					if want != nil && !errors.Is(err, want) {
						t.Errorf("error = %v, missing cause %v", err, want)
					}
				}
				if tc.broadcastErr == nil && tc.storeErr == nil && err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestDeviceCallbackFailuresAreReportedWithoutRetry(t *testing.T) {
	app := utils_test.CreateStore()
	entity := utils_test.CreateEntity("brightness", "numeric", 0.0)
	entity.Category = bridge.MeasurementCategory
	device := utils_test.CreateDeviceWithExposes("fixture", "fixture", []*devices.Entity{entity})
	if err := app.StoreDevice(device.FriendlyName, device); err != nil {
		t.Fatal(err)
	}
	cfg := settings.NewDeviceConfig(device.Id)
	cfg.MetricsEnabled = true
	cfg.DebounceOverrides["brightness"] = utils.IntervalFromSeconds(0)
	if err := app.AppConfig().SetDeviceConfigOverrides(cfg); err != nil {
		t.Fatal(err)
	}
	var calls []string
	storageErr, broadcastErr := errors.New("fixture storage failed"), errors.New("fixture broadcast failed")
	w := &failureEventHub{calls: &calls, err: broadcastErr}
	h := &HubController{
		store: &failureStore{AppStore: app, calls: &calls, err: storageErr}, eventHub: w,
		automationEngine: disabledFailureEngine{}, DeviceAvailabilityTimeoutOverrideInHours: 1,
	}
	h.registrar = services.NewHubRegisterService(h.store, w, 0)
	// Observe existing logger output without replacing process stdout. The hook is
	// process-wide, so keep only this device's output reports: background work left
	// by other tests may log unrelated errors while this test runs.
	var mu sync.Mutex
	var messages []string
	reportPrefix := "device " + device.Id + " output failed"
	utils.RegisterRemoteLoggerHook(mocks.NewMockRemoteLoggerEmitter(func(_ string, data interface{}) error {
		payload := data.(map[string]interface{})
		if msg, _ := payload["message"].(string); payload["level"] == "error" && strings.HasPrefix(msg, reportPrefix) {
			mu.Lock()
			messages = append(messages, payload["message"].(string))
			mu.Unlock()
		}
		return nil
	}))
	utils.EnableRemoteLoggerHook(true)
	t.Cleanup(func() { utils.EnableRemoteLoggerHook(false); utils.RemoveRemoteLoggerHook() })
	processor := h.createDeviceProcessor()
	t.Cleanup(func() {
		disabled := settings.NewDeviceConfig(device.Id)
		disabled.Disabled = true
		processor.OnDeviceConfigUpdated(disabled)
	})
	for _, value := range []float64{1, 2} {
		if err := processor.CreateOrUpdateDevice(device.FriendlyName, "mqtt", map[string]interface{}{"brightness": value}); err != nil {
			t.Fatal(err)
		}
	}
	h.handleDeviceAvailabilityChanged(devices.NewUpdatePackage(device.Id))
	if want := []string{"metrics", "broadcast", "store", "metrics", "broadcast", "store", "broadcast"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("effects = %v, want %v (no retries or suppressed effects)", calls, want)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(messages) != 5 {
		t.Fatalf("error reports = %v, want five", messages)
	}
	for _, stage := range []string{"store metrics", "broadcast deviceAdded", "broadcast deviceUpdated", "store device", "broadcast availability"} {
		if !strings.Contains(strings.Join(messages, "\n"), stage) {
			t.Errorf("missing stage %q in %v", stage, messages)
		}
	}
}
