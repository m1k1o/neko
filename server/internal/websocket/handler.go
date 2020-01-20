package websocket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"n.eko.moe/neko/internal/config"
	"n.eko.moe/neko/internal/event"
	"n.eko.moe/neko/internal/message"
	"n.eko.moe/neko/internal/session"
	"n.eko.moe/neko/internal/utils"
	"n.eko.moe/neko/internal/webrtc"
)

func New(sessions *session.SessionManager, webrtc *webrtc.WebRTCManager, conf *config.WebSocket) *WebSocketHandler {
	logger := log.With().Str("module", "websocket").Logger()

	return &WebSocketHandler{
		logger:   logger,
		conf:     conf,
		sessions: sessions,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		handler: &MessageHandler{
			logger:   logger.With().Str("subsystem", "handler").Logger(),
			sessions: sessions,
			webrtc:   webrtc,
			banned:   make(map[string]bool),
			locked:   false,
		},
	}
}

// Send pings to peer with this period. Must be less than pongWait.
const pingPeriod = 60 * time.Second

type WebSocketHandler struct {
	logger   zerolog.Logger
	upgrader websocket.Upgrader
	handler  *MessageHandler
	conf     *config.WebSocket
	sessions *session.SessionManager
	shutdown chan bool
}

func (ws *WebSocketHandler) Start() error {

	go func() {
		defer func() {
			ws.logger.Info().Msg("shutdown")
		}()

		for {
			select {
			case <-ws.shutdown:
				return
			}
		}
	}()

	ws.sessions.OnCreated(func(id string, session *session.Session) {
		if err := ws.handler.SessionCreated(id, session); err != nil {
			ws.logger.Warn().Str("id", id).Err(err).Msg("session created with and error")
		} else {
			ws.logger.Debug().Str("id", id).Msg("session created")
		}
	})

	ws.sessions.OnConnected(func(id string, session *session.Session) {
		if err := ws.handler.SessionConnected(id, session); err != nil {
			ws.logger.Warn().Str("id", id).Err(err).Msg("session connected with and error")
		} else {
			ws.logger.Debug().Str("id", id).Msg("session connected")
		}
	})

	ws.sessions.OnDestroy(func(id string) {
		if err := ws.handler.SessionDestroyed(id); err != nil {
			ws.logger.Warn().Str("id", id).Err(err).Msg("session destroyed with and error")
		} else {
			ws.logger.Debug().Str("id", id).Msg("session destroyed")
		}
	})

	return nil
}

func (ws *WebSocketHandler) Shutdown() error {
	ws.shutdown <- true
	return nil
}

func (ws *WebSocketHandler) Upgrade(w http.ResponseWriter, r *http.Request) error {
	ws.logger.Debug().Msg("attempting to upgrade connection")

	socket, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Error().Err(err).Msg("failed to upgrade connection")
		return err
	}

	id, admin, err := ws.authenticate(r)
	if err != nil {
		ws.logger.Warn().Err(err).Msg("authenticatetion failed")

		if err = socket.WriteJSON(message.Disconnect{
			Event:   event.SYSTEM_DISCONNECT,
			Message: "invalid password",
		}); err != nil {
			ws.logger.Error().Err(err).Msg("failed to send disconnect")
		}

		if err = socket.Close(); err != nil {
			return err
		}
		return nil
	}

	ok, reason, err := ws.handler.SocketConnected(id, socket)
	if err != nil {
		ws.logger.Error().Err(err).Msg("connection failed")
		return err
	}

	if !ok {
		if err = socket.WriteJSON(message.Disconnect{
			Event:   event.SYSTEM_DISCONNECT,
			Message: reason,
		}); err != nil {
			ws.logger.Error().Err(err).Msg("failed to send disconnect")
		}

		if err = socket.Close(); err != nil {
			return err
		}

		return nil
	}

	ws.sessions.New(id, admin, socket)

	ws.logger.
		Debug().
		Str("session", id).
		Str("address", socket.RemoteAddr().String()).
		Msg("new connection created")

	defer func() {
		ws.logger.
			Debug().
			Str("session", id).
			Str("address", socket.RemoteAddr().String()).
			Msg("session ended")
	}()

	ws.handle(socket, id)
	return nil
}

func (ws *WebSocketHandler) authenticate(r *http.Request) (string, bool, error) {
	id, err := utils.NewUID(32)
	if err != nil {
		return "", false, err
	}

	passwords, ok := r.URL.Query()["password"]
	if !ok || len(passwords[0]) < 1 {
		return "", false, fmt.Errorf("no password provided")
	}

	if passwords[0] == ws.conf.AdminPassword {
		return id, true, nil
	}

	if passwords[0] == ws.conf.Password {
		return id, false, nil
	}

	return "", false, fmt.Errorf("invalid password: %s", passwords[0])
}

func (ws *WebSocketHandler) handle(socket *websocket.Conn, id string) {
	bytes := make(chan []byte)
	cancel := make(chan struct{})
	ticker := time.NewTicker(pingPeriod)

	go func() {
		defer func() {
			ticker.Stop()
			ws.logger.Debug().Str("address", socket.RemoteAddr().String()).Msg("handle socket ending")
			ws.handler.SocketDisconnected(id)
		}()

		for {
			_, raw, err := socket.ReadMessage()
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
				Str("raw", string(raw)).
				Msg("recieved message from client")
			if err := ws.handler.Message(id, raw); err != nil {
				ws.logger.Error().Err(err).Msg("message handler has failed")
				return
			}
		case <-cancel:
			return
		case _ = <-ticker.C:
			if err := socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
