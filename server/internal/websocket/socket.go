package websocket

import (
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	id         string
	connection *websocket.Conn
	mu         sync.Mutex
}

func (socket *WebSocket) Address() *string {
	remote := socket.connection.RemoteAddr()
	address := strings.SplitN(remote.String(), ":", -1)
	if len(address[0]) < 1 {
		return nil
	}
	return &address[0]
}

func (socket *WebSocket) Send(v interface{}) error {
	socket.mu.Lock()
	defer socket.mu.Unlock()
	if socket.connection == nil {
		return nil
	}

	return socket.connection.WriteJSON(v)
}

func (socket *WebSocket) Destroy() error {
	if socket.connection == nil {
		return nil
	}

	return socket.connection.Close()
}
