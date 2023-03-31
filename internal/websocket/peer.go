package websocket

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
	"github.com/demodesk/neko/pkg/utils"
)

type WebSocketPeerCtx struct {
	mu         sync.Mutex
	logger     zerolog.Logger
	connection *websocket.Conn
}

func newPeer(connection *websocket.Conn) *WebSocketPeerCtx {
	logger := log.With().
		Str("module", "websocket").
		Str("submodule", "peer").
		Logger()

	return &WebSocketPeerCtx{
		logger:     logger,
		connection: connection,
	}
}

func (peer *WebSocketPeerCtx) setSessionID(sessionId string) {
	peer.logger = peer.logger.With().Str("session_id", sessionId).Logger()
}

func (peer *WebSocketPeerCtx) Send(event string, payload any) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		peer.logger.Err(err).Str("event", event).Msg("message marshalling has failed")
		return
	}

	err = peer.connection.WriteJSON(types.WebSocketMessage{
		Event:   event,
		Payload: raw,
	})

	if err != nil {
		peer.logger.Err(err).Str("event", event).Msg("send message error")
		return
	}

	// log events if not ignored
	if ok, _ := utils.ArrayIn(event, nologEvents); !ok {
		if len(raw) > maxPayloadLogLength {
			raw = []byte("<truncated>")
		}

		peer.logger.Debug().
			Str("address", peer.connection.RemoteAddr().String()).
			Str("event", event).
			Str("payload", string(raw)).
			Msg("sending message to client")
	}
}

func (peer *WebSocketPeerCtx) Ping() error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return errors.New("peer connection not found")
	}

	// application level heartbeat
	if err := peer.connection.WriteJSON(types.WebSocketMessage{
		Event: event.SYSTEM_HEARTBEAT,
	}); err != nil {
		return err
	}

	return peer.connection.WriteMessage(websocket.PingMessage, nil)
}

func (peer *WebSocketPeerCtx) Destroy(reason string) {
	peer.Send(
		event.SYSTEM_DISCONNECT,
		message.SystemDisconnect{
			Message: reason,
		})

	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection != nil {
		err := peer.connection.Close()
		peer.logger.Err(err).Msg("peer connection destroyed")
		peer.connection = nil
	}
}
