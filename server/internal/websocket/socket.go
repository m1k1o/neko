package websocket

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	id         string
	address    string
	ws         *WebSocketHandler
	connection *websocket.Conn
	mu         sync.Mutex
}

func (socket *WebSocket) Address() string {
	//remote := socket.connection.RemoteAddr()
	address := strings.SplitN(socket.address, ":", -1)
	if len(address[0]) < 1 {
		return socket.address
	}
	return address[0]
}

func (socket *WebSocket) Send(v interface{}) error {
	socket.mu.Lock()
	defer socket.mu.Unlock()
	if socket.connection == nil {
		return nil
	}

	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}

	socket.ws.logger.Debug().
		Str("session", socket.id).
		Str("address", socket.connection.RemoteAddr().String()).
		Str("raw", string(raw)).
		Msg("sending message to client")

	return socket.connection.WriteMessage(websocket.TextMessage, raw)
}

func (socket *WebSocket) Destroy() error {
	if socket.connection == nil {
		return nil
	}

	return socket.connection.Close()
}
