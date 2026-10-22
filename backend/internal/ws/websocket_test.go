package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"node-herder/internal/auth"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

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
