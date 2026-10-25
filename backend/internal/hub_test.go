package hub

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	defaults "node-herder/configs/integrations"
	"node-herder/internal/automations"
	"node-herder/internal/controllers"
	connections "node-herder/internal/integrations"
	mcpserver "node-herder/internal/mcp/server"
	metricservices "node-herder/internal/metrics/services"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/repository"
	"node-herder/testing/hubharness"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

const browserDialID = "0x0000000000000001"

// drainChannel empties a buffered fixture queue without blocking.
func drainChannel[T any](ch <-chan T) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

// Synthetic Zigbee2MQTT metadata. This fixture never creates a network MQTT client.
var browserBridge = []byte(`[
 {"ieee_address":"0x0000000000000001","friendly_name":"Test dial","type":"EndDevice","power_source":"Battery","interview_completed":true,
  "definition":{"description":"Test dial","exposes":[
   {"type":"enum","name":"action","property":"action","access":1,"values":["idle","button_1_press_release"]},
   {"type":"numeric","name":"linkquality","property":"linkquality","access":1,"value_min":0,"value_max":255}]}},
 {"ieee_address":"0x0000000000000002","friendly_name":"Test light","type":"Router","power_source":"Mains (single phase)","interview_completed":true,
  "definition":{"description":"Test light","exposes":[{"type":"light","features":[
   {"type":"binary","name":"state","property":"state","access":7,"value_on":"ON","value_off":"OFF","value_toggle":"TOGGLE"},
   {"type":"numeric","name":"brightness","property":"brightness","access":7,"value_min":0,"value_max":254}]}]}}
]`)

// TestAutomationBrowserHub is an opt-in server owned by Playwright's webServer.
// It runs production auth, HTTP, WS, ingestion and automations over fake MQTT.
// Ordinary go test runs skip it. Signals and the Go test timeout bound its lifetime.
func TestAutomationBrowserHub(t *testing.T) {
	if os.Getenv("NODEHERDER_BROWSER_HUB") != "1" {
		t.Skip("only run as an isolated Playwright fixture")
	}
	for key, value := range map[string]string{
		"APP_ENV": "development", "FRONTEND_BASE_URL": "http://localhost:4100",
		"FRONTEND_ALLOWED_ORIGINS": "http://localhost:4100", "JWT_SECRET_KEY": "isolated-browser-fixture-only",
		"OAUTH_CALLBACK_URL": "http://localhost:4120/api/auth/callback", "OFFLINE_STRICT_LOCAL": "true",
		"AUTH_COOKIE_SECURE": "false", "TRUSTED_PROXY_CIDRS": "",
	} {
		t.Setenv(key, value)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	bus := ws.NewWsHub()
	bus.Start()
	t.Cleanup(func() { _ = bus.Close() })
	h := hubharness.NewFromBridgeWithEventHub(t, browserBridge, 2, bus,
		controllers.WithContext(ctx), controllers.WithAutomationHandlers(automations.DefaultAutomationHandlers(ctx)))
	h.Seed(t, map[string]map[string]any{
		"Test dial":  {"action": "idle", "linkquality": 100.0},
		"Test light": {"state": "OFF", "brightness": 100.0},
	})

	commands := make(chan map[string]any, 16)
	reports := make(chan string, 16)
	dialDone := make(chan struct{}, 16)
	var lightMu sync.Mutex
	lightOn := false
	h.Paho.OnPublish(func(topic string, payload []byte) {
		if topic != "zigbee2mqtt/Test light/set" {
			return
		}
		var command map[string]any
		if err := json.Unmarshal(payload, &command); err != nil {
			t.Errorf("invalid MQTT command: %v", err)
			return
		}
		select {
		case commands <- command:
		default:
			t.Error("unexpected MQTT command overflow")
		}
		if command["state"] != "TOGGLE" {
			return // a wrong command must not receive a fake successful toggle report
		}
		lightMu.Lock()
		lightOn = !lightOn
		state := "OFF"
		if lightOn {
			state = "ON"
		}
		lightMu.Unlock()
		payload, _ = json.Marshal(map[string]any{"state": state})
		h.Paho.Deliver("zigbee2mqtt/Test light", payload)
	})
	h.EventHub.OnBroadcast(func(event string, data interface{}) {
		p, ok := data.(*devices.UpdatePackage)
		if !ok || event != ws.DeviceUpdated {
			return
		}
		if p.Id == browserDialID {
			select {
			case dialDone <- struct{}{}:
			default:
				t.Error("unexpected dial event overflow")
			}
		} else if state, ok := p.Data["state"].(string); ok {
			select {
			case reports <- state:
			default:
				t.Error("unexpected light report overflow")
			}
		}
	})
	connectionService := browserConnectionService(t, ctx)
	router, err := newRouter(4120, bus, h.Hub, h.Store, nil, ctx, WithConnections(connectionService))
	if err != nil {
		t.Fatal(err)
	}
	router.PublicGET("/api/__test/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"fixture": "nodeherder-go-browser", "run_id": os.Getenv("NODEHERDER_E2E_RUN_ID")})
	}))

	router.POST("/api/__test/connections/reset", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, e := connectionService.Instances()
		if e != nil {
			http.Error(w, "reset failed", 500)
			return
		}
		for _, item := range items {
			if e = connectionService.DeleteInstance(item.ID, item.ETag); e != nil {
				http.Error(w, "reset failed", 500)
				return
			}
		}
		definition, e := connectionService.Definition("waste_collection/whenisbins")
		if e != nil {
			http.Error(w, "reset failed", 500)
			return
		}
		seed, _ := defaults.Files.ReadFile("waste_collection/whenisbins.yaml")
		if _, e = connectionService.SaveDefinition(definition.Key, string(seed), definition.ETag, false, nil); e != nil {
			http.Error(w, "reset failed", 500)
			return
		}
		w.WriteHeader(204)
	}))
	// Serialize fixture controls. Queues are consumed before a response, so the
	// next identical press cannot observe the previous event or acknowledgement.
	var controlMu sync.Mutex
	router.POST("/api/__test/reset", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controlMu.Lock()
		defer controlMu.Unlock()
		drainChannel(commands)
		drainChannel(reports)
		drainChannel(dialDone)
		err := h.EventHub.DeleteAutomation(map[string]any{"id": browserDialID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// A failed run can leave the light ON. Report it OFF through the same lane
		// as a device confirmation so the next case starts from the seeded state.
		lightMu.Lock()
		wasOn := lightOn
		lightOn = false
		lightMu.Unlock()
		if !wasOn {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.Paho.Deliver("zigbee2mqtt/Test light", []byte(`{"state":"OFF"}`))
		timeout := time.After(hubharness.Timeout)
		for {
			select {
			case state := <-reports:
				if state == "OFF" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			case <-timeout:
				http.Error(w, "light did not confirm OFF during reset", http.StatusGatewayTimeout)
				return
			}
		}
	}))
	router.POST("/api/__test/dial", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controlMu.Lock()
		defer controlMu.Unlock()
		requestCtx, stop := context.WithTimeout(r.Context(), hubharness.Timeout)
		defer stop()
		h.Paho.Deliver("zigbee2mqtt/Test dial", []byte(`{"action":"button_1_press_release"}`))
		select {
		case <-dialDone: // automation execution precedes this broadcast
		case <-requestCtx.Done():
			http.Error(w, "dial event was not processed", http.StatusGatewayTimeout)
			return
		}
		var command map[string]any
		select {
		case command = <-commands:
		case <-requestCtx.Done():
			http.Error(w, "automation did not publish", http.StatusGatewayTimeout)
			return
		}
		select {
		case <-commands:
			http.Error(w, "one dial event published duplicate commands", http.StatusConflict)
			return
		default:
		}
		select {
		case state := <-reports:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"command": command, "confirmedState": state})
		case <-requestCtx.Done():
			http.Error(w, "light did not confirm", http.StatusGatewayTimeout)
		}
	}))
	listener, err := net.Listen("tcp", "127.0.0.1:4120")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: router, ReadHeaderTimeout: 5 * time.Second}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	select {
	case <-ctx.Done():
	case err := <-served:
		t.Errorf("browser hub stopped unexpectedly: %v", err)
	}
	stopCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	if err := server.Shutdown(stopCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		t.Error(err)
	}
}

// Connections browser traffic uses a synthetic provider and temp encrypted storage, never the network.
type browserProvider struct{}

func (browserProvider) Do(r *http.Request) (*http.Response, error) {
	var body string
	switch {
	case strings.HasSuffix(r.URL.Path, "/addresses"):
		body = `{"required_input":"property_id","candidates":[{"id":"p:synthetic","label":"1 Synthetic Street, Test Town"}]}`
	case strings.HasSuffix(r.URL.Path, "/lookups"), strings.Contains(r.URL.Path, "/schedules/"):
		now := time.Now().UTC()
		result := map[string]any{"property_id": "p:synthetic", "address_match": "exact", "evidence_granularity": "property", "date_confidence": "published_calendar", "date_completeness": "limited_horizon", "retrieved_at": now.Format(time.RFC3339), "refresh_due_at": now.Add(24 * time.Hour).Format(time.RFC3339), "collections": []any{map[string]any{"name": "Paper and card", "waste_type": "paper_card", "bin_colour": "blue", "dates": []string{now.AddDate(0, 0, 7).Format("2006-01-02")}, "dates_complete": true}}}
		result["bins"] = result["collections"]
		value := any(result)
		if strings.HasSuffix(r.URL.Path, "/lookups") {
			value = map[string]any{"id": "synthetic-lookup", "status": "done", "result": result}
		}
		raw, _ := json.Marshal(value)
		body = string(raw)
	default:
		body = `{"problem":"not_found"}`
		return &http.Response{StatusCode: 404, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	}
	return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
}
func browserConnectionService(t *testing.T, ctx context.Context) *connections.Service {
	t.Helper()
	repo, e := repository.OpenIntegrationRepository(t.TempDir(), repository.IntegrationKeyring{ActiveKeyID: "fixture", Keys: map[string]string{"fixture": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32))}})
	if e != nil {
		t.Fatal(e)
	}
	files, e := repository.OpenIntegrationDefinitions(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	service, e := connections.NewService(ctx, connections.Options{Repository: repo, Definitions: files, HTTP: browserProvider{}})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if e := service.Close(closeCtx); e != nil {
			t.Error(e)
		}
	})
	return service
}

func TestConnectionsDoNotEnterPublicDevicesOrMCP(t *testing.T) {
	for key, value := range map[string]string{"APP_ENV": "development", "OAUTH_CALLBACK_URL": "http://localhost:4120/api/auth/callback", "JWT_SECRET_KEY": "isolated-public-boundary"} {
		t.Setenv(key, value)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h := hubharness.NewFromBridge(t, browserBridge, 2)
	service := browserConnectionService(t, ctx)
	p, e := service.CreateSetup("waste_collection/whenisbins", "", "")
	require.NoError(t, e)
	_, e = service.SearchSetup(p.ID, "XX1 1XX", p.ETag)
	require.NoError(t, e)
	require.Eventually(t, func() bool { var e error; p, e = service.Setup(p.ID); return e == nil && p.Status == "choose" }, 3*time.Second, 10*time.Millisecond)
	p, e = service.SaveSetup(p.ID, "p:synthetic", "Private schedule sentinel", p.ETag)
	require.NoError(t, e)
	require.Eventually(t, func() bool { i, e := service.Instance(p.InstanceID); return e == nil && len(i.Resources) == 1 }, 3*time.Second, 10*time.Millisecond)
	mcp := mcpserver.New(h.Store, h.Store.AppConfig(), metricservices.NewQueryService(h.Store))
	router, e := newRouter(4120, h.EventHub, h.Hub, h.Store, mcp, ctx, WithConnections(service))
	require.NoError(t, e)
	requests := []struct{ method, path, body string }{{"GET", "/api/context/devices", ""}, {"POST", "/api/mcp", `{"jsonrpc":"2.0","id":1,"method":"resources/list"}`}, {"POST", "/api/mcp", `{"jsonrpc":"2.0","id":2,"method":"resources/read","params":{"uri":"nodeherder://system-prompt"}}`}}
	for _, request := range requests {
		r := httptest.NewRequest(request.method, request.path, strings.NewReader(request.body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		require.Equal(t, 200, w.Code, w.Body.String())
		require.NotContains(t, w.Body.String(), `"error"`)
		for _, private := range []string{"Private schedule sentinel", "waste_schedule/v1", "p:synthetic"} {
			require.NotContains(t, w.Body.String(), private)
		}
	}
	denied := httptest.NewRecorder()
	router.ServeHTTP(denied, httptest.NewRequest("GET", "/api/integrations/catalog", nil))
	require.Equal(t, 401, denied.Code)
	state, e := h.Store.LoadHubState()
	require.NoError(t, e)
	wire, e := json.Marshal(state)
	require.NoError(t, e)
	require.NotContains(t, string(wire), "Private schedule sentinel")
	require.NotContains(t, string(wire), "waste_schedule/v1")
}
