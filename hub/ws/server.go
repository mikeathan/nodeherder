package ws

import (
	"fmt"
)

type Server struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewServer() *Server {
	return &Server{
		clients:    map[*Client]bool{},
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Server) Run() {
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

func (h *Server) Broadcast(message []byte) {
	h.broadcast <- message
}
