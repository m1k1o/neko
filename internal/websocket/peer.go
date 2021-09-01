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

func (peer *WebSocketPeerCtx) Send(event string, payload interface{}) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		peer.logger.Error().Err(err).Str("event", event).Msg("message marshalling has failed")
		return
	}

	err = peer.connection.WriteJSON(types.WebSocketMessage{
		Event:   event,
		Payload: raw,
	})

	if err != nil {
		peer.logger.Error().Err(err).Str("event", event).Msg("send message error")
		return
	}

	peer.logger.Debug().
		Str("address", peer.connection.RemoteAddr().String()).
		Str("event", event).
		Str("payload", string(raw)).
		Msg("sending message to client")
}

func (peer *WebSocketPeerCtx) Destroy() {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return
	}

	peer.Send(
		event.SYSTEM_DISCONNECT,
		message.SystemDisconnect{
			Message: "connection destroyed",
		})

	if err := peer.connection.Close(); err != nil {
		peer.logger.Warn().Err(err).Msg("peer connection destroyed with an error")
	} else {
		peer.logger.Info().Msg("peer connection destroyed")
	}

	peer.connection = nil
}
