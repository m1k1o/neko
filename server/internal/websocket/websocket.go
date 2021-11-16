package websocket

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/config"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
	"m1k1o/neko/internal/utils"
)

func New(sessions types.SessionManager, remote types.RemoteManager, broadcast types.BroadcastManager, webrtc types.WebRTCManager, conf *config.WebSocket) *WebSocketHandler {
	logger := log.With().Str("module", "websocket").Logger()

	locks := make(map[string]string)
	for _, lock := range conf.Locks {
		locks[lock] = "" // empty session ID
	}

	if len(conf.Locks) > 0 {
		logger.Info().Msgf("locked resources: %+v", conf.Locks)
	}

	return &WebSocketHandler{
		logger:   logger,
		shutdown: make(chan interface{}),
		conf:     conf,
		sessions: sessions,
		remote:   remote,
		upgrader: websocket.Upgrader{
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
			locked:    locks,
		},
		conns: 0,
	}
}

// Send pings to peer with this period. Must be less than pongWait.
const pingPeriod = 60 * time.Second

type WebSocketHandler struct {
	logger   zerolog.Logger
	wg       sync.WaitGroup
	shutdown chan interface{}
	upgrader websocket.Upgrader
	sessions types.SessionManager
	remote   types.RemoteManager
	conf     *config.WebSocket
	handler  *MessageHandler
	conns    uint32
}

func (ws *WebSocketHandler) Start() {
	ws.sessions.OnCreated(func(id string, session types.Session) {
		if err := ws.handler.SessionCreated(id, session); err != nil {
			ws.logger.Warn().Str("id", id).Err(err).Msg("session created with and error")
		} else {
			ws.logger.Debug().Str("id", id).Msg("session created")
		}
	})

	ws.sessions.OnConnected(func(id string, session types.Session) {
		if err := ws.handler.SessionConnected(id, session); err != nil {
			ws.logger.Warn().Str("id", id).Err(err).Msg("session connected with and error")
		} else {
			ws.logger.Debug().Str("id", id).Msg("session connected")
		}
	})

	ws.sessions.OnDestroy(func(id string, session types.Session) {
		if err := ws.handler.SessionDestroyed(id); err != nil {
			ws.logger.Warn().Str("id", id).Err(err).Msg("session destroyed with and error")
		} else {
			ws.logger.Debug().Str("id", id).Msg("session destroyed")
		}
	})

	ws.wg.Add(1)
	go func() {
		defer func() {
			ws.logger.Info().Msg("shutdown")
			ws.wg.Done()
		}()

		current := ws.remote.ReadClipboard()

		for {
			select {
			case <-ws.shutdown:
				return
			default:
				time.Sleep(100 * time.Millisecond)

				if !ws.sessions.HasHost() {
					continue
				}

				text := ws.remote.ReadClipboard()
				if text == current {
					continue
				}

				session, ok := ws.sessions.GetHost()
				if ok {
					err := session.Send(message.Clipboard{
						Event: event.CONTROL_CLIPBOARD,
						Text:  text,
					})

					if err != nil {
						ws.logger.Err(err).Msg("unable to synchronize clipboard")
					}
				}

				current = text
			}
		}
	}()
}

func (ws *WebSocketHandler) Shutdown() error {
	close(ws.shutdown)
	ws.wg.Wait()
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

		if err = connection.WriteJSON(message.SystemMessage{
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

	ok, reason, err := ws.handler.Connected(admin, socket)
	if err != nil {
		ws.logger.Error().Err(err).Msg("connection failed")
		return err
	}

	if !ok {
		if err = connection.WriteJSON(message.SystemMessage{
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

	atomic.AddUint32(&ws.conns, uint32(1))

	defer func() {
		ws.logger.
			Debug().
			Str("session", id).
			Str("address", connection.RemoteAddr().String()).
			Msg("session ended")

		atomic.AddUint32(&ws.conns, ^uint32(0))
	}()

	ws.handle(connection, id)
	return nil
}

func (ws *WebSocketHandler) Stats() types.Stats {
	host := ""
	session, ok := ws.sessions.GetHost()
	if ok {
		host = session.ID()
	}

	return types.Stats{
		Connections: atomic.LoadUint32(&ws.conns),
		Host:        host,
		Members:     ws.sessions.Members(),
	}
}

func (ws *WebSocketHandler) IsAdmin(password string) (bool, error) {
	if password == ws.conf.AdminPassword {
		return true, nil
	}

	if password == ws.conf.Password {
		return false, nil
	}

	return false, fmt.Errorf("invalid password")
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

	isAdmin, err := ws.IsAdmin(passwords[0])
	return id, ip, isAdmin, err
}

func (ws *WebSocketHandler) handle(connection *websocket.Conn, id string) {
	bytes := make(chan []byte)
	cancel := make(chan struct{})
	ticker := time.NewTicker(pingPeriod)

	ws.wg.Add(1)
	go func() {
		defer func() {
			ticker.Stop()
			ws.logger.Debug().Str("address", connection.RemoteAddr().String()).Msg("handle socket ending")
			ws.handler.Disconnected(id)
			ws.wg.Done()
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
		case <-ws.shutdown:
			if err := connection.WriteJSON(message.SystemMessage{
				Event:   event.SYSTEM_DISCONNECT,
				Message: "server_shutdown",
			}); err != nil {
				ws.logger.Err(err).Msg("failed to send disconnect")
			}

			if err := connection.Close(); err != nil {
				ws.logger.Err(err).Msg("connection closed with an error")
			}
			return
		case <-cancel:
			return
		case <-ticker.C:
			if err := connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
