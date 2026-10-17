package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"node-herder/internal/automations"
	"node-herder/internal/lanes"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/store"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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

type recordingRemovalStore struct {
	store.AppStore
	removed atomic.Bool
}

func (s *recordingRemovalStore) RemoveDeviceById(id string) error {
	s.removed.Store(true)
	return s.AppStore.RemoveDeviceById(id)
}

type recordingRemovalLanes struct {
	*lanes.Executor
	entered chan string
}

func (l *recordingRemovalLanes) Do(ctx context.Context, key string, task func() error) error {
	l.entered <- key
	return l.Executor.Do(ctx, key, task)
}

func TestBridgeRemovalWaitsForDeviceLaneAndLeavesOtherLanesFree(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	executor := &recordingRemovalLanes{Executor: lanes.New(), entered: make(chan string, 1)}
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer func() {
		releaseOnce.Do(func() { close(release) })
		cancel()
		shutdown, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		if err := executor.Shutdown(shutdown); err != nil {
			t.Error(err)
		}
	}()
	app := &recordingRemovalStore{AppStore: utils_test.CreateStoreFromDeviceRepo(repository.NewMemoryDeviceRepo())}
	registrar := services.NewHubRegisterService(app, &mocks.MockEventHub{}, 0)
	seeded := make(chan *devices.Device, 1)
	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: time.Hour,
		OnNewDevice: func(device *devices.Device) {
			seeded <- device
			<-release // The first snapshot has not reached storage yet.
			if err := app.StoreDevice(device.FriendlyName, device); err != nil {
				t.Error(err)
			}
		},
		OnDeviceAvailabilityChanged: func(*devices.UpdatePackage) {},
	}
	processor := services.NewDeviceProcessorBuilder().WithContext(ctx).WithRegistrar(registrar).WithStore(app).WithEvents(events).WithAutomationQuerier(mocks.NewMockAutomationDeviceQuerier()).Build()
	var calls []string
	controller := &HubController{ctx: ctx, registrar: registrar, deviceLanes: executor,
		responseHandlers: map[string]handler{}, eventHub: &failureEventHub{calls: &calls},
		getDeviceProcessor: func() *services.DeviceProcessor { return processor }}
	seedResult := make(chan error, 1)
	if err := executor.Submit("retire-lane", func() {
		seedResult <- processor.CreateOrUpdateDevice("retire-lane", "mqtt", map[string]interface{}{"temperature": float64(1)})
	}); err != nil {
		t.Fatal(err)
	}
	var device *devices.Device
	select {
	case device = <-seeded:
	case <-time.After(5 * time.Second):
		t.Fatal("seed did not start")
	}
	payload, err := json.Marshal(map[string]interface{}{"status": "ok", "data": map[string]interface{}{"id": device.Id}})
	if err != nil {
		t.Fatal(err)
	}
	handler := controller.responseHandler("bridge/response/device/remove")
	removed := make(chan error, 1)
	go func() { removed <- handler.ProcessPayload("bridge/response/device/remove", "mqtt", payload) }()
	select {
	case key := <-executor.entered:
		if key != device.FriendlyName {
			t.Fatalf("wrong retirement lane %q", key)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("removal bypassed device lane")
	}
	if app.removed.Load() {
		t.Fatal("storage deletion bypassed in-flight Seed")
	}
	unrelated := make(chan struct{})
	if err := executor.Submit("other-device", func() { close(unrelated) }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-unrelated:
	case <-time.After(5 * time.Second):
		t.Fatal("removal blocked unrelated lanes")
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case err := <-seedResult:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Seed did not finish")
	}
	select {
	case err := <-removed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("removal did not finish")
	}
	if _, err := app.FindDeviceById(device.Id); err == nil {
		t.Fatal("removed device was reinserted by earlier work")
	}
	if !reflect.DeepEqual(calls, []string{"broadcast"}) {
		t.Fatalf("removal response changed: %v", calls)
	}
}
