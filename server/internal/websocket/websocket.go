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

const CONTROL_PROTECTION_SESSION = "by_control_protection"

func New(sessions types.SessionManager, remote types.RemoteManager, broadcast types.BroadcastManager, webrtc types.WebRTCManager, conf *config.WebSocket) *WebSocketHandler {
	logger := log.With().Str("module", "websocket").Logger()

	locks := make(map[string]string)

	// if control protection is enabled
	if conf.ControlProtection {
		locks["control"] = CONTROL_PROTECTION_SESSION
		logger.Info().Msgf("control locked on behalf of control protection")
	}

	// apply default locks
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
			banned:    make(map[string]string),
			locked:    locks,
		},
		serverStartedAt: time.Now(),
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

	// stats
	conns           uint32
	serverStartedAt time.Time
	lastAdminLeftAt *time.Time
	lastUserLeftAt  *time.Time
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

		// if control protection is enabled and at least one admin
		// and if room was locked on behalf control protection, unlock
		sess, ok := ws.handler.locked["control"]
		if ok && ws.conf.ControlProtection && sess == CONTROL_PROTECTION_SESSION && len(ws.sessions.Admins()) > 0 {
			delete(ws.handler.locked, "control")
			ws.logger.Info().Msgf("control unlocked on behalf of control protection")

			if err := ws.sessions.Broadcast(
				message.AdminLock{
					Event:    event.ADMIN_UNLOCK,
					ID:       id,
					Resource: "control",
				}, nil); err != nil {
				ws.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_UNLOCK)
			}
		}

		// remove outdated stats
		if session.Admin() {
			ws.lastAdminLeftAt = nil
		} else {
			ws.lastUserLeftAt = nil
		}
	})

	ws.sessions.OnDestroy(func(id string, session types.Session) {
		if err := ws.handler.SessionDestroyed(id); err != nil {
			ws.logger.Warn().Str("id", id).Err(err).Msg("session destroyed with and error")
		} else {
			ws.logger.Debug().Str("id", id).Msg("session destroyed")
		}

		membersCount := len(ws.sessions.Members())
		adminCount := len(ws.sessions.Admins())

		// if control protection is enabled and no admin
		// and room is not locked, lock
		_, ok := ws.handler.locked["control"]
		if !ok && ws.conf.ControlProtection && adminCount == 0 {
			ws.handler.locked["control"] = CONTROL_PROTECTION_SESSION
			ws.logger.Info().Msgf("control locked and released on behalf of control protection")
			ws.handler.adminRelease(id, session)

			if err := ws.sessions.Broadcast(
				message.AdminLock{
					Event:    event.ADMIN_LOCK,
					ID:       id,
					Resource: "control",
				}, nil); err != nil {
				ws.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_LOCK)
			}
		}

		// if this was the last admin
		if session.Admin() && adminCount == 0 {
			now := time.Now()
			ws.lastAdminLeftAt = &now
		}

		// if this was the last user
		if !session.Admin() && membersCount-adminCount == 0 {
			now := time.Now()
			ws.lastUserLeftAt = &now
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

	id, err := utils.NewUID(32)
	if err != nil {
		ws.logger.Error().Err(err).Msg("failed to generate user id")
		return err
	}

	connection, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Error().Err(err).Msg("failed to upgrade connection")
		return err
	}

	admin, err := ws.authenticate(r)
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
		address:    utils.GetHttpRequestIP(r, ws.conf.Proxy),
		connection: connection,
	}

	ok, reason := ws.handler.Connected(admin, socket)
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

		Banned: ws.handler.banned,
		Locked: ws.handler.locked,

		ServerStartedAt: ws.serverStartedAt,
		LastAdminLeftAt: ws.lastAdminLeftAt,
		LastUserLeftAt:  ws.lastUserLeftAt,

		ControlProtection: ws.conf.ControlProtection,
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

func (ws *WebSocketHandler) authenticate(r *http.Request) (bool, error) {
	passwords, ok := r.URL.Query()["password"]
	if !ok || len(passwords[0]) < 1 {
		return false, fmt.Errorf("no password provided")
	}

	return ws.IsAdmin(passwords[0])
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
