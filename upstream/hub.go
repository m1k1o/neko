package main

import "log"

type Hub struct {
	hosts      map[*Client]bool
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	close      chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		hosts:      map[*Client]bool{},
		clients:    map[*Client]bool{},
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		close:      make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.close:
			log.Println("Closing all connections")
			for client := range h.hosts {
				h.unregister <- client
			}
			for client := range h.clients {
				h.unregister <- client
			}
		case client := <-h.register:
			switch client.connectionType {
			case ClientConn:
				h.clients[client] = true
			case HostConn:
				h.hosts[client] = true
			}
			log.Printf("New connection: %s", client.connectionType)
		case client := <-h.unregister:
			log.Printf("Disconnecting %s", client.connectionType)
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
			}
		case raw := <-h.broadcast:
			for client := range h.hosts {
				client.send <- raw
			}
		}
	}
}
