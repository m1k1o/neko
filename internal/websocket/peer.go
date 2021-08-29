package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
)

type WebSocketPeerCtx struct {
	mu         sync.Mutex
	logger     zerolog.Logger
	session    types.Session
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

	peer.logger.Debug().
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
