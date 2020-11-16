package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"

	"demodesk/neko/internal/types"
)

type WebSocketCtx struct {
	session    types.Session
	ws         *WebSocketManagerCtx
	connection *websocket.Conn
	mu         sync.Mutex
}

func (socket *WebSocketCtx) Send(v interface{}) error {
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
		Str("session", socket.session.ID()).
		Str("address", socket.connection.RemoteAddr().String()).
		Str("raw", string(raw)).
		Msg("sending message to client")

	return socket.connection.WriteMessage(websocket.TextMessage, raw)
}

func (socket *WebSocketCtx) Destroy() error {
	if socket.connection == nil {
		return nil
	}

	return socket.connection.Close()
}
