package routes_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"node-herder/mocks"
	repository "node-herder/repository/devices"
	"node-herder/transport/controllers"
	"node-herder/transport/routes"
	"strings"
	"testing"
	"time"
)

const device1BatterySource = `{"id":"device 1","conn":"mqtt","power_source":"battery","humidity":92.49999999999999,"temperature":19.000000000000004,"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":47,"battery":98}`
const device1RootPayloadBatterySource = `{"id":"device 1", "timestamp":"2023-08-01T16:30:04Z", "readings":{"humidity":92.49999999999999,"temperature":19.000000000000004}}`
const device1InvalidRootPayloadBatterySource = `{"id":"device 1", "timestamp":"2023-08-01T16:30:04Z", "readingstest":{"humidity":92.49999999999999,"temperature":19.000000000000004}}`
const missingDeviceIdPayload = `{"conn":"mqtt","power_source":"battery","humidity":92.49999999999999,"temperature":19.000000000000004,"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":47,"battery":98}`

func TestHandleMissingDeviceIdPayload(t *testing.T) {
	repo := repository.NewMemoryDeviceRepo()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, mqtt, repo, context.Background())
	bodyReader := strings.NewReader(string(missingDeviceIdPayload))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h := routes.NewDataCollectorHandler(hub)
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
	repo := repository.NewMemoryDeviceRepo()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, mqtt, repo, context.Background())
	bodyReader := strings.NewReader(string("test"))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h := routes.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestHandleSuccesfullyRootPayload(t *testing.T) {

	id := "device 1"
	repo := repository.NewMemoryDeviceRepo()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, mqtt, repo, context.Background())
	bodyReader := strings.NewReader(string(device1RootPayloadBatterySource))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h := routes.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	time.Sleep(200 * time.Millisecond)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	device, err := repo.FindDevice(id)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if device == nil {
		t.Fatalf("want %s got %s", id, "nil")
	}
	if device.Id != id {
		t.Fatalf("want %s got %s", id, device.Id)
	}
}

func TestHandleInvalidRootPayload(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, mqtt, repo, context.Background())
	bodyReader := strings.NewReader(string(device1InvalidRootPayloadBatterySource))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h := routes.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	body, _ := io.ReadAll(w.Body)
	expectedBody := "invalid data: data structure not containign valid root payload\n"
	if string(body) != expectedBody {
		t.Errorf("error reading body got %v want %v", string(body), expectedBody)
	}
}

func TestProcessorHandleRootPayloadWithTimestamp(t *testing.T) {

	timestamp := "2023-07-20T19:48:35+01:00"

	id := "device1"
	device1RootPayload := `{"nickname":"device1","timestamp":"2023-07-20T19:48:35+01:00", "readings":{"humidity":92.1,"temperature":19.3}}`

	repo := repository.NewMemoryDeviceRepo()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, mqtt, repo, context.Background())
	bodyReader := strings.NewReader(string(device1RootPayload))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h := routes.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	want, _ := time.Parse(time.RFC3339, timestamp)

	time.Sleep(100 * time.Millisecond)
	device, err := repo.FindDevice(id)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if device.Id != id {
		t.Fatalf("want %s got %s", id, device.Id)
	}

	if device.Stats["last_seen"] == nil {
		t.Fatalf("want %s got %s", "last_seen", "nil")
	}

	if device.Stats["last_seen"] != want {
		t.Fatalf("want %s got %s", want, device.Stats["last_seen"])

	}
}
func TestHandleSuccesfullyPayload(t *testing.T) {
	id := "device 1"
	repo := repository.NewMemoryDeviceRepo()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, mqtt, repo, context.Background())
	bodyReader := strings.NewReader(string(device1BatterySource))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h := routes.NewDataCollectorHandler(hub)
	h.ServeHTTP(w, req)

	time.Sleep(200 * time.Millisecond)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	device, err := repo.FindDevice(id)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if device == nil {
		t.Fatalf("want %s got %s", id, "nil")
	}
	if device.Id != id {
		t.Fatalf("want %s got %s", id, device.Id)
	}
}

func TestHandleUnsuportedMediaType(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, mqtt, repo, context.Background())

	bodyReader := strings.NewReader(string(device1BatterySource))
	req := httptest.NewRequest(http.MethodPost, "/collect", bodyReader)
	w := httptest.NewRecorder()

	h := routes.NewDataCollectorHandler(hub)
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
