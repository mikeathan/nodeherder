package hub

import (
	"encoding/json"
	"fmt"
)

type device struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

type EventHub struct {
	clients    map[*EventClient]bool
	broadcast  chan []byte
	register   chan *EventClient
	unregister chan *EventClient
}

func NewEventHub() *EventHub {
	return &EventHub{
		clients:    map[*EventClient]bool{},
		broadcast:  make(chan []byte),
		register:   make(chan *EventClient),
		unregister: make(chan *EventClient),
	}
}

func (h *EventHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			fmt.Println("hub: client registered")
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				fmt.Println("hub: client unregistered")
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					fmt.Println("broadcast failed, client closed")
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *EventHub) Broadcast(message []byte) {
	h.broadcast <- message
	// for client := range h.clients {
	// 	client.send <- message
	// }
}
