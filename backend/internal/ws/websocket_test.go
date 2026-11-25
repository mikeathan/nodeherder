package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"node-herder/internal/auth"
	observationsmodel "node-herder/models/observations"
	"node-herder/utils"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

// countingLogEmitter records remote "logger" frames so a test can prove a WS callback
// did not broadcast its own diagnostic.
type countingLogEmitter struct {
	calls    atomic.Int64
	messages atomic.Value
}

func (e *countingLogEmitter) Broadcast(_ string, data interface{}) error {
	e.calls.Add(1)
	if payload, ok := data.(map[string]interface{}); ok {
		msg, _ := payload["message"].(string)
		// Case-insensitive: the pre-fix callback logged "Session error" with a capital S.
		if strings.Contains(strings.ToLower(msg), "session error") {
			previous, _ := e.messages.Load().([]string)
			e.messages.Store(append(append([]string(nil), previous...), msg))
		}
	}
	return nil
}

// rebroadcastEmitter models production wiring, where the remote logger hook emits
// through the same WS hub it is logging about.
type rebroadcastEmitter struct{ socket WebSocket }

func (e rebroadcastEmitter) Broadcast(name string, data interface{}) error {
	return e.socket.Broadcast(name, data)
}

func awaitSessions(t *testing.T, socket *webSocketImpl, want int) []*melody.Session {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		sessions, err := socket.conn.Sessions()
		if err == nil && len(sessions) == want {
			return sessions
		}
		if time.Now().After(deadline) {
			t.Fatalf("sessions = %v (err %v), want %d", sessions, err, want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// awaitOtherSession waits for a session other than the excluded (closed) one to be
// registered. Melody unregisters a closed session once readPump returns, so the two
// sessions do not necessarily coexist.
func awaitOtherSession(t *testing.T, socket *webSocketImpl, excluded *melody.Session) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		sessions, err := socket.conn.Sessions()
		if err == nil {
			for _, session := range sessions {
				if session != excluded {
					return
				}
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("no replacement session registered: %v (err %v)", sessions, err)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func closeWithoutBlocking(closeFn func()) {
	done := make(chan struct{})
	go func() { defer close(done); closeFn() }()
	select {
	case <-done:
	case <-time.After(time.Second):
	}
}

// A closed-but-registered session makes melody invoke errorHandler on its hub goroutine.
// With remote logging enabled, the old code logged that error through BroadcastFilter,
// re-entering the very goroutine that was iterating sessions: a permanent deadlock.
// The producer must keep broadcasting *while* the session closes; a single broadcast
// after it has closed misses the window and does not exercise the bug.
func TestRemoteLoggingDoesNotWedgeBroadcastOnClosedSession(t *testing.T) {
	socket := NewWebSocket().(*webSocketImpl)
	socket.Start(func([]byte) {})
	utils.RegisterRemoteLoggerHook(rebroadcastEmitter{socket: socket})
	utils.EnableRemoteLoggerHook(true)
	t.Cleanup(func() {
		utils.EnableRemoteLoggerHook(false)
		utils.RemoveRemoteLoggerHook()
		closeWithoutBlocking(func() { _ = socket.Close() })
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = socket.HandleRequest(w, r)
	}))
	defer closeWithoutBlocking(server.Close)
	defer server.CloseClientConnections()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	sessions := awaitSessions(t, socket, 1)

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			if err := socket.Broadcast("state", "ready"); err != nil {
				return
			}
		}
	}()

	// Close the session from the server side while the producer is running. The client
	// never answers the close frame, so readPump stays blocked and melody keeps the
	// session registered while the hub goroutine iterates it.
	closed := make(chan struct{})
	go func() {
		defer close(closed)
		_ = sessions[0].Close()
	}()
	select {
	case <-closed:
	case <-time.After(3 * time.Second):
		t.Fatal("session did not close")
	}

	// A fixed hub keeps serving broadcasts: the producer exits on its own only when stop
	// is closed. A wedged hub blocks BroadcastFilter forever, so wg.Wait never returns.
	time.Sleep(50 * time.Millisecond)
	close(stop)
	done := make(chan struct{})
	go func() { defer close(done); wg.Wait() }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("broadcast wedged on the melody hub goroutine")
	}

	// The hub must still register a new client and deliver to it.
	healthy, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("new client could not upgrade: %v", err)
	}
	defer healthy.Close()
	awaitOtherSession(t, socket, sessions[0])
	_ = healthy.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		if err := socket.Broadcast("state", "healthy"); err != nil {
			t.Fatal(err)
		}
		_, message, err := healthy.ReadMessage()
		if err != nil {
			t.Fatalf("healthy client received no frame: %v", err)
		}
		var event EventMessage
		if json.Unmarshal(message, &event) == nil && event.Type == "state" && event.Payload == "healthy" {
			return
		}
	}
}

// A slow client must not stall producers or healthy clients, and session failures
// must stay local instead of producing remote "logger" frames.
func TestSlowClientDoesNotBlockFanoutAndErrorsStayLocal(t *testing.T) {
	socket := NewWebSocket().(*webSocketImpl)
	connected := make(chan *melody.Session, 2)
	socket.Start(func([]byte) {})
	socket.conn.HandleConnect(func(s *melody.Session) { connected <- s })

	emitter := &countingLogEmitter{}
	utils.RegisterRemoteLoggerHook(emitter)
	utils.EnableRemoteLoggerHook(true)
	t.Cleanup(func() {
		utils.EnableRemoteLoggerHook(false)
		utils.RemoveRemoteLoggerHook()
		closeWithoutBlocking(func() { _ = socket.Close() })
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = socket.HandleRequest(w, r)
	}))
	defer closeWithoutBlocking(server.Close)
	defer server.CloseClientConnections()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	slow, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer slow.Close()
	healthy, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer healthy.Close()

	var slowSession *melody.Session
	select {
	case slowSession = <-connected:
	case <-time.After(3 * time.Second):
		t.Fatal("slow session did not connect")
	}
	select {
	case <-connected:
	case <-time.After(3 * time.Second):
		t.Fatal("healthy session did not connect")
	}

	// Drain the healthy client. Only the slow session's bounded queue is filled, so the
	// healthy client is never a saturation candidate on any CPU count.
	_ = healthy.SetReadDeadline(time.Now().Add(10 * time.Second))
	delivered := make(chan string, 1)
	go func() {
		for {
			_, message, err := healthy.ReadMessage()
			if err != nil {
				return
			}
			var event EventMessage
			if json.Unmarshal(message, &event) == nil && event.Type == "state" {
				if payload, ok := event.Payload.(string); ok && payload == "healthy" {
					select {
					case delivered <- payload:
					default:
					}
				}
			}
		}
	}()

	// Saturate only the slow session, bypassing Broadcast so the healthy client sees
	// exactly one small frame afterwards.
	payload := strings.Repeat("x", 4<<10)
	for i := 0; i <= messageQueueSize; i++ {
		_ = slowSession.Write([]byte(payload))
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < messageQueueSize; i++ {
			if err := socket.Broadcast("state", fmt.Sprintf("%d", i)); err != nil {
				return
			}
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("producer blocked on a saturated client queue")
	}

	if diagnostics := socket.Diagnostics(); diagnostics.ClientDrops == 0 {
		t.Fatalf("saturated client was not counted: %+v", diagnostics)
	}
	if reported, _ := emitter.messages.Load().([]string); len(reported) != 0 {
		t.Fatalf("WS session failures emitted remote frames: %v", reported)
	}

	if err := socket.Broadcast("state", "healthy"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-delivered:
	case <-time.After(3 * time.Second):
		t.Fatal("healthy client stopped receiving while another client was saturated")
	}
}

func TestBroadcastRejectsOversizedFrame(t *testing.T) {
	socket := NewWebSocket().(*webSocketImpl)
	socket.Start(func([]byte) {})
	server := socketServer(t, socket)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := socket.Broadcast("large", strings.Repeat("x", maxBroadcastBytes)); err == nil {
		t.Fatal("oversized frame was accepted")
	}
	if got := socket.Diagnostics().OversizedFrames; got != 1 {
		t.Fatalf("oversized frames = %d, want 1", got)
	}
	_ = conn.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("oversized frame was still delivered")
	}
}

// The largest legitimate observations snapshot must fit under the outbound bound, or a
// future cap or identity-limit change silently drops every observation update.
func TestLargestObservationSnapshotFitsBroadcastBound(t *testing.T) {
	identities := make([]observationsmodel.IdentityView, observationsmodel.MaxSnapshotIdentities)
	signal := make([]int, 20)
	reason := strings.Repeat("evidence ", 12)
	readings := make([]observationsmodel.SourceReading, 32)
	classification := &observationsmodel.ClassificationView{
		Class: "infrastructure", Reason: reason, Confidence: 0.75,
		HoursObserved: 24, DaysConsidered: 7, Advisory: true,
	}
	rssi := -60
	for i := range identities {
		identities[i] = observationsmodel.IdentityView{
			Key:            fmt.Sprintf("addr:%04x-%s", i, strings.Repeat("k", 32)),
			Kind:           "advertisement",
			Vendor:         0x004c,
			Stable:         true,
			Label:          strings.Repeat("l", 64),
			Decision:       "unknown",
			Packets:        ^uint64(0),
			RSSI:           &rssi,
			RSSISmoothed:   &rssi,
			Sources:        readings,
			Trend:          "approaching",
			TrendDeltaDB:   8.5,
			Signal:         signal,
			Classification: classification,
		}
	}
	snapshot := observationsmodel.Snapshot{
		Identities: identities,
		Sources: []observationsmodel.SourceDescriptor{{
			ID: "ble-1", Name: strings.Repeat("n", 64),
			Media:  []observationsmodel.Medium{observationsmodel.MediumBLEAdvertisement},
			Facets: []string{"kind", "vendor", "stable"},
			Diagnostics: observationsmodel.SourceDiagnostics{
				Advertisements: ^uint64(0), Decoded: ^uint64(0), Dropped: ^uint64(0),
			},
		}},
		Arrivals: 256, Truncated: true, Observing: true,
	}

	socket := NewWebSocket().(*webSocketImpl)
	socket.Start(func([]byte) {})
	server := socketServer(t, socket)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := socket.Broadcast("observations", snapshot); err != nil {
		t.Fatalf("largest legitimate snapshot was rejected: %v", err)
	}
	frame, err := json.Marshal(EventMessage{Type: "observations", Payload: snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if len(frame) >= maxBroadcastBytes {
		t.Fatalf("snapshot frame %d bytes leaves no headroom under maxBroadcastBytes %d", len(frame), maxBroadcastBytes)
	}
}

func socketServer(t *testing.T, socket *webSocketImpl, middlewares ...func(http.Handler) http.Handler) *httptest.Server {
	t.Helper()
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = socket.HandleRequest(w, r) })
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	t.Cleanup(func() { _ = socket.Close() })
	return server
}
func TestWebSocketConnectionMemoryAndLargeMessages(t *testing.T) {
	socket := NewWebSocket().(*webSocketImpl)
	socket.Start(func([]byte) {})
	server := socketServer(t, socket)
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	runtime.GC()
	runtime.ReadMemStats(&after)
	if delta := int64(after.HeapAlloc) - int64(before.HeapAlloc); delta > 4<<20 {
		t.Fatalf("one idle connection allocated %d bytes", delta)
	}
	payload := strings.Repeat("x", 100<<10)
	if err := socket.Broadcast("large", payload); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, body, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), payload) {
		t.Fatal("small buffers truncated a large message")
	}
	conn.Close()
	deadline := time.Now().Add(5 * time.Second)
	for socket.conn.Len() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("disconnected session retained")
		}
		runtime.Gosched()
	}
}
func TestWebSocketClosesSlowClientOnQueueOverflow(t *testing.T) {
	socket := NewWebSocket().(*webSocketImpl)
	socket.Start(func([]byte) {})
	fill := make(chan struct{})
	done := make(chan struct{})
	// Hold the producer before Melody starts its write pump, making overflow deterministic.
	socket.conn.HandleConnect(func(session *melody.Session) {
		<-fill
		for i := 0; i <= messageQueueSize; i++ {
			_ = session.Write([]byte("state"))
		}
		close(done)
	})
	server := socketServer(t, socket)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	close(fill)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("queue producer blocked")
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("overflowed client kept receiving an incomplete stream")
	}
}

func TestWebSocketDisconnectPreservesEventContract(t *testing.T) {
	socket := NewWebSocket().(*webSocketImpl)
	socket.Start(func([]byte) {})
	disconnected := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = socket.HandleRequest(w, r)
		if r.URL.Path == "/closing" {
			close(disconnected)
		}
	}))
	defer server.Close()
	defer socket.Close()
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	remaining, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer remaining.Close()
	closing, _, err := websocket.DefaultDialer.Dial(url+"/closing", nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = closing.Close()
	select {
	case <-disconnected:
	case <-time.After(5 * time.Second):
		t.Fatal("disconnect handler did not finish")
	}
	if err := socket.Broadcast("state", "ready"); err != nil {
		t.Fatal(err)
	}
	_ = remaining.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := remaining.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var event EventMessage
	if err := json.Unmarshal(message, &event); err != nil || event.Type != "state" {
		t.Fatalf("disconnect emitted a non-contract frame: %q", message)
	}
}

func TestWebSocketIncomingMessageLimit(t *testing.T) {
	socket := NewWebSocket().(*webSocketImpl)
	received := make(chan int, 2)
	socket.Start(func(message []byte) { received <- len(message) })
	server := socketServer(t, socket)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.WriteMessage(websocket.TextMessage, []byte(strings.Repeat("x", maxMessageSize))); err != nil {
		t.Fatal(err)
	}
	select {
	case size := <-received:
		if size != maxMessageSize {
			t.Fatalf("valid message truncated: %d", size)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("valid message was not dispatched")
	}
	// The read pump enforces the configured limit before domain dispatch.
	_ = conn.WriteMessage(websocket.TextMessage, []byte(strings.Repeat("x", maxMessageSize+1)))
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseMessageTooBig) {
		t.Fatalf("oversized message was not rejected with 1009: %v", err)
	}
	select {
	case <-received:
		t.Fatal("oversized message reached the domain handler")
	default:
	}
}

func TestWebSocketClosesRevokedAndExpiredSessions(t *testing.T) {
	for _, expired := range []bool{false, true} {
		t.Run(map[bool]string{false: "revoked", true: "expired"}[expired], func(t *testing.T) {
			lifetime := time.Hour
			if expired {
				lifetime = 2 * time.Second
			}
			jwt := auth.NewJWTService(auth.JWTConfig{Secret: []byte("test-only-secret"), Expiration: lifetime})
			blacklist := auth.NewTokenBlacklist()
			socket := NewWebSocket().(*webSocketImpl)
			received := make(chan struct{}, 1)
			socket.Start(func([]byte) { received <- struct{}{} })
			server := socketServer(t, socket, auth.Auth(jwt, blacklist))
			token, err := jwt.GenerateJWT("user", "User")
			if err != nil {
				t.Fatal(err)
			}
			headers := http.Header{"Authorization": []string{"Bearer " + token}}
			conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), headers)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			if !expired {
				blacklist.Add(token, time.Now().Add(time.Hour))
			}
			_ = conn.SetReadDeadline(time.Now().Add(4 * time.Second))
			if _, _, err := conn.ReadMessage(); err == nil {
				t.Fatal("invalidated session remained open")
			}
			_ = conn.WriteMessage(websocket.TextMessage, []byte("safe-test"))
			select {
			case <-received:
				t.Fatal("invalidated session dispatched a message")
			default:
			}
		})
	}
}

func TestWebSocketOriginPolicy(t *testing.T) {
	t.Setenv("FRONTEND_ALLOWED_ORIGINS", "https://frontend.example")
	socket := NewWebSocket().(*webSocketImpl)
	socket.Start(func([]byte) {})
	server := socketServer(t, socket)
	for _, origin := range []string{"", server.URL, "https://frontend.example", "https://untrusted.example"} {
		header := http.Header{}
		if origin != "" {
			header.Set("Origin", origin)
		}
		conn, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), header)
		if origin == "https://untrusted.example" {
			if err == nil {
				conn.Close()
				t.Fatal("unlisted origin accepted")
			}
			if response == nil || response.StatusCode != 403 {
				t.Fatalf("origin rejection: %v", response)
			}
		} else {
			if err != nil {
				t.Fatalf("allowed origin %q: %v", origin, err)
			}
			conn.Close()
		}
	}
}

func BenchmarkWebSocketConnection(b *testing.B) {
	socket := NewWebSocket().(*webSocketImpl)
	socket.Start(func([]byte) {})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = socket.HandleRequest(w, r) }))
	defer server.Close()
	defer socket.Close()
	address := "ws" + strings.TrimPrefix(server.URL, "http")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(address, nil)
		if err != nil {
			b.Fatal(err)
		}
		conn.Close()
	}
}
