package websocket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/config"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
)

func New(sessions types.SessionManager, remote types.RemoteManager, broadcast types.BroadcastManager, webrtc types.WebRTCManager, conf *config.WebSocket) *WebSocketHandler {
	logger := log.With().Str("module", "websocket").Logger()

	return &WebSocketHandler{
		logger:    logger,
		conf:      conf,
		sessions:  sessions,
		remote:    remote,
		upgrader:  websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		handler: &MessageHandler{
			logger:    logger.With().Str("subsystem", "handler").Logger(),
			remote:    remote,
			broadcast: broadcast,
			sessions:  sessions,
			webrtc:    webrtc,
			banned:    make(map[string]bool),
			locked:    false,
		},
	}
}

// Send pings to peer with this period. Must be less than pongWait.
const pingPeriod = 60 * time.Second

type WebSocketHandler struct {
	logger    zerolog.Logger
	upgrader  websocket.Upgrader
	sessions  types.SessionManager
	remote    types.RemoteManager
	conf      *config.WebSocket
	handler   *MessageHandler
	shutdown  chan bool
}

func (ws *WebSocketHandler) Start() {
	ws.sessions.OnCreated(func(session types.Session) {
		if err := ws.handler.SessionCreated(session); err != nil {
			ws.logger.Warn().Str("id", session.ID()).Err(err).Msg("session created with and error")
		} else {
			ws.logger.Debug().Str("id", session.ID()).Msg("session created")
		}
	})

	ws.sessions.OnConnected(func(session types.Session) {
		if err := ws.handler.SessionConnected(session); err != nil {
			ws.logger.Warn().Str("id", session.ID()).Err(err).Msg("session connected with and error")
		} else {
			ws.logger.Debug().Str("id", session.ID()).Msg("session connected")
		}
	})

	ws.sessions.OnDestroy(func(id string) {
		if err := ws.handler.SessionDestroyed(id); err != nil {
			ws.logger.Warn().Str("id", id).Err(err).Msg("session destroyed with and error")
		} else {
			ws.logger.Debug().Str("id", id).Msg("session destroyed")
		}
	})

	go func() {
		defer func() {
			ws.logger.Info().Msg("shutdown")
		}()

		current := ws.remote.ReadClipboard()

		for {
			select {
			case <-ws.shutdown:
				return
			default:
				if ws.sessions.HasHost() {
					text := ws.remote.ReadClipboard()
					if text != current {
						session := ws.sessions.GetHost()
						if session != nil {
							if err := session.Send(message.Clipboard{
								Event: event.CONTROL_CLIPBOARD,
								Text:  text,
							}); err != nil {
								ws.logger.Warn().Err(err).Msg("could not sync clipboard")
							}
						}
						current = text
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (ws *WebSocketHandler) Shutdown() error {
	ws.shutdown <- true
	return nil
}

func (ws *WebSocketHandler) Upgrade(w http.ResponseWriter, r *http.Request) error {
	ws.logger.Debug().Msg("attempting to upgrade connection")

	connection, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Error().Err(err).Msg("failed to upgrade connection")
		return err
	}

	id, ip, admin, err := ws.authenticate(r)
	if err != nil {
		ws.logger.Warn().Err(err).Msg("authentication failed")

		if err = connection.WriteJSON(message.Disconnect{
			Event:   event.SYSTEM_DISCONNECT,
			Message: "invalid_password",
		}); err != nil {
			ws.logger.Error().Err(err).Msg("failed to send disconnect")
		}

		if err = connection.Close(); err != nil {
			return err
		}
		return nil
	}

	socket := &WebSocket{
		id:         id,
		ws:         ws,
		address:    ip,
		connection: connection,
	}

	ok, reason, err := ws.handler.Connected(id, socket)
	if err != nil {
		ws.logger.Error().Err(err).Msg("connection failed")
		return err
	}

	if !ok {
		if err = connection.WriteJSON(message.Disconnect{
			Event:   event.SYSTEM_DISCONNECT,
			Message: reason,
		}); err != nil {
			ws.logger.Error().Err(err).Msg("failed to send disconnect")
		}

		if err = connection.Close(); err != nil {
			return err
		}

		return nil
	}

	ws.sessions.New(id, admin, socket)

	ws.logger.
		Debug().
		Str("session", id).
		Str("address", connection.RemoteAddr().String()).
		Msg("new connection created")

	defer func() {
		ws.logger.
			Debug().
			Str("session", id).
			Str("address", connection.RemoteAddr().String()).
			Msg("session ended")
	}()

	ws.handle(connection, id)
	return nil
}

func (ws *WebSocketHandler) authenticate(r *http.Request) (string, string, bool, error) {
	ip := r.RemoteAddr

	if ws.conf.Proxy {
		ip = utils.ReadUserIP(r)
	}

	id, err := utils.NewUID(32)
	if err != nil {
		return "", ip, false, err
	}

	passwords, ok := r.URL.Query()["password"]
	if !ok || len(passwords[0]) < 1 {
		return "", ip, false, fmt.Errorf("no password provided")
	}

	if passwords[0] == ws.conf.AdminPassword {
		return id, ip, true, nil
	}

	if passwords[0] == ws.conf.Password {
		return id, ip, false, nil
	}

	return "", ip, false, fmt.Errorf("invalid password: %s", passwords[0])
}

func (ws *WebSocketHandler) handle(connection *websocket.Conn, id string) {
	bytes := make(chan []byte)
	cancel := make(chan struct{})
	ticker := time.NewTicker(pingPeriod)

	go func() {
		defer func() {
			ticker.Stop()
			ws.logger.Debug().Str("address", connection.RemoteAddr().String()).Msg("handle socket ending")
			if err := ws.handler.Disconnected(id); err != nil {
				ws.logger.Warn().Err(err).Msg("socket disconnected with error")
			}
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
				Str("session", id).
				Str("address", connection.RemoteAddr().String()).
				Str("raw", string(raw)).
				Msg("received message from client")
			if err := ws.handler.Message(id, raw); err != nil {
				ws.logger.Error().Err(err).Msg("message handler has failed")
			}
		case <-cancel:
			return
		case <-ticker.C:
			if err := connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
