package main

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
			for client := range h.hosts {
				h.unregister <- client
			}
			for client := range h.clients {
				h.unregister <- client
			}
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case raw := <-h.broadcast:
			for client := range h.hosts {
				select {
				case client.send <- raw:
				default:
					h.unregister <- client
				}
			}
		}
	}
}
