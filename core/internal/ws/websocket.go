package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"node-herder/utils"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

const (
	// Maximum message size allowed from peer.
	maxMessageSize = 1 * 1024 * 1024

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
	conn    *melody.Melody
	clients map[int64]bool
}

func NewWebSocket() WebSocket {
	server := melody.New()
	server.Upgrader.ReadBufferSize = maxMessageSize
	server.Upgrader.WriteBufferSize = maxMessageSize
	server.Config.PingPeriod = pingPeriod
	server.Config.PongWait = pongWait
	server.Config.WriteWait = writeWait
	server.Config.MaxMessageSize = maxMessageSize
	server.Config.MessageBufferSize = maxMessageSize
	//		ConcurrentMessageHandling bool

	return &webSocketImpl{
		conn:    server,
		clients: map[int64]bool{},
	}
}

func (h *webSocketImpl) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	return h.conn.HandleRequest(w, r)
}

func (h *webSocketImpl) Broadcast(eventName string, data interface{}) error {
	var wsData = EventMessage{Type: eventName, Payload: data}
	bytes, err := json.Marshal(wsData)
	if err != nil {
		utils.LogErrorf("WebSocket.Broadcast failed to marshal server payload %s", err.Error())

		return errors.New("failed to marshal server payload")
	}

	return h.conn.Broadcast(bytes)
}

func (h *webSocketImpl) Close() error {
	utils.LogDebugf("WebSocket.Close")
	return h.conn.Close()
}

func (h *webSocketImpl) Start(messageHandler func(message []byte)) {

	h.conn.HandleConnect(func(s *melody.Session) {
		utils.LogDebugf("WebSocket: New client connected")
		id := clientId.Add(1)

		h.clients[id] = true
		s.Set("id", id)

		//s.Write([]byte(fmt.Sprintf("client id %d connected", id)))
	})

	h.conn.HandleDisconnect(func(s *melody.Session) {
		if id, ok := s.Get("id"); ok {
			s.Write([]byte(fmt.Sprintf("WebSocket: client id %d disconnected", id)))

			h.clients[id.(int64)] = false
			h.conn.BroadcastOthers([]byte(fmt.Sprintf("dis %d", id)), s)
		} else {
			utils.LogDebug("WebSocket: client diconnected")
		}
	})

	h.conn.HandleError(func(s *melody.Session, err error) {
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
		if len(msg) > maxMessageSize {
			// Handle message too large error
			utils.LogError("WebSocket: message too large")
			s.CloseWithMsg([]byte(fmt.Sprintf("%d message too large", websocket.CloseMessageTooBig)))
			return
		}

		messageHandler(msg)
	})
}
