package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"node-herder/internal/auth"
	"node-herder/utils"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

const (
	// maxMessageSize bounds a client frame on the way in; melody enforces it on reads.
	maxMessageSize = 1 * 1024 * 1024
	// maxBroadcastBytes bounds a server-generated frame on the way out. It is larger than
	// the inbound limit because a full metrics range can legitimately exceed any client
	// request, while the largest observations snapshot is capped by MaxSnapshotIdentities.
	// This bound stops a runaway payload from filling every client's queue; a test asserts
	// that the largest legitimate snapshot fits.
	maxBroadcastBytes = 8 * maxMessageSize
	// Socket buffers are bytes; the outbound queue is a count of messages.
	socketBufferSize = 4 * 1024
	messageQueueSize = 256

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// sessionErrorLogEvery bounds local diagnostics. A log line from a melody callback
	// must not broadcast, and a slow client must not fill the log with one line per frame.
	sessionErrorLogEvery = 100
)

var clientId atomic.Int64

// BroadcastDiagnostics is the observable failure state of one WS server. All counts are
// monotonic for the process lifetime and are read locally, never published as frames.
type BroadcastDiagnostics struct {
	// SessionErrors counts melody errorHandler callbacks, including normal disconnects.
	SessionErrors uint64
	// ClientDrops counts callbacks for a full queue or a closed session, which is the
	// condition that detaches a client; a slow client raises this, not DroppedFrames.
	ClientDrops uint64
	// DroppedFrames counts frames rejected by Session.Write (a closed session). Melody
	// reports a full queue through its error callback instead, so it is counted above.
	DroppedFrames uint64
	// OversizedFrames counts frames rejected for exceeding maxBroadcastBytes.
	OversizedFrames uint64
}

type EventMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (e EventMessage) MarshalJSON() ([]byte, error) {
	p := e.Payload
	if v, ok := e.Payload.([]byte); ok {
		p = string(v)
	}
	return json.Marshal(&struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}{
		Type:    e.Type,
		Payload: p,
	})
}

type WebSocket interface {
	HandleRequest(w http.ResponseWriter, r *http.Request) error
	Broadcast(eventName string, data interface{}) error
	Diagnostics() BroadcastDiagnostics
	Close() error
	Start(messageHandler func(message []byte))
}

type webSocketImpl struct {
	conn *melody.Melody
	// fanoutMu keeps the per-session write order stable across concurrent producers.
	// Invariant: nothing reachable from inside the locked region (Session.Write and its
	// synchronous errorHandler callback) may emit a hook-bearing log or call Broadcast,
	// because the mutex is not reentrant. Session validity is checked before the lock.
	fanoutMu sync.Mutex

	// sessions is the fan-out registry. melody registers a session by sending on an
	// unbuffered channel and only then runs the connect callback, so the hub's own session
	// set can still be missing the session when that callback runs. A snapshot from
	// melody.Sessions() right then omits it, and a response to the client's first request
	// is dropped: the client waits forever. This registry is populated in the connect
	// callback, before melody starts writePump/readPump, so a message cannot be handled
	// before its session is broadcast-visible. Guarded by sessionsMu; never held with
	// fanoutMu.
	sessionsMu sync.RWMutex
	sessions   map[*melody.Session]struct{}
	// closed reports whether Close has run, so Broadcast keeps returning ErrClosed instead
	// of writing to sessions that are being torn down.
	closed atomic.Bool

	sessionErrors   atomic.Uint64
	clientDrops     atomic.Uint64
	droppedFrames   atomic.Uint64
	oversizedFrames atomic.Uint64
}

// Diagnostics reports fan-out failures for health checks. It never logs or publishes a
// frame, so reading it cannot amplify a slow client through the logger.
func (h *webSocketImpl) Diagnostics() BroadcastDiagnostics {
	return BroadcastDiagnostics{
		SessionErrors:   h.sessionErrors.Load(),
		ClientDrops:     h.clientDrops.Load(),
		DroppedFrames:   h.droppedFrames.Load(),
		OversizedFrames: h.oversizedFrames.Load(),
	}
}

func NewWebSocket() WebSocket {
	server := melody.New()
	server.Upgrader.ReadBufferSize = socketBufferSize
	server.Upgrader.WriteBufferSize = socketBufferSize
	server.Config.PingPeriod = pingPeriod
	server.Config.PongWait = pongWait
	server.Config.WriteWait = writeWait
	server.Config.MaxMessageSize = maxMessageSize
	server.Config.MessageBufferSize = messageQueueSize
	server.Upgrader.CheckOrigin = allowedWebSocketOrigin(utils.GetFrontendAllowedOrigins())
	// melody.New starts the hub goroutine, which marks the instance open as its first
	// statement. A request accepted before that step is rejected as closed, so the first
	// client can receive a bad handshake. Wait for the goroutine to reach its loop; the
	// bound keeps a pathological scheduler from stalling construction.
	for attempt := 0; attempt < 1000; attempt++ {
		if _, err := server.Sessions(); err == nil {
			break
		}
		runtime.Gosched()
	}
	return &webSocketImpl{
		conn:     server,
		sessions: make(map[*melody.Session]struct{}),
	}
}

func allowedWebSocketOrigin(origins []string) func(*http.Request) bool {
	allowed := make(map[string]bool, len(origins))
	for _, origin := range origins {
		allowed[origin] = true
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" || allowed[origin] {
			return true
		}
		parsed, err := url.Parse(origin)
		return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") &&
			parsed.User == nil && parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == "" &&
			strings.EqualFold(parsed.Host, r.Host)
	}
}

func (h *webSocketImpl) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var connection *websocket.Conn
	stop := auth.WatchSession(r, func() {
		mu.Lock()
		defer mu.Unlock()
		if connection != nil {
			_ = connection.Close()
		}
	})
	defer stop()
	attach := func(session *melody.Session) {
		mu.Lock()
		defer mu.Unlock()
		connection = session.WebsocketConnection()
		if !auth.SessionValid(r) {
			_ = connection.Close()
		}
	}
	return h.conn.HandleRequestWithKeys(w, r, map[string]any{"attachSession": attach})
}

func (h *webSocketImpl) Broadcast(eventName string, data interface{}) error {
	var wsData = EventMessage{Type: eventName, Payload: data}

	bytes, err := json.Marshal(wsData)
	if err != nil {
		utils.LogLocalErrorf("WebSocket.Broadcast failed to marshal server payload %s", err.Error())
		return errors.New("failed to marshal server payload")
	}
	if len(bytes) > maxBroadcastBytes {
		h.oversizedFrames.Add(1)
		return fmt.Errorf("websocket frame %d bytes exceeds the %d byte limit", len(bytes), maxBroadcastBytes)
	}

	// BroadcastFilter sends on an unbuffered channel owned by melody's single hub goroutine.
	// A log line emitted from that goroutine re-enters Broadcast through the remote logger
	// hook and wedges the hub permanently. Fan out from the producer instead, from this
	// server's own registry: melody.Sessions() can still be missing a session that is inside
	// its connect callback, and a first request's response would then be dropped. Session
	// validity is checked before fanoutMu; Session.Write is a non-blocking enqueue onto the
	// per-client bounded queue.
	if h.closed.Load() {
		return melody.ErrClosed
	}
	eligible := h.fanoutSessions()

	h.fanoutMu.Lock()
	defer h.fanoutMu.Unlock()
	for _, session := range eligible {
		if err := session.Write(bytes); err != nil {
			h.droppedFrames.Add(1)
			_ = session.WebsocketConnection().Close()
		}
	}
	return nil
}

// fanoutSessions snapshots the registered sessions that are still authorised. It copies
// under sessionsMu so the write loop holds the registry lock for as short a time as
// possible; sessionsMu is never acquired while fanoutMu is held.
func (h *webSocketImpl) fanoutSessions() []*melody.Session {
	h.sessionsMu.RLock()
	defer h.sessionsMu.RUnlock()
	eligible := make([]*melody.Session, 0, len(h.sessions))
	for session := range h.sessions {
		if auth.SessionValid(session.Request) {
			eligible = append(eligible, session)
		}
	}
	return eligible
}

// trackSession makes a new session broadcast-visible. melody runs the connect callback
// before writePump/readPump, so registering here precedes any message the session can
// send; the hub's own add may still be pending.
func (h *webSocketImpl) trackSession(session *melody.Session) {
	h.sessionsMu.Lock()
	h.sessions[session] = struct{}{}
	h.sessionsMu.Unlock()
}

func (h *webSocketImpl) forgetSession(session *melody.Session) {
	h.sessionsMu.Lock()
	delete(h.sessions, session)
	h.sessionsMu.Unlock()
}

func (h *webSocketImpl) Close() error {
	h.closed.Store(true)
	utils.LogLocalDebugf("WebSocket.Close")
	return h.conn.Close()
}

func (h *webSocketImpl) Start(messageHandler func(message []byte)) {

	h.conn.HandleConnect(func(s *melody.Session) {
		// Make the session broadcast-visible before it can send or receive a message.
		// melody has already upgraded and sent on its unbuffered register channel, but the
		// hub goroutine may not have added the session to its set yet, so a producer-side
		// snapshot there would omit it.
		h.trackSession(s)
		if attach, ok := s.Get("attachSession"); ok {
			attach.(func(*melody.Session))(s)
		}
		utils.LogLocalDebugf("WebSocket: New client connected")
		id := clientId.Add(1)

		s.Set("id", id)
	})

	h.conn.HandleDisconnect(func(s *melody.Session) {
		h.forgetSession(s)
		if id, ok := s.Get("id"); ok {
			utils.LogLocalDebugf("WebSocket: client id %d disconnected", id)
		} else {
			utils.LogLocalDebug("WebSocket: client disconnected")
		}
	})

	// Melody invokes this callback from its hub goroutine (via writeMessage during a
	// broadcast), from each session's writePump and from readPump. It must never emit a
	// remotely delivered log line: that would call Broadcast and re-enter the hub goroutine.
	h.conn.HandleError(func(s *melody.Session, err error) {
		total := h.sessionErrors.Add(1)
		overflow := errors.Is(err, melody.ErrMessageBufferFull) || errors.Is(err, melody.ErrWriteClosed)
		if overflow {
			h.clientDrops.Add(1)
			// Closing the raw connection makes readPump return so melody unregisters the
			// session; a full queue has already lost the frame and cannot be reconciled.
			_ = s.WebsocketConnection().Close()
		}
		if total == 1 || total%sessionErrorLogEvery == 0 {
			if id, ok := s.Get("id"); ok {
				utils.LogLocalWarnf("WebSocket: client id %d session error (%d total): %s", id, total, err.Error())
			} else {
				utils.LogLocalWarnf("WebSocket: session error (%d total): %s", total, err.Error())
			}
		}
	})

	h.conn.HandleClose(func(s *melody.Session, code int, reason string) error {
		if id, ok := s.Get("id"); ok {
			utils.LogLocalDebugf("WebSocket: client id %d Session closed: %d, %s", id, code, reason)
		} else {
			utils.LogLocalDebug("WebSocket: client session closed")
		}

		return nil
	})

	h.conn.HandleMessage(func(s *melody.Session, msg []byte) {
		if !auth.SessionValid(s.Request) {
			_ = s.WebsocketConnection().Close()
			return
		}
		defer func() {
			if recover() != nil {
				_ = s.WebsocketConnection().Close()
				utils.LogLocalWarnf("WebSocket: message handler failed; connection closed")
			}
		}()
		messageHandler(msg)
	})
}
