package hub

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type device struct {
	Name    string      `json:"name"`
	Payload interface{} `json:"payload"`
}
type eventHub struct {
	clients map[*websocket.Conn]bool
}

func newEventHub() *eventHub {
	return &eventHub{
		clients: map[*websocket.Conn]bool{},
	}
}

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // for debug only ??
	}
)

func (h *eventHub) Broadcast(message []byte) {
	for conn := range h.clients {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}
