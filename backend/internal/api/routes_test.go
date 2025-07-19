package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"node-herder/internal/api"
	"node-herder/internal/controllers"
	"node-herder/internal/fs"
	"node-herder/mocks"
	"node-herder/models/hub"
	"node-herder/models/logging"
	utils_test "node-herder/testing"
	"node-herder/utils"
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

	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())
	bodyReader := strings.NewReader(string(missingDeviceIdPayload))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
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
	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())
	bodyReader := strings.NewReader(string("test"))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
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
	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())
	bodyReader := strings.NewReader(string(device1RootPayloadBatterySource))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
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
		t.Fatalf(err.Error())
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
	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())
	bodyReader := strings.NewReader(string(device1InvalidRootPayloadBatterySource))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
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
	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())
	bodyReader := strings.NewReader(string(device1RootPayload))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
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
		t.Fatalf(err.Error())
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
	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())
	bodyReader := strings.NewReader(string(gasNodePayload))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
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
		t.Fatalf(err.Error())
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
	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())

	bodyReader := strings.NewReader(string(device1BatterySource))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
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
	req := httptest.NewRequest(http.MethodPost, "/logfile", bodyReader)
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

	req := httptest.NewRequest(http.MethodGet, "/listlogs", nil)
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

	// Trigger dirty
	mockStore.TriggerDirty()

	// Should reload
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

}

// func TestServeHTTP_ExpirationTriggersReload(t *testing.T) {
// 	mock := &mockAppStore{
// 		state: &hub.HubState{Version: "stale"},
// 	}
// 	handler := handler.NewHubStateHandler(mock, 1*time.Millisecond)

// 	// Initial load
// 	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
// 	mock.loaded = false

// 	time.Sleep(5 * time.Millisecond)

// 	// Should reload due to TTL expiry
// 	rr := httptest.NewRecorder()
// 	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

// 	if !mock.loaded {
// 		t.Error("expected LoadHubState to be called after TTL expiration")
// 	}
// }

// func TestServeHTTP_LoadError(t *testing.T) {
// 	mock := &mockAppStore{
// 		err: errors.New("boom"),
// 	}

// 	handler := handler.NewHubStateHandler(mock, 5*time.Minute)

// 	rr := httptest.NewRecorder()
// 	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

// 	if rr.Code != http.StatusInternalServerError {
// 		t.Errorf("expected 500, got %d", rr.Code)
// 	}
// }
