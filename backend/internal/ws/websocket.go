package ws

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"node-herder/internal/auth"
	"node-herder/utils"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

const (
	// Maximum message size allowed from peer.
	maxMessageSize = 1 * 1024 * 1024
	// Socket buffers are bytes; the outbound queue is a count of messages.
	socketBufferSize = 4 * 1024
	messageQueueSize = 256

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var clientId atomic.Int64

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
	Close() error
	Start(messageHandler func(message []byte))
}

type webSocketImpl struct {
	conn *melody.Melody
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
	return &webSocketImpl{
		conn: server,
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
		utils.LogErrorf("WebSocket.Broadcast failed to marshal server payload %s", err.Error())

		return errors.New("failed to marshal server payload")
	}

	return h.conn.BroadcastFilter(bytes, func(session *melody.Session) bool { return auth.SessionValid(session.Request) })
}

func (h *webSocketImpl) Close() error {
	utils.LogDebugf("WebSocket.Close")
	return h.conn.Close()
}

func (h *webSocketImpl) Start(messageHandler func(message []byte)) {

	h.conn.HandleConnect(func(s *melody.Session) {
		if attach, ok := s.Get("attachSession"); ok {
			attach.(func(*melody.Session))(s)
		}
		utils.LogDebugf("WebSocket: New client connected")
		id := clientId.Add(1)

		s.Set("id", id)
	})

	h.conn.HandleDisconnect(func(s *melody.Session) {
		if id, ok := s.Get("id"); ok {
			utils.LogDebugf("WebSocket: client id %d disconnected", id)
		} else {
			utils.LogDebug("WebSocket: client disconnected")
		}
	})

	h.conn.HandleError(func(s *melody.Session, err error) {
		if errors.Is(err, melody.ErrMessageBufferFull) {
			// A lost state update cannot be reconciled on this connection. Close
			// directly: enqueueing another message on a full queue would recurse.
			_ = s.WebsocketConnection().Close()
			return
		}
		if id, ok := s.Get("id"); ok {
			utils.LogErrorf("WebSocket: client id %d Session error: %s\n", id, err.Error())
		} else {
			utils.LogErrorf("WebSocket: client Session error: %s\n", err.Error())
		}
	})

	h.conn.HandleClose(func(s *melody.Session, code int, reason string) error {
		if id, ok := s.Get("id"); ok {
			utils.LogDebugf("WebSocket: client id %d Session closed: %d, %s\n", id, code, reason)
		} else {
			utils.LogDebug("WebSocket: client session closed")
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
				utils.LogError("WebSocket: message handler failed; connection closed")
			}
		}()
		messageHandler(msg)
	})
}
