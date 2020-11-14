package websocket

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/websocket/handler"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"

	"demodesk/neko/internal/types"
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
		upgrader:  websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		handler: handler.New(sessions, desktop, capture, webrtc),
	}
}

// Send pings to peer with this period. Must be less than pongWait.
const pingPeriod = 60 * time.Second

type WebSocketManagerCtx struct {
	logger    zerolog.Logger
	upgrader  websocket.Upgrader
	sessions  types.SessionManager
	desktop   types.DesktopManager
	handler   *handler.MessageHandlerCtx
	shutdown  chan bool
}

func (ws *WebSocketManagerCtx) Start() {
	ws.sessions.OnCreated(func(session types.Session) {
		if err := ws.handler.SessionCreated(session); err != nil {
			ws.logger.Warn().Str("id", session.ID()).Err(err).Msg("session created with an error")
		} else {
			ws.logger.Debug().Str("id", session.ID()).Msg("session created")
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

	go func() {
		ws.logger.Info().Msg("clipboard loop started")

		defer func() {
			ws.logger.Info().Msg("clipboard loop stopped")
		}()

		current := ws.desktop.ReadClipboard()

		for {
			select {
			case <-ws.shutdown:
				return
			default:
				session := ws.sessions.GetHost()
				if session != nil {
					break
				}

				text := ws.desktop.ReadClipboard()
				if text == current {
					break
				}

				// TODO: Refactor
				if err := session.Send(message.Clipboard{
					Event: event.CONTROL_CLIPBOARD,
					Text:  text,
				}); err != nil {
					ws.logger.Warn().Err(err).Msg("could not sync clipboard")
				}

				current = text
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (ws *WebSocketManagerCtx) Shutdown() error {
	ws.shutdown <- true
	return nil
}

func (ws *WebSocketManagerCtx) Upgrade(w http.ResponseWriter, r *http.Request) error {
	ws.logger.Debug().Msg("attempting to upgrade connection")

	connection, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Error().Err(err).Msg("failed to upgrade connection")
		return err
	}

	session, err := ws.sessions.Authenticate(r)
	if err != nil {
		ws.logger.Warn().Err(err).Msg("authentication failed")

		// TODO: Refactor
		if err = connection.WriteJSON(message.Disconnect{
			Event:   event.SYSTEM_DISCONNECT,
			Message: "authentication failed",
		}); err != nil {
			ws.logger.Error().Err(err).Msg("failed to send disconnect")
		}

		return connection.Close()
	}

	// TODO: Refactor.
	ip := r.RemoteAddr
	// if allow poxy {
	// 	ip = utils.ReadUserIP(r)
	// }

	socket := &WebSocketCtx{
		session:    session,
		ws:         ws,
		address:    ip,
		connection: connection,
	}

	ok, reason := ws.handler.Connected(session, socket)
	if !ok {
		// TODO: Refactor
		if err = connection.WriteJSON(message.Disconnect{
			Event:   event.SYSTEM_DISCONNECT,
			Message: reason,
		}); err != nil {
			ws.logger.Error().Err(err).Msg("failed to send disconnect")
		}

		return connection.Close()
	}

	session.SetSocket(socket)

	ws.logger.
		Debug().
		Str("session", session.ID()).
		Str("address", connection.RemoteAddr().String()).
		Msg("connection started")

	defer func() {
		ws.logger.
			Debug().
			Str("session", session.ID()).
			Str("address", connection.RemoteAddr().String()).
			Msg("connection ended")
	}()

	ws.handle(connection, session)
	return nil
}

func (ws *WebSocketManagerCtx) handle(connection *websocket.Conn, session types.Session) {
	bytes := make(chan []byte)
	cancel := make(chan struct{})
	ticker := time.NewTicker(pingPeriod)

	go func() {
		defer func() {
			ticker.Stop()
			ws.logger.Debug().Str("address", connection.RemoteAddr().String()).Msg("handle socket ending")
			session.SetConnected(false)
		}()

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

			if err := ws.handler.Message(session, raw); err != nil {
				ws.logger.Error().Err(err).Msg("message handler has failed")
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
