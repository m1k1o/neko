package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"

	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
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

	var errs []error

	// send disconnect
	err := peer.Send(
		message.SystemDisconnect{
			Event:   event.SYSTEM_DISCONNECT,
			Message: "connection destroyed",
		})
	errs = append(errs, err)

	// close connection
	err = peer.connection.Close()
	errs = append(errs, err)

	return utils.ErrorsJoin(errs)
}
