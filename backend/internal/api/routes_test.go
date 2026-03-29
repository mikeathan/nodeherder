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
	"node-herder/models/assistant"
	"node-herder/models/bridge"
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

	tests := []struct {
		name           string
		requestedFile  string
		expectedStatus int
	}{
		{
			name:           "valid log file",
			requestedFile:  "nodeherder.log",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid nested log file",
			requestedFile:  "logs/nodeherder.log",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid extension",
			requestedFile:  "config.json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "directory traversal",
			requestedFile:  "../../etc/passwd.log",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "absolute path",
			requestedFile:  "/var/log/syslog.log",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockFileBuffer := []byte("test file")

			mockFs := fs.NewFileSystem(
				fs.WithFileLoader(mocks.NewMockFileLoader(mockFileBuffer)),
				fs.WithFileWalker(mocks.NewMockWalker([]string{tc.requestedFile})))

			reqJson, err := json.Marshal(logging.NewFileLogRequest(tc.requestedFile, logging.LoadAction))
			if err != nil {
				t.Errorf("error marshalling body: %v", err)
			}
			bodyReader := strings.NewReader(string(reqJson))
			req := httptest.NewRequest(http.MethodPost, "/api/logfile", bodyReader)
			req.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h := api.NewLogFileHandler(mockFs)
			h.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
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
	handler := api.NewMetricsQueryHandler(ratelimiter.NewWindowRateLimiter(1, time.Second), store)

	req := httptest.NewRequest(http.MethodPost, "/api/metrics/query", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", w.Code)
	}
}

func TestMetricsQueryHandler_InvalidJSON(t *testing.T) {
	store := utils_test.CreateStore()
	handler := api.NewMetricsQueryHandler(ratelimiter.NewWindowRateLimiter(1, time.Second), store)

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
		Exposes:   []string{"temperature"},
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

	handler := api.NewMetricsQueryHandler(ratelimiter.NewWindowRateLimiter(5, time.Second), store)
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
		Exposes:   []string{"temperature"},
		Time: metricsdomain.TimeQuery{
			From: base,
			To:   base.Add(10 * time.Minute),
		},
		Aggregation: metricsdomain.AggNone,
		Limit:       1,
		SortDesc:    true,
	}

	body, _ := json.Marshal(payload)
	handler := api.NewMetricsQueryHandler(ratelimiter.NewWindowRateLimiter(1, time.Hour), store)

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

	appStore, err := store.NewAppStore(deviceRepo, metricsRepo, configCache, nil)
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

func TestDeviceContextHandler_ReturnsDeviceContext(t *testing.T) {

	// load real store to get read hubstate data
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

	handler := api.NewDeviceContextHandler(mockStore, 1*time.Minute)
	req := httptest.NewRequest(http.MethodGet, "/api/context/devices", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if callCount != 1 {
		t.Fatalf("expected LoadHubState called once, got %d", callCount)
	}

	var resp api.DeviceContextResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Version != api.DeviceContextVersion {
		t.Fatalf("expected version %s, got %s", api.DeviceContextVersion, resp.Version)
	}

	if len(resp.Devices) == 0 {
		t.Fatalf("expected devices in context")
	}

	if resp.GeneratedAt.IsZero() {
		t.Fatalf("expected generatedAt to be set")
	}
}

func TestDeviceContextHandler_FiltersNonMeasurementExposes(t *testing.T) {
	store, err := utils_test.CreateStoreWithDevices()
	if err != nil {
		t.Fatalf("error creating store: %v", err)
	}

	hubState, _ := store.LoadHubState()

	mockStore := mocks.NewMockAppStoreWithLoadStateFunc(func() (*hub.HubState, error) {
		return hubState, nil
	})

	handler := api.NewDeviceContextHandler(mockStore, 1*time.Second)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/context/devices", nil)

	handler.ServeHTTP(rr, req)

	var resp api.DeviceContextResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	for _, d := range resp.Devices {
		for _, e := range d.Exposes {
			switch e.Name {
			case "linkquality", "battery", "battery_low":
				t.Fatalf("non-measurement expose leaked into context: %s", e.Name)
			}
		}
	}
}

func TestDeviceContextHandler_EnumAndBinaryValuesFlattened(t *testing.T) {
	store, err := utils_test.CreateStoreWithDevices()
	if err != nil {
		t.Fatalf("error creating store: %v", err)
	}

	hubState, _ := store.LoadHubState()

	mockStore := mocks.NewMockAppStoreWithLoadStateFunc(func() (*hub.HubState, error) {
		return hubState, nil
	})

	handler := api.NewDeviceContextHandler(mockStore, 1*time.Second)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/context/devices", nil)
	handler.ServeHTTP(rr, req)

	var resp api.DeviceContextResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	foundEnum := false
	foundBinary := false

	for _, d := range resp.Devices {
		for _, e := range d.Exposes {
			switch e.Type {
			case bridge.EnumDataType:
				foundEnum = true
				if len(e.Values) == 0 {
					t.Fatalf("enum expose %s has no values", e.Name)
				}

			case bridge.BinaryDataType:
				foundBinary = true
				if e.ValueOn == nil || e.ValueOff == nil {
					t.Fatalf(
						"binary expose %s missing semantics (on=%v off=%v)",
						e.Name, e.ValueOn, e.ValueOff,
					)
				}
			}
		}
	}

	if !foundEnum {
		t.Fatalf("expected at least one enum expose")
	}
	if !foundBinary {
		t.Fatalf("expected at least one binary expose")
	}
}

func TestDeviceContextHandler_RateLimit(t *testing.T) {
	store, err := utils_test.CreateStoreWithDevices()
	if err != nil {
		t.Fatalf("error creating store: %v", err)
	}

	hubState, _ := store.LoadHubState()

	mockStore := mocks.NewMockAppStoreWithLoadStateFunc(func() (*hub.HubState, error) {
		return hubState, nil
	})

	handler := api.NewDeviceContextHandler(mockStore, 1*time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/api/context/devices", nil)

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first call 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second call 429, got %d", w2.Code)
	}
}

func TestAssistantConversationsHandler(t *testing.T) {
	mockRepo := mocks.NewMockAssistantRepo()
	s, cleanup := utils_test.CreateAssistantStore(t, mockRepo, nil)
	defer cleanup()

	// Add some mock data
	now := time.Now()
	conv := assistant.NewConversation("conv-1", "hello title", now)
	mockRepo.Conversations["conv-1"] = conv

	handler := api.NewAssistantConversationsHandler(s)
	req := httptest.NewRequest(http.MethodGet, "/api/assistant/conversations", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string][]assistant.ConversationSummary
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	summaries, ok := resp["conversations"]
	if !ok {
		t.Fatalf("expected 'conversations' key in response")
	}

	if len(summaries) != 1 || summaries[0].ID != "conv-1" {
		t.Errorf("unexpected conversations payload: %+v", summaries)
	}
}

func TestAssistantHistoryHandler(t *testing.T) {
	mockRepo := mocks.NewMockAssistantRepo()
	s, cleanup := utils_test.CreateAssistantStore(t, mockRepo, nil)
	defer cleanup()

	now := time.Now()
	conv := assistant.NewConversation("conv-1", "my title", now)
	mockRepo.Conversations["conv-1"] = conv

	router := api.NewRouter()
	router.GET("/api/assistant/history/:id", api.NewAssistantHistoryHandler(s))

	req := httptest.NewRequest(http.MethodGet, "/api/assistant/history/conv-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp assistant.Conversation
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != "conv-1" {
		t.Errorf("expected ID conv-1, got %s", resp.ID)
	}
}

func TestAssistantHistoryHandler_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockAssistantRepo()
	s, cleanup := utils_test.CreateAssistantStore(t, mockRepo, nil)
	defer cleanup()

	router := api.NewRouter()
	router.GET("/api/assistant/history/:id", api.NewAssistantHistoryHandler(s))

	req := httptest.NewRequest(http.MethodGet, "/api/assistant/history/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", w.Code)
	}
}

func TestAssistantDeleteHandler(t *testing.T) {
	mockRepo := mocks.NewMockAssistantRepo()
	s, cleanup := utils_test.CreateAssistantStore(t, mockRepo, nil)
	defer cleanup()

	now := time.Now()
	conv := assistant.NewConversation("conv-1", "delete me", now)
	mockRepo.Conversations["conv-1"] = conv

	router := api.NewRouter()
	router.DELETE("/api/assistant/history/:id", api.NewAssistantDeleteHandler(s))

	req := httptest.NewRequest(http.MethodDelete, "/api/assistant/history/conv-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	if len(mockRepo.Conversations) != 0 {
		t.Error("expected conversation to be deleted")
	}
}

func TestAssistantMessageHandler(t *testing.T) {
	mockLLM := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"reply": "I am a mock assistant"}`))
	}))
	defer mockLLM.Close()

	appCfg := settings.NewAppConfig()
	appCfg.Hub.Assistant.Url = mockLLM.URL

	mockRepo := mocks.NewMockAssistantRepo()
	s, cleanup := utils_test.CreateAssistantStore(t, mockRepo, appCfg)
	defer cleanup()

	handler := api.NewAssistantMessageHandler(s)

	payload := []byte(`{"conversation_id": "test-conv", "message": "ping"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/assistant/message", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	// Verify both user and assistant messages were saved
	if !mockRepo.SaveCalled {
		t.Fatal("expected conversation to be saved")
	}

	savedConv, err := mockRepo.LoadConversation("test-conv")
	if err != nil {
		t.Fatalf("expected test-conv to exist, got err: %v", err)
	}

	if len(savedConv.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(savedConv.Messages))
	}

	if savedConv.Messages[0].Role != assistant.RoleUser || savedConv.Messages[0].Content != "ping" {
		t.Errorf("unexpected user message: %+v", savedConv.Messages[0])
	}
	if savedConv.Messages[1].Role != assistant.RoleAssistant || savedConv.Messages[1].Content != "I am a mock assistant" {
		t.Errorf("unexpected assistant message: %+v", savedConv.Messages[1])
	}
}
