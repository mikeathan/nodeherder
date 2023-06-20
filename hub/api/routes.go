package api

import (
	"fmt"
	"log"
	"net/http"
	"node-herder/hub/ws"
	"node-herder/node"
	"strings"

	"github.com/gorilla/websocket"
)

type WsHandler struct {
	path string
}

func NewHandler(path string) *WsHandler {
	return &WsHandler{
		path: path,
	}
}

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

func (h *WsHandler) UseWebSockets(store node.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Compare(r.URL.Path, h.path) != 0 {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		conn, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("ws upgrade error:", err)
			return
		}

		client := ws.RegisterConnection(nil, conn)

		devices := store.ListAll()
		client.Broadcast(ws.DeviceUpdated, devices)
		fmt.Printf("handler: client connected\n")
	}
}
