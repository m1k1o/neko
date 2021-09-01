package websocket

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/websocket/handler"
)

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	webrtc types.WebRTCManager,
) *WebSocketManagerCtx {
	logger := log.With().Str("module", "websocket").Logger()

	return &WebSocketManagerCtx{
		logger:   logger,
		sessions: sessions,
		desktop:  desktop,
		handler:  handler.New(sessions, desktop, capture, webrtc),
		handlers: []types.WebSocketHandler{},
	}
}

// Send pings to peer with this period. Must be less than pongWait.
const pingPeriod = 10 * time.Second

type WebSocketManagerCtx struct {
	logger   zerolog.Logger
	sessions types.SessionManager
	desktop  types.DesktopManager
	handler  *handler.MessageHandlerCtx
	handlers []types.WebSocketHandler
}

func (manager *WebSocketManagerCtx) Start() {
	manager.sessions.OnCreated(func(session types.Session) {
		err := manager.handler.SessionCreated(session)
		manager.logger.Err(err).
			Str("session_id", session.ID()).
			Msg("session created")
	})

	manager.sessions.OnDeleted(func(session types.Session) {
		err := manager.handler.SessionDeleted(session)
		manager.logger.Err(err).
			Str("session_id", session.ID()).
			Msg("session deleted")
	})

	manager.sessions.OnConnected(func(session types.Session) {
		err := manager.handler.SessionConnected(session)
		manager.logger.Err(err).
			Str("session_id", session.ID()).
			Msg("session connected")
	})

	manager.sessions.OnDisconnected(func(session types.Session) {
		err := manager.handler.SessionDisconnected(session)
		manager.logger.Err(err).
			Str("session_id", session.ID()).
			Msg("session disconnected")
	})

	manager.sessions.OnProfileChanged(func(session types.Session) {
		err := manager.handler.SessionProfileChanged(session)
		manager.logger.Err(err).
			Str("session_id", session.ID()).
			Msg("session profile changed")
	})

	manager.sessions.OnStateChanged(func(session types.Session) {
		err := manager.handler.SessionStateChanged(session)
		manager.logger.Err(err).
			Str("session_id", session.ID()).
			Msg("session state changed")
	})

	manager.sessions.OnHostChanged(func(session types.Session) {
		payload := message.ControlHost{
			HasHost: session != nil,
		}

		if payload.HasHost {
			payload.HostID = session.ID()
		}

		manager.sessions.Broadcast(event.CONTROL_HOST, payload, nil)

		manager.logger.Debug().
			Bool("has_host", payload.HasHost).
			Str("host_id", payload.HostID).
			Msg("session host changed")
	})

	manager.desktop.OnClipboardUpdated(func() {
		session := manager.sessions.GetHost()
		if session == nil || !session.Profile().CanAccessClipboard {
			return
		}

		manager.logger.Debug().Msg("sync clipboard")

		data, err := manager.desktop.ClipboardGetText()
		if err != nil {
			manager.logger.Err(err).Msg("could not get clipboard content")
			return
		}

		session.Send(
			event.CLIPBOARD_UPDATED,
			message.ClipboardData{
				Text: data.Text,
				// TODO: Send HTML?
			})
	})

	manager.fileChooserDialogEvents()

	manager.logger.Info().Msg("websocket starting")
}

func (manager *WebSocketManagerCtx) Shutdown() error {
	manager.logger.Info().Msg("websocket shutdown")
	return nil
}

func (manager *WebSocketManagerCtx) AddHandler(handler types.WebSocketHandler) {
	manager.handlers = append(manager.handlers, handler)
}

func (manager *WebSocketManagerCtx) Upgrade(w http.ResponseWriter, r *http.Request, checkOrigin types.CheckOrigin) {
	manager.logger.Debug().
		Str("address", r.RemoteAddr).
		Str("agent", r.UserAgent()).
		Msg("attempting to upgrade connection")

	upgrader := websocket.Upgrader{
		CheckOrigin: checkOrigin,
	}

	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		manager.logger.Err(err).Msg("failed to upgrade connection")
		return
	}

	session, err := manager.sessions.Authenticate(r)
	if err != nil {
		manager.logger.Warn().Err(err).Msg("authentication failed")

		// TODO: Better handling...
		raw, err := json.Marshal(message.SystemDisconnect{
			Message: err.Error(),
		})

		if err != nil {
			manager.logger.Err(err).Msg("failed to create disconnect event")
		}

		err = connection.WriteJSON(
			types.WebSocketMessage{
				Event:   event.SYSTEM_DISCONNECT,
				Payload: raw,
			})

		if err != nil {
			manager.logger.Err(err).Msg("failed to send disconnect event")
		}

		if err := connection.Close(); err != nil {
			manager.logger.Warn().Err(err).Msg("connection closed with an error")
		}

		return
	}

	// use session id with defeault logger context
	logger := manager.logger.With().Str("session_id", session.ID()).Logger()

	// create new peer
	peer := &WebSocketPeerCtx{
		logger:     logger,
		session:    session,
		connection: connection,
	}

	if !session.Profile().CanConnect {
		logger.Warn().Msg("connection disabled")

		peer.Send(
			event.SYSTEM_DISCONNECT,
			message.SystemDisconnect{
				Message: "connection disabled",
			})

		peer.Destroy()
		return
	}

	if session.State().IsConnected {
		logger.Warn().Msg("already connected")

		if !manager.sessions.MercifulReconnect() {
			peer.Send(
				event.SYSTEM_DISCONNECT,
				message.SystemDisconnect{
					Message: "already connected",
				})

			peer.Destroy()
			return
		}

		logger.Info().Msg("replacing peer connection")
	}

	session.SetWebSocketPeer(peer)

	logger.Info().
		Str("address", connection.RemoteAddr().String()).
		Str("agent", r.UserAgent()).
		Msg("connection started")

	session.SetWebSocketConnected(peer, true)

	defer func() {
		logger.Info().
			Str("address", connection.RemoteAddr().String()).
			Str("agent", r.UserAgent()).
			Msg("connection ended")

		session.SetWebSocketConnected(peer, false)
	}()

	manager.handle(connection, session)
}

func (manager *WebSocketManagerCtx) handle(connection *websocket.Conn, session types.Session) {
	// add session id to logger context
	logger := manager.logger.With().Str("session_id", session.ID()).Logger()

	bytes := make(chan []byte)
	cancel := make(chan struct{})

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	go func() {
		for {
			_, raw, err := connection.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Warn().Err(err).Msg("read message error")
				} else {
					logger.Debug().Err(err).Msg("read message error")
				}

				close(cancel)
				break
			}

			bytes <- raw
		}
	}()

	for {
		select {
		case raw := <-bytes:
			data := types.WebSocketMessage{}
			if err := json.Unmarshal(raw, &data); err != nil {
				logger.Err(err).Msg("message unmarshalling has failed")
				break
			}

			logger.Debug().
				Str("address", connection.RemoteAddr().String()).
				Str("event", data.Event).
				Str("payload", string(data.Payload)).
				Msg("received message from client")

			handled := manager.handler.Message(session, data)
			for _, handler := range manager.handlers {
				if handled {
					break
				}

				handled = handler(session, data)
			}

			if !handled {
				logger.Warn().Str("event", data.Event).Msg("unhandled message")
			}
		case <-cancel:
			return
		case <-ticker.C:
			if err := connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Err(err).Msg("ping message has failed")
				return
			}
		}
	}
}
