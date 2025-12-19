package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"node-herder/internal/api"
	"node-herder/internal/controllers"
	"node-herder/internal/fs"
	metricsdomain "node-herder/internal/metrics/domain"
	metricsquery "node-herder/internal/metrics/query"
	"node-herder/internal/ratelimiter"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/models/logging"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/store"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"os"
	"strings"
	"testing"
	"time"
)

const device1BatterySource = `{"name":"device 1","conn":"mqtt","power_source":"battery","humidity":92.49999999999999,"temperature":19.000000000000004,"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":47,"battery":98}`
const device1RootPayloadBatterySource = `{"name":"device 1", "timestamp":"2023-08-01T16:30:04Z", "readings":{"humidity":92.49999999999999,"temperature":19.000000000000004}}`
const device1InvalidRootPayloadBatterySource = `{"name":"device 1", "timestamp":"2023-08-01T16:30:04Z", "readingstest":{"humidity":92.49999999999999,"temperature":19.000000000000004}}`
const missingDeviceIdPayload = `{"conn":"mqtt","power_source":"battery","humidity":92.49999999999999,"temperature":19.000000000000004,"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":47,"battery":98}`
const gasNodePayload = `{"label": "gas_monitor", "node_id": "3", "temperature": "22.2 *C", "humidity": "33 %RH", "air_quality_score": "95 %", "PM1.0": "1 ug/m3 (ultrafine particles)", "PM2.5": "1 ug/m3 (combustion particles, organic compounds, metal)", "PM10.0": "2 ug/m3 (dust, pollen, mould spores)", "timestamp": 1691517687.940329}`

func TestHandleMissingDeviceIdPayload(t *testing.T) {
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	store := utils_test.CreateStore()

	hub := controllers.RegisterHubController(ws, store, mqtt)
	bodyReader := strings.NewReader(string(missingDeviceIdPayload))
	req := httptest.NewRequest(http.MethodPost, "/api/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h := api.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	body, _ := io.ReadAll(w.Body)
	expectedError := "invalid data: data structure is missing node id\n"
	if string(body) != expectedError {
		t.Errorf("handler returned wrong status code: got %v want %v", string(body), expectedError)
	}
}

func TestHandleInvalidDataPayload(t *testing.T) {

	store := utils_test.CreateStore()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, store, mqtt)
	bodyReader := strings.NewReader(string("test"))
	req := httptest.NewRequest(http.MethodPost, "/api/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h := api.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestHandleSuccesfullyRootPayload(t *testing.T) {

	name := "device 1"
	ws := &mocks.NopWsServer{}
	store := utils_test.CreateStore()

	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, store, mqtt)
	bodyReader := strings.NewReader(string(device1RootPayloadBatterySource))
	req := httptest.NewRequest(http.MethodPost, "/api/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h := api.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	time.Sleep(200 * time.Millisecond)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	id := utils.HashName(name)
	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	if device == nil {
		t.Fatalf("want %s got %s", name, "nil")
	}
	id = utils.HashName(name)
	if device.Id != id {
		t.Fatalf("want %s got %s", id, device.Id)
	}
}

func TestHandleInvalidRootPayload(t *testing.T) {

	store := utils_test.CreateStore()
	ws := &mocks.NopWsServer{}

	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, store, mqtt)
	bodyReader := strings.NewReader(string(device1InvalidRootPayloadBatterySource))
	req := httptest.NewRequest(http.MethodPost, "/api/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h := api.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	body, _ := io.ReadAll(w.Body)
	expectedBody := "invalid data: data structure not containing valid payload section\n"
	if string(body) != expectedBody {
		t.Errorf("error reading body got %v want %v", string(body), expectedBody)
	}
}

func TestProcessorHandleRootPayloadWithTimestamp(t *testing.T) {

	timestamp := "2023-07-20T19:48:35+01:00"

	name := "device1"
	device1RootPayload := `{"nickname":"device1","timestamp":"2023-07-20T19:48:35+01:00", "readings":{"humidity":92.1,"temperature":19.3}}`

	store := utils_test.CreateStore()

	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, store, mqtt)
	bodyReader := strings.NewReader(string(device1RootPayload))
	req := httptest.NewRequest(http.MethodPost, "/api/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h := api.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	ts, _ := time.Parse(time.RFC3339, timestamp)
	want := ts.Format(time.RFC3339)
	time.Sleep(100 * time.Millisecond)

	id := utils.HashName(name)
	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatal(err.Error())
	}

	if device.Id != id {
		t.Fatalf("want %s got %s", id, device.Id)
	}

	if device.LastSeen == "" {
		t.Fatalf("want %s got %s", "last_seen", "nil")
	}

	if device.LastSeen != want {
		t.Fatalf("want %s got %s", want, device.LastSeen)
	}
}

func TestHandleSuccesfullyPayload(t *testing.T) {
	name := "gas_monitor"
	ws := &mocks.NopWsServer{}
	store := utils_test.CreateStore()

	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, store, mqtt)
	bodyReader := strings.NewReader(string(gasNodePayload))
	req := httptest.NewRequest(http.MethodPost, "/api/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h := api.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	time.Sleep(200 * time.Millisecond)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	id := utils.HashName(name)
	device, err := store.FindDeviceById(id)

	if err != nil {
		t.Fatal(err.Error())
	}
	if device == nil {
		t.Fatalf("want %s got %s", name, "nil")
	}

	if device.Id != id {
		t.Fatalf("want %s got %s", id, device.Id)
	}
}

func TestHandleUnsuportedMediaType(t *testing.T) {

	store := utils_test.CreateStore()
	ws := &mocks.NopWsServer{}

	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, store, mqtt)

	bodyReader := strings.NewReader(string(device1BatterySource))
	req := httptest.NewRequest(http.MethodPost, "/api/collect", bodyReader)
	w := httptest.NewRecorder()

	h := api.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusUnsupportedMediaType {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnsupportedMediaType)
	}
	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Errorf("error reading body got %v want nil", err)
	}

	expectedBody := "Content-Type header is not application/json\n"
	if string(body) != expectedBody {
		t.Errorf("error reading body got %v want %v", string(body), expectedBody)
	}
}

func TestLoadLogFileHandler(t *testing.T) {

	mockFileBuffer := []byte("test file")
	mockFile := "nodeherder.log"

	fs := fs.NewFileSystem(
		fs.WithFileLoader(mocks.NewMockFileLoader(mockFileBuffer)),
		fs.WithFileWalker(mocks.NewMockWalker([]string{mockFile})))

	reqJson, err := json.Marshal(logging.NewFileLogRequest(mockFile, logging.LoadAction))
	if err != nil {
		t.Errorf("error reading body got %v want nil", err)
	}
	bodyReader := strings.NewReader(string(reqJson))
	req := httptest.NewRequest(http.MethodPost, "/api/logfile", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h := api.NewLogFileHandler(fs)

	h.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body, _ := io.ReadAll(w.Body)
	if string(body) != string(mockFileBuffer) {
		t.Errorf("error reading body got %v want %v", string(body), string(mockFileBuffer))
	}

}
func TestHandleListLogFiles(t *testing.T) {

	mockeFiles := []string{"nodeherder.log", "nodeherder2.log", "nodeherder3.log", "nodeherder4.log"}

	mockFileBuffer := []byte("test file")

	fs := fs.NewFileSystem(
		fs.WithFileLoader(mocks.NewMockFileLoader(mockFileBuffer)),
		fs.WithFileWalker(mocks.NewMockWalker(mockeFiles)))

	req := httptest.NewRequest(http.MethodGet, "/api/listlogs", nil)
	w := httptest.NewRecorder()

	h := api.NewListFileLogsHandler(fs)
	h.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Errorf("error reading body got %v want nil", err)
	}

	resultFiles := []string{}
	err = json.Unmarshal(body, &resultFiles)
	if err != nil {
		t.Errorf("error reading body got %v want nil", err)
	}

	for i := range mockeFiles {
		if resultFiles[i] != mockeFiles[i] {
			t.Errorf("error reading body got %v want %v", resultFiles[i], mockeFiles[i])
		}
	}
}

func TestHubStateHandler_ReturnsHubState(t *testing.T) {

	// load devices from file
	store, err := utils_test.CreateStoreWithDevices()
	if err != nil {
		t.Fatalf("error creating store: %v", err)
	}

	handler := api.NewHubStateHandler(store, 5*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp *hub.HubState
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if resp == nil || resp.Devices == nil || resp.Config == nil {
		t.Fatalf("expected hub state, got nil")
	}

	// assert app config
	// 	if len(hubState.Config.Hub.Devices.Overrides) != len(inputAppConfig.Hub.Devices.Overrides) {
	// 		t.Fatalf("Expected numer of appconfig devices. want %v', got '%v'", len(inputAppConfig.Hub.Devices.Overrides), len(hubState.Config.Hub.Devices.Overrides))
	// 	}
	// 	for id, d := range inputAppConfig.Hub.Devices.Overrides {
	// 		gotDeviceConfig := hubState.Config.Hub.Devices.Overrides[id]
	// 		if d.Id != gotDeviceConfig.Id {
	// 			t.Fatalf("Expected device id %v', got '%v'", d.Id, gotDeviceConfig.Id)
	// 		}
	// 		if d.Disabled != gotDeviceConfig.Disabled {
	// 			t.Fatalf("Expected Disabled %v', got '%v'", d.Disabled, gotDeviceConfig.Disabled)
	// 		}
	// 		if d.MetricsEnabled != gotDeviceConfig.MetricsEnabled {
	// 			t.Fatalf("Expected MetricsEnabled %v', got '%v'", d.MetricsEnabled, gotDeviceConfig.MetricsEnabled)
	// 		}

	//		if d.RateLimit.Value != gotDeviceConfig.RateLimit.Value {
	//			t.Fatalf("Expected RateLimit.Value %v', got '%v'", d.RateLimit.Value, gotDeviceConfig.RateLimit.Value)
	//		}
	//		if d.RateLimit.Unit != gotDeviceConfig.RateLimit.Unit {
	//			t.Fatalf("Expected RateLimit.Unit %v', got '%v'", d.RateLimit.Unit, gotDeviceConfig.RateLimit.Unit)
	//		}
	//	}
}

func TestHubStateHandler_ReturnsCacheedState(t *testing.T) {

	// load real store to get readl hubstate data
	store, err := utils_test.CreateStoreWithDevices()
	if err != nil {
		t.Fatalf("error creating store: %v", err)
	}

	hubState, err := store.LoadHubState()
	if err != nil {
		t.Fatalf("error loading hub state: %v", err)

	}
	// now configure mock store to assert that the hub state is loaded only once
	callCount := 0
	loadHubStateFunc := (func() (*hub.HubState, error) {
		callCount++
		return hubState, nil
	})

	mockStore := mocks.NewMockAppStoreWithLoadStateFunc(loadHubStateFunc)

	handler := api.NewHubStateHandler(mockStore, 5*time.Minute)

	// First call: loads state
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr1.Code)
	}

	if callCount != 1 {
		t.Fatalf("expected call count 1, got %d", callCount)
	}
	// Second call: should reuse cache
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr2.Code)
	}
	if callCount != 1 {
		t.Fatalf("expected call count 1, got %d", callCount)
	}
}

func TestHubStateHandler_DirtyFlagTriggersReload(t *testing.T) {
	store, err := utils_test.CreateStoreWithDevices()
	if err != nil {
		t.Fatalf("error creating store: %v", err)
	}

	hubState, err := store.LoadHubState()
	if err != nil {
		t.Fatalf("error loading hub state: %v", err)
	}
	// now configure mock store to assert that the hub state is loaded only once
	callCount := 0
	loadHubStateFunc := (func() (*hub.HubState, error) {
		callCount++
		return hubState, nil
	})

	mockStore := mocks.NewMockAppStoreWithLoadStateFunc(loadHubStateFunc)

	handler := api.NewHubStateHandler(mockStore, 5*time.Minute)

	// Initial load
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if callCount != 1 {
		t.Fatalf("expected call count 1, got %d", callCount)
	}
	// Trigger dirty
	mockStore.TriggerDirty()

	// Should reload
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if callCount != 2 {
		t.Fatalf("expected call count 2, got %d", callCount)
	}
}

func TestHubStateHandler_ExpirationTriggersReload(t *testing.T) {
	store, err := utils_test.CreateStoreWithDevices()
	if err != nil {
		t.Fatalf("error creating store: %v", err)
	}

	hubState, err := store.LoadHubState()
	if err != nil {
		t.Fatalf("error loading hub state: %v", err)
	}

	callCount := 0
	loadHubStateFunc := (func() (*hub.HubState, error) {
		callCount++
		return hubState, nil
	})

	mockStore := mocks.NewMockAppStoreWithLoadStateFunc(loadHubStateFunc)

	handler := api.NewHubStateHandler(mockStore, 1*time.Millisecond)

	// Initial load
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if callCount != 1 {
		t.Fatalf("expected call count 1, got %d", callCount)
	}
	time.Sleep(2 * time.Millisecond)

	// Should reload due to TTL expiry
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if callCount != 2 {
		t.Fatalf("expected call count 2, got %d", callCount)
	}
}

func TestAutomationTriggerHandler_MissingParams(t *testing.T) {
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	store := utils_test.CreateStore()

	hub := controllers.RegisterHubController(ws, store, mqtt)

	handler := api.NewAutomationTriggerHandler((*controllers.HubController)(hub), 1*time.Second)

	req := httptest.NewRequest("GET", "/automation/trigger", nil) // no params
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", w.Code)
	}
}

func TestAutomationTriggerHandler_Cases(t *testing.T) {
	cases := []struct {
		name          string
		automationId  string
		triggerName   string
		rateLimit     time.Duration
		shouldSucceed bool
	}{
		{
			name:          "success case",
			automationId:  "123",
			triggerName:   "test",
			rateLimit:     1 * time.Second,
			shouldSucceed: true,
		},
		{
			name:          "missing automationId",
			automationId:  "",
			triggerName:   "test",
			rateLimit:     1 * time.Second,
			shouldSucceed: false,
		},
		{
			name:          "missing triggerName",
			automationId:  "123",
			triggerName:   "",
			rateLimit:     1 * time.Second,
			shouldSucceed: false,
		},
		{
			name:          "rate limit exceeded",
			automationId:  "123",
			triggerName:   "test",
			rateLimit:     1 * time.Hour,
			shouldSucceed: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			callCount := 0
			mock := mocks.NewMockAutomationTrigger(func(automationId, triggerName string) error {
				callCount++
				return nil
			})

			handler := api.NewAutomationTriggerHandler(mock, c.rateLimit)

			if c.name == "rate limit exceeded" {
				body := map[string]string{"automationId": "123", "triggerName": "test"}
				b, _ := json.Marshal(body)

				// First call should succeed
				req1 := httptest.NewRequest("POST", "/automation/trigger", bytes.NewReader(b))
				req1.Header.Set("Content-Type", "application/json")
				w1 := httptest.NewRecorder()
				handler.ServeHTTP(w1, req1)
				if w1.Code != http.StatusOK {
					t.Errorf("expected first call to succeed, got %d", w1.Code)
				}

				// Second call immediately should be rate-limited
				req2 := httptest.NewRequest("POST", "/automation/trigger", bytes.NewReader(b))
				req2.Header.Set("Content-Type", "application/json")
				w2 := httptest.NewRecorder()
				handler.ServeHTTP(w2, req2)
				if w2.Code != http.StatusTooManyRequests {
					t.Errorf("expected 429 TooManyRequests, got %d", w2.Code)
				}
				return
			}

			payload := map[string]string{"automationId": c.automationId, "triggerName": c.triggerName}
			b, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/automation/trigger", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if c.shouldSucceed && w.Code != http.StatusOK {
				t.Errorf("expected 200 OK, got %d", w.Code)
			}
			if !c.shouldSucceed && w.Code == http.StatusOK {
				t.Errorf("expected failure, got 200 OK")
			}
		})
	}
}

func TestMetricsQueryHandler_InvalidContentType(t *testing.T) {
	store := utils_test.CreateStore()
	handler := api.NewMetricsQueryHandler(ratelimiter.NewRateLimiter(), 1*time.Second, store)

	req := httptest.NewRequest(http.MethodPost, "/api/metrics/query", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", w.Code)
	}
}

func TestMetricsQueryHandler_InvalidJSON(t *testing.T) {
	store := utils_test.CreateStore()
	handler := api.NewMetricsQueryHandler(ratelimiter.NewRateLimiter(), 1*time.Second, store)

	req := httptest.NewRequest(http.MethodPost, "/api/metrics/query", strings.NewReader(`bad`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMetricsQueryHandler_SuccessWithLimitSort(t *testing.T) {
	store, metricsRepo, mockClock, cleanup := createMetricsQueryTestStore(t)
	defer cleanup()

	deviceName := "device1"
	deviceID := utils.HashName(deviceName)

	entity := devices.NewEntity("temperature")
	entity.Type = "numeric"
	entity.Data.SetValue(float32(0))

	dev := &devices.Device{
		Id:           deviceID,
		FriendlyName: deviceName,
		Exposes:      map[string]*devices.Entity{"temperature": entity},
	}

	if err := store.StoreDevice(deviceName, dev); err != nil {
		t.Fatalf("failed to store device: %v", err)
	}

	base := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	timestamps := []time.Time{
		base.Add(1 * time.Minute),
		base.Add(2 * time.Minute),
		base.Add(3 * time.Minute),
	}
	values := []float32{1, 2, 3}

	for i, ts := range timestamps {
		mockClock.SetMockTime(ts)
		if err := metricsRepo.Store(deviceID, map[string]any{"temperature": values[i]}); err != nil {
			t.Fatalf("failed to store metrics: %v", err)
		}
	}

	payload := metricsquery.MetricsQueryRequest{
		DeviceIds: []string{deviceID},
		Expose:    "temperature",
		Time: metricsdomain.TimeQuery{
			From: base,
			To:   base.Add(10 * time.Minute),
		},
		Aggregation: metricsdomain.AggNone,
		Limit:       2,
		SortDesc:    true,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/metrics/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := api.NewMetricsQueryHandler(ratelimiter.NewRateLimiter(), 1*time.Second, store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var response []metricsquery.MetricsQueryResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 || len(response[0].Values) != 1 {
		t.Fatalf("expected 1 response and 1 value, got %+v", response)
	}

	data, ok := response[0].Values[0].Value.([]any)
	if !ok {
		t.Fatalf("unexpected value type: %T", response[0].Values[0].Value)
	}

	if len(data) != 2 {
		t.Fatalf("expected 2 data points, got %d", len(data))
	}

	first := data[0].(map[string]any)
	second := data[1].(map[string]any)
	firstX := int64(first["x"].(float64))
	secondX := int64(second["x"].(float64))

	if firstX != timestamps[2].UnixMilli() || secondX != timestamps[1].UnixMilli() {
		t.Fatalf("unexpected order after sort/limit: %v, %v", firstX, secondX)
	}
}

func TestMetricsQueryHandler_RateLimit(t *testing.T) {
	store, metricsRepo, mockClock, cleanup := createMetricsQueryTestStore(t)
	defer cleanup()

	deviceName := "device1"
	deviceID := utils.HashName(deviceName)

	entity := devices.NewEntity("temperature")
	entity.Type = "numeric"
	entity.Data.SetValue(float32(0))

	dev := &devices.Device{
		Id:           deviceID,
		FriendlyName: deviceName,
		Exposes:      map[string]*devices.Entity{"temperature": entity},
	}

	if err := store.StoreDevice(deviceName, dev); err != nil {
		t.Fatalf("failed to store device: %v", err)
	}

	base := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	mockClock.SetMockTime(base)
	if err := metricsRepo.Store(deviceID, map[string]any{"temperature": float32(1)}); err != nil {
		t.Fatalf("failed to store metrics: %v", err)
	}

	payload := metricsquery.MetricsQueryRequest{
		DeviceIds: []string{deviceID},
		Expose:    "temperature",
		Time: metricsdomain.TimeQuery{
			From: base,
			To:   base.Add(10 * time.Minute),
		},
		Aggregation: metricsdomain.AggNone,
		Limit:       1,
		SortDesc:    true,
	}

	body, _ := json.Marshal(payload)
	handler := api.NewMetricsQueryHandler(ratelimiter.NewRateLimiter(), 1*time.Hour, store)

	req1 := httptest.NewRequest(http.MethodPost, "/api/metrics/query", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first call 200, got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/metrics/query", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second call 429, got %d", w2.Code)
	}
}

func createMetricsQueryTestStore(t *testing.T) (store.AppStore, metricsdomain.Repository, *mocks.MockClock, func()) {
	t.Helper()

	tempfile := utils_test.Tempfile()
	metricsRepo, mockClock, err := utils_test.CreateMetricsRepo(tempfile)
	if err != nil {
		os.Remove(tempfile)
		t.Fatalf("failed to create metrics repo: %v", err)
	}

	deviceRepo := repository.NewMemoryDeviceRepo()
	settingsRepo := &mocks.NopSettingsrepo{}
	configCache, err := settings.NewAppConfigCache(settingsRepo, []settings.Task{})
	if err != nil {
		os.Remove(tempfile)
		t.Fatalf("failed to create config cache: %v", err)
	}

	appStore, err := store.NewAppStore(deviceRepo, metricsRepo, configCache)
	if err != nil {
		os.Remove(tempfile)
		t.Fatalf("failed to create store: %v", err)
	}

	cleanup := func() {
		_ = metricsRepo.Close()
		_ = deviceRepo.Close()
		_ = os.Remove(tempfile)
	}

	return appStore, metricsRepo, mockClock, cleanup
}
