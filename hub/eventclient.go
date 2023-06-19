package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	DeviceUpdated   = "deviceUpdated"
	ClientConnected = "connected"
)

var _eventhub *EventHub

type wsEvent struct {
	Name string
	Data interface{}
}

type WsHandler struct {
	path string
}

func NewWsHandler(path string) *WsHandler {
	return &WsHandler{
		path: path,
	}
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // for debug only ??
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			http.Error(w, reason.Error(), status)
		},
	}
)

type EventClient struct {
	hub *EventHub

	conn *websocket.Conn
	send chan []byte
}

func (c *EventClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.hub.broadcast <- message
	}
}

func (c *EventClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if strings.Compare(r.URL.Path, h.path) != 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("ws upgrade error:", err)
		return
	}

	client := &EventClient{hub: _eventhub, conn: conn, send: make(chan []byte)}
	client.hub.register <- client

	go client.readPump()
	go client.writePump()

	fmt.Printf("handler: client connected\n")
}

func Init(path string) *WsHandler {

	if _eventhub != nil {
		log.Fatal("Init - eventhub is already initialized")
		return nil
	}

	_eventhub = NewEventHub()
	go _eventhub.Run()

	handler := NewWsHandler(path)

	http.Handle(path, handler)
	return handler
}

func Broadcast(event string, data interface{}) {

	if _eventhub == nil {
		log.Fatal("Broadcast - eventhub is not initialized")
		return
	}

	var wsData = wsEvent{Name: event, Data: data}
	bytes, err := json.Marshal(wsData)
	if err != nil {
		panic(err)
	}
	_eventhub.Broadcast(bytes)
}
