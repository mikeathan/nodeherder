package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var _eventhub *eventHub

type wsHandler struct {
	path string
}

// const (
// 	// Time allowed to write a message to the peer.
// 	writeWait = 10 * time.Second

// 	// Time allowed to read the next pong message from the peer.
// 	pongWait = 60 * time.Second

// 	// Send pings to peer with this period. Must be less than pongWait.
// 	pingPeriod = (pongWait * 9) / 10

// 	// Maximum message size allowed from peer.
// 	maxMessageSize = 512
// )

// var (
// 	newline = []byte{'\n'}
// 	space   = []byte{' '}
// )

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

// type Client struct {
// 	hub *eventHub

// 	// The websocket connection.
// 	conn *websocket.Conn

// 	// Buffered channel of outbound messages.
// 	send chan []byte
// }

// func (c *Client) readPump() {
// 	defer func() {
// 		c.hub.unregister <- c
// 		c.conn.Close()
// 	}()
// 	c.conn.SetReadLimit(maxMessageSize)
// 	c.conn.SetReadDeadline(time.Now().Add(pongWait))
// 	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
// 	for {
// 		_, message, err := c.conn.ReadMessage()
// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				log.Printf("error: %v", err)
// 			}
// 			break
// 		}
// 		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
// 		c.hub.broadcast <- message
// 	}
// }

// func (c *Client) writePump() {
// 	ticker := time.NewTicker(pingPeriod)
// 	defer func() {
// 		ticker.Stop()
// 		c.conn.Close()
// 	}()
// 	for {
// 		select {
// 		case message, ok := <-c.send:
// 			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
// 			if !ok {
// 				// The hub closed the channel.
// 				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
// 				return
// 			}

// 			w, err := c.conn.NextWriter(websocket.TextMessage)
// 			if err != nil {
// 				return
// 			}
// 			w.Write(message)

// 			// Add queued chat messages to the current websocket message.
// 			n := len(c.send)
// 			for i := 0; i < n; i++ {
// 				w.Write(newline)
// 				w.Write(<-c.send)
// 			}

// 			if err := w.Close(); err != nil {
// 				return
// 			}
// 		case <-ticker.C:
// 			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
// 			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
// 				return
// 			}
// 		}
// 	}
// }

// func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
// 	client.hub.register <- client

// 	go client.writePump()
// 	go client.readPump()
// }

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if strings.Compare(r.URL.Path, h.path) != 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	connection, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("ws upgrade error:", err)
		return
	}

	_eventhub.clients[connection] = true
	fmt.Printf("client connected\n")

	for {
		mt, message, err := connection.ReadMessage()
		if err != nil || mt == websocket.CloseMessage {
			break
		}

		fmt.Printf("message from client %s \n", string(message))
	}

	fmt.Printf("client disconnected\n")
	delete(_eventhub.clients, connection)
	connection.Close()
}

func Init(path string) {

	if _eventhub != nil {
		log.Fatal("Init - eventhub is already initialized")
		return
	}

	_eventhub = newEventHub()

	handler := &wsHandler{path: path}

	http.Handle(path, handler)
}

func Broadcast(event interface{}) {

	if _eventhub == nil {
		log.Fatal("Broadcast - eventhub is not initialized")
		return
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	_eventhub.Broadcast(bytes)
}
