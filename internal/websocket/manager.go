package websocket

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/websocket/handler"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
)

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	webrtc types.WebRTCManager,
) *WebSocketManagerCtx {
	logger := log.With().Str("module", "websocket").Logger()

	return &WebSocketManagerCtx{
		logger:    logger,
		sessions:  sessions,
		desktop:   desktop,
		handler:   handler.New(sessions, desktop, capture, webrtc),
		handlers:  []types.HandlerFunction{},
	}
}

// Send pings to peer with this period. Must be less than pongWait.
const pingPeriod = 60 * time.Second

type WebSocketManagerCtx struct {
	logger    zerolog.Logger
	sessions  types.SessionManager
	desktop   types.DesktopManager
	handler   *handler.MessageHandlerCtx
	handlers  []types.HandlerFunction
}

func (ws *WebSocketManagerCtx) Start() {
	ws.sessions.OnCreated(func(session types.Session) {
		if err := ws.handler.SessionCreated(session); err != nil {
			ws.logger.Warn().Str("id", session.ID()).Err(err).Msg("session created with an error")
		} else {
			ws.logger.Debug().Str("id", session.ID()).Msg("session created")
		}
	})

	ws.sessions.OnDeleted(func(session types.Session) {
		if err := ws.handler.SessionDeleted(session); err != nil {
			ws.logger.Warn().Str("id", session.ID()).Err(err).Msg("session deleted with an error")
		} else {
			ws.logger.Debug().Str("id", session.ID()).Msg("session deleted")
		}
	})

	ws.sessions.OnConnected(func(session types.Session) {
		if err := ws.handler.SessionConnected(session); err != nil {
			ws.logger.Warn().Str("id", session.ID()).Err(err).Msg("session connected with an error")
		} else {
			ws.logger.Debug().Str("id", session.ID()).Msg("session connected")
		}
	})

	ws.sessions.OnDisconnected(func(session types.Session) {
		if err := ws.handler.SessionDisconnected(session); err != nil {
			ws.logger.Warn().Str("id", session.ID()).Err(err).Msg("session disconnected with an error")
		} else {
			ws.logger.Debug().Str("id", session.ID()).Msg("session disconnected")
		}
	})

	ws.sessions.OnProfileChanged(func(session types.Session) {
		if err := ws.handler.SessionProfileChanged(session); err != nil {
			ws.logger.Warn().Str("id", session.ID()).Err(err).Msg("session profile changed with an error")
		} else {
			ws.logger.Debug().Str("id", session.ID()).Msg("session profile changed")
		}
	})

	ws.sessions.OnStateChanged(func(session types.Session) {
		if err := ws.handler.SessionStateChanged(session); err != nil {
			ws.logger.Warn().Str("id", session.ID()).Err(err).Msg("session state changed with an error")
		} else {
			ws.logger.Debug().Str("id", session.ID()).Msg("session state changed")
		}
	})

	// TOOD: Throttle events.
	ws.desktop.OnCursorChanged(func(serial uint64) {
		cur := ws.desktop.GetCursorImage()
		uri, err := utils.GetCursorImageURI(cur)
		if err != nil {
			ws.logger.Warn().Err(err).Msg("could create cursor image")
			return
		}

		ws.sessions.Broadcast(message.CursorImage{
			Event:  event.CURSOR_IMAGE,
			Uri:    uri,
			Width:  cur.Width,
			Height: cur.Height,
			X:      cur.Xhot,
			Y:      cur.Yhot,
		}, nil)
	})

	ws.desktop.OnClipboardUpdated(func() {
		session := ws.sessions.GetHost()
		if session == nil || !session.CanAccessClipboard() {
			return
		}

		text, err := ws.desktop.ClipboardGetPlainText()
		if err != nil {
			ws.logger.Warn().Err(err).Msg("could not get clipboard content")
		}

		if err := session.Send(message.ClipboardData{
			Event: event.CLIPBOARD_UPDATED,
			Text:  text,
		}); err != nil {
			ws.logger.Warn().Err(err).Msg("could not sync clipboard")
		}
	})

	ws.fileChooserDialogEvents()
}

func (ws *WebSocketManagerCtx) Shutdown() error {
	return nil
}

func (ws *WebSocketManagerCtx) AddHandler(handler types.HandlerFunction) {
	ws.handlers = append(ws.handlers, handler)
}

func (ws *WebSocketManagerCtx) Upgrade(w http.ResponseWriter, r *http.Request, checkOrigin types.CheckOrigin) error {
	ws.logger.Debug().Msg("attempting to upgrade connection")

	upgrader := websocket.Upgrader{
		CheckOrigin: checkOrigin,
	}

	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Error().Err(err).Msg("failed to upgrade connection")
		return err
	}

	session, err := ws.sessions.Authenticate(r)
	if err != nil {
		ws.logger.Warn().Err(err).Msg("authentication failed")

		// TODO: Refactor, return error code.
		if err = connection.WriteJSON(
			message.SystemDisconnect{
				Event:   event.SYSTEM_DISCONNECT,
				Message: err.Error(),
			}); err != nil {
			ws.logger.Error().Err(err).Msg("failed to send disconnect")
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
			ws.logger.Error().Err(err).Msg("failed to send disconnect")
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
			ws.logger.Error().Err(err).Msg("failed to send disconnect")
		}

		return connection.Close()
	}

	session.SetWebSocketPeer(&WebSocketPeerCtx{
		session:    session,
		ws:         ws,
		connection: connection,
	})

	ws.logger.
		Debug().
		Str("session", session.ID()).
		Str("address", connection.RemoteAddr().String()).
		Msg("connection started")

	session.SetWebSocketConnected(true)

	defer func() {
		ws.logger.
			Debug().
			Str("session", session.ID()).
			Str("address", connection.RemoteAddr().String()).
			Msg("connection ended")

		session.SetWebSocketConnected(false)
	}()

	ws.handle(connection, session)
	return nil
}

func (ws *WebSocketManagerCtx) handle(connection *websocket.Conn, session types.Session) {
	bytes := make(chan []byte)
	cancel := make(chan struct{})

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	go func() {
		for {
			_, raw, err := connection.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					ws.logger.Warn().Err(err).Msg("read message error")
				} else {
					ws.logger.Debug().Err(err).Msg("read message error")
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
			ws.logger.Debug().
				Str("session", session.ID()).
				Str("address", connection.RemoteAddr().String()).
				Str("raw", string(raw)).
				Msg("received message from client")

			handled := ws.handler.Message(session, raw)
			for _, handler := range ws.handlers {
				if handled {
					break
				}

				handled = handler(session, raw)
			}

			if !handled {
				ws.logger.Warn().Msg("unhandled message")
			}
		case <-cancel:
			return
		case <-ticker.C:
			if err := connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				ws.logger.Error().Err(err).Msg("ping message has failed")
				return
			}
		}
	}
}
