package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
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

func (peer *WebSocketPeerCtx) Destroy() {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return
	}

	if err := peer.Send(
		message.SystemDisconnect{
			Event:   event.SYSTEM_DISCONNECT,
			Message: "connection destroyed",
		}); err != nil {
		peer.logger.Warn().Err(err).Msg("failed to send disconnect event")
	}

	if err := peer.connection.Close(); err != nil {
		peer.logger.Warn().Err(err).Msg("peer connection destroyed with an error")
	} else {
		peer.logger.Info().Msg("peer connection destroyed")
	}

	peer.connection = nil
}
