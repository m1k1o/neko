package websocket

import (
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
		handlers: []types.HandlerFunction{},
	}
}

// Send pings to peer with this period. Must be less than pongWait.
const pingPeriod = 10 * time.Second

type WebSocketManagerCtx struct {
	logger   zerolog.Logger
	sessions types.SessionManager
	desktop  types.DesktopManager
	handler  *handler.MessageHandlerCtx
	handlers []types.HandlerFunction
	shutdown chan bool
}

func (manager *WebSocketManagerCtx) Start() {
	manager.sessions.OnCreated(func(session types.Session) {
		if err := manager.handler.SessionCreated(session); err != nil {
			manager.logger.Warn().Str("id", session.ID()).Err(err).Msg("session created with an error")
		} else {
			manager.logger.Debug().Str("id", session.ID()).Msg("session created")
		}
	})

	manager.sessions.OnDeleted(func(session types.Session) {
		if err := manager.handler.SessionDeleted(session); err != nil {
			manager.logger.Warn().Str("id", session.ID()).Err(err).Msg("session deleted with an error")
		} else {
			manager.logger.Debug().Str("id", session.ID()).Msg("session deleted")
		}
	})

	manager.sessions.OnConnected(func(session types.Session) {
		if err := manager.handler.SessionConnected(session); err != nil {
			manager.logger.Warn().Str("id", session.ID()).Err(err).Msg("session connected with an error")
		} else {
			manager.logger.Debug().Str("id", session.ID()).Msg("session connected")
		}
	})

	manager.sessions.OnDisconnected(func(session types.Session) {
		if err := manager.handler.SessionDisconnected(session); err != nil {
			manager.logger.Warn().Str("id", session.ID()).Err(err).Msg("session disconnected with an error")
		} else {
			manager.logger.Debug().Str("id", session.ID()).Msg("session disconnected")
		}
	})

	manager.sessions.OnProfileChanged(func(session types.Session) {
		if err := manager.handler.SessionProfileChanged(session); err != nil {
			manager.logger.Warn().Str("id", session.ID()).Err(err).Msg("session profile changed with an error")
		} else {
			manager.logger.Debug().Str("id", session.ID()).Msg("session profile changed")
		}
	})

	manager.sessions.OnStateChanged(func(session types.Session) {
		if err := manager.handler.SessionStateChanged(session); err != nil {
			manager.logger.Warn().Str("id", session.ID()).Err(err).Msg("session state changed with an error")
		} else {
			manager.logger.Debug().Str("id", session.ID()).Msg("session state changed")
		}
	})

	manager.desktop.OnClipboardUpdated(func() {
		session := manager.sessions.GetHost()
		if session == nil || !session.CanAccessClipboard() {
			return
		}

		data, err := manager.desktop.ClipboardGetText()
		if err != nil {
			manager.logger.Warn().Err(err).Msg("could not get clipboard content")
			return
		}

		if err := session.Send(message.ClipboardData{
			Event: event.CLIPBOARD_UPDATED,
			Text:  data.Text,
			// TODO: Send HTML?
		}); err != nil {
			manager.logger.Warn().Err(err).Msg("could not sync clipboard")
		}
	})

	manager.fileChooserDialogEvents()
}

func (manager *WebSocketManagerCtx) Shutdown() error {
	manager.shutdown <- true
	return nil
}

func (manager *WebSocketManagerCtx) AddHandler(handler types.HandlerFunction) {
	manager.handlers = append(manager.handlers, handler)
}

func (manager *WebSocketManagerCtx) Upgrade(w http.ResponseWriter, r *http.Request, checkOrigin types.CheckOrigin) error {
	manager.logger.Debug().Msg("attempting to upgrade connection")

	upgrader := websocket.Upgrader{
		CheckOrigin: checkOrigin,
	}

	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		manager.logger.Error().Err(err).Msg("failed to upgrade connection")
		return err
	}

	session, err := manager.sessions.AuthenticateRequest(r)
	if err != nil {
		manager.logger.Warn().Err(err).Msg("authentication failed")

		// TODO: Refactor, return error code.
		if err = connection.WriteJSON(
			message.SystemDisconnect{
				Event:   event.SYSTEM_DISCONNECT,
				Message: err.Error(),
			}); err != nil {
			manager.logger.Error().Err(err).Msg("failed to send disconnect")
		}

		return connection.Close()
	}

	if !session.CanConnect() {
		// TODO: Refactor, return error code.
		if err = connection.WriteJSON(
			message.SystemDisconnect{
				Event:   event.SYSTEM_DISCONNECT,
				Message: "connection disabled",
			}); err != nil {
			manager.logger.Error().Err(err).Msg("failed to send disconnect")
		}

		return connection.Close()
	}

	if session.IsConnected() {
		// TODO: Refactor, return error code.
		if err = connection.WriteJSON(
			message.SystemDisconnect{
				Event:   event.SYSTEM_DISCONNECT,
				Message: "already connected",
			}); err != nil {
			manager.logger.Error().Err(err).Msg("failed to send disconnect")
		}

		return connection.Close()
	}

	session.SetWebSocketPeer(&WebSocketPeerCtx{
		session:    session,
		manager:    manager,
		connection: connection,
	})

	manager.logger.
		Debug().
		Str("session", session.ID()).
		Str("address", connection.RemoteAddr().String()).
		Msg("connection started")

	session.SetWebSocketConnected(true)

	defer func() {
		manager.logger.
			Debug().
			Str("session", session.ID()).
			Str("address", connection.RemoteAddr().String()).
			Msg("connection ended")

		session.SetWebSocketConnected(false)
	}()

	manager.handle(connection, session)
	return nil
}

func (manager *WebSocketManagerCtx) handle(connection *websocket.Conn, session types.Session) {
	bytes := make(chan []byte)
	cancel := make(chan struct{})

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	go func() {
		for {
			_, raw, err := connection.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					manager.logger.Warn().Err(err).Msg("read message error")
				} else {
					manager.logger.Debug().Err(err).Msg("read message error")
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
			manager.logger.Debug().
				Str("session", session.ID()).
				Str("address", connection.RemoteAddr().String()).
				Str("raw", string(raw)).
				Msg("received message from client")

			handled := manager.handler.Message(session, raw)
			for _, handler := range manager.handlers {
				if handled {
					break
				}

				handled = handler(session, raw)
			}

			if !handled {
				manager.logger.Warn().Msg("unhandled message")
			}
		case <-cancel:
			return
		case <-ticker.C:
			if err := connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				manager.logger.Error().Err(err).Msg("ping message has failed")
				return
			}
		}
	}
}
