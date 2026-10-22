package hub

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"node-herder/internal/automations"
	"node-herder/internal/controllers"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/testing/hubharness"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"
)

const browserDialID = "0x0000000000000001"

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
	router, err := newRouter(4120, bus, h.Hub, h.Store, nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	router.PublicGET("/api/__test/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Serialize fixture controls. Queues are consumed before a response, so the
	// next identical press cannot observe the previous event or acknowledgement.
	var controlMu sync.Mutex
	router.POST("/api/__test/reset", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controlMu.Lock()
		defer controlMu.Unlock()
		err := h.EventHub.DeleteAutomation(map[string]any{"id": browserDialID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// A successful run finishes OFF. Refuse stale state rather than silently
		// acknowledging a reset that has not travelled through the device lane.
		lightMu.Lock()
		on := lightOn
		lightMu.Unlock()
		if on {
			http.Error(w, "previous browser run did not finish OFF", http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusNoContent)
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
