package hub

import (
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
		log.Fatal("eventhub is initialized")
		return
	}

	_eventhub = newEventHub()
	handler := &wsHandler{path: path}

	http.Handle("/"+path, handler)
}

func Broadcast(event interface{}) {

	if _eventhub == nil {
		log.Fatal("eventhub is not initialized")
		return
	}

	_eventhub.Broadcast(nil)
}
