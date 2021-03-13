package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"

	"demodesk/neko/internal/types"
)

type WebSocketPeerCtx struct {
	mu         sync.Mutex
	session    types.Session
	manager    *WebSocketManagerCtx
	connection *websocket.Conn
}

func (peer *WebSocketPeerCtx) Send(v interface{}) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return nil
	}

	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}

	peer.manager.logger.Debug().
		Str("session_id", peer.session.ID()).
		Str("address", peer.connection.RemoteAddr().String()).
		Str("raw", string(raw)).
		Msg("sending message to client")

	return peer.connection.WriteMessage(websocket.TextMessage, raw)
}

func (peer *WebSocketPeerCtx) Destroy() error {
	if peer.connection == nil {
		return nil
	}

	return peer.connection.Close()
}
