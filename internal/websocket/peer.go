package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"

	"demodesk/neko/internal/types"
)

type WebSocketPeerCtx struct {
	session    types.Session
	ws         *WebSocketManagerCtx
	connection *websocket.Conn
	mu         sync.Mutex
}

func (websocket_peer *WebSocketPeerCtx) Send(v interface{}) error {
	websocket_peer.mu.Lock()
	defer websocket_peer.mu.Unlock()

	if websocket_peer.connection == nil {
		return nil
	}

	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}

	websocket_peer.ws.logger.Debug().
		Str("session", websocket_peer.session.ID()).
		Str("address", websocket_peer.connection.RemoteAddr().String()).
		Str("raw", string(raw)).
		Msg("sending message to client")

	return websocket_peer.connection.WriteMessage(websocket.TextMessage, raw)
}

func (websocket_peer *WebSocketPeerCtx) Destroy() error {
	if websocket_peer.connection == nil {
		return nil
	}

	return websocket_peer.connection.Close()
}
