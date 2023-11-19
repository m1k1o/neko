package websocket

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
	"m1k1o/neko/internal/utils"
	"m1k1o/neko/internal/websocket/handler"
	"m1k1o/neko/internal/websocket/state"
)

const CONTROL_PROTECTION_SESSION = "by_control_protection"

func New(sessions types.SessionManager, desktop types.DesktopManager, capture types.CaptureManager, webrtc types.WebRTCManager, conf *config.WebSocket) *WebSocketHandler {
	logger := log.With().Str("module", "websocket").Logger()

	state := state.New(conf.FileTransferEnabled, conf.FileTransferPath)

	// if control protection is enabled
	if conf.ControlProtection {
		state.Lock("control", CONTROL_PROTECTION_SESSION)
		logger.Info().Msgf("control locked on behalf of control protection")
	}

	// create file transfer directory if not exists
	if conf.FileTransferEnabled {
		if _, err := os.Stat(conf.FileTransferPath); os.IsNotExist(err) {
			err = os.Mkdir(conf.FileTransferPath, os.ModePerm)
			logger.Err(err).Msg("creating file transfer directory")
		}
	}

	// apply default locks
	for _, lock := range conf.Locks {
		state.Lock(lock, "") // empty session ID
	}

	if len(conf.Locks) > 0 {
		logger.Info().Msgf("locked resources: %+v", conf.Locks)
	}

	handler := handler.New(
		sessions,
		desktop,
		capture,
		webrtc,
		state,
	)

	return &WebSocketHandler{
		logger:   logger,
		shutdown: make(chan interface{}),
		conf:     conf,
		sessions: sessions,
		desktop:  desktop,
		webrtc:   webrtc,
		state:    state,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		handler:         handler,
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
	desktop  types.DesktopManager
	webrtc   types.WebRTCManager
	state    *state.State
	conf     *config.WebSocket
	handler  *handler.MessageHandler

	// stats
	conns           uint32
	serverStartedAt time.Time
	lastAdminLeftAt *time.Time
	lastUserLeftAt  *time.Time
}

func (ws *WebSocketHandler) Start() {
	go func() {
		for {
			e, ok := <-ws.sessions.GetEventsChannel()
			if !ok {
				ws.logger.Info().Msg("session channel was closed")
				return
			}

			switch e.Type {
			case types.SESSION_CREATED:
				if err := ws.handler.SessionCreated(e.Id, e.Session); err != nil {
					ws.logger.Warn().Str("id", e.Id).Err(err).Msg("session created with and error")
				} else {
					ws.logger.Debug().Str("id", e.Id).Msg("session created")
				}
			case types.SESSION_CONNECTED:
				if err := ws.handler.SessionConnected(e.Id, e.Session); err != nil {
					ws.logger.Warn().Str("id", e.Id).Err(err).Msg("session connected with and error")
				} else {
					ws.logger.Debug().Str("id", e.Id).Msg("session connected")
				}

				// if control protection is enabled and at least one admin
				// and if room was locked on behalf control protection, unlock
				sess, ok := ws.state.GetLocked("control")
				if ok && ws.conf.ControlProtection && sess == CONTROL_PROTECTION_SESSION && len(ws.sessions.Admins()) > 0 {
					ws.state.Unlock("control")
					ws.sessions.SetControlLocked(false) // TODO: Handle locks in sessions as flags.
					ws.logger.Info().Msgf("control unlocked on behalf of control protection")

					if err := ws.sessions.Broadcast(
						message.AdminLock{
							Event:    event.ADMIN_UNLOCK,
							ID:       e.Id,
							Resource: "control",
						}, nil); err != nil {
						ws.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_UNLOCK)
					}
				}

				// remove outdated stats
				if e.Session.Admin() {
					ws.lastAdminLeftAt = nil
				} else {
					ws.lastUserLeftAt = nil
				}
			case types.SESSION_DESTROYED:
				if err := ws.handler.SessionDestroyed(e.Id); err != nil {
					ws.logger.Warn().Str("id", e.Id).Err(err).Msg("session destroyed with and error")
				} else {
					ws.logger.Debug().Str("id", e.Id).Msg("session destroyed")
				}

				membersCount := len(ws.sessions.Members())
				adminCount := len(ws.sessions.Admins())

				// if control protection is enabled and no admin
				// and room is not locked, lock
				ok := ws.state.IsLocked("control")
				if !ok && ws.conf.ControlProtection && adminCount == 0 {
					ws.state.Lock("control", CONTROL_PROTECTION_SESSION)
					ws.sessions.SetControlLocked(true) // TODO: Handle locks in sessions as flags.
					ws.logger.Info().Msgf("control locked and released on behalf of control protection")
					ws.handler.AdminRelease(e.Id, e.Session)

					if err := ws.sessions.Broadcast(
						message.AdminLock{
							Event:    event.ADMIN_LOCK,
							ID:       e.Id,
							Resource: "control",
						}, nil); err != nil {
						ws.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_LOCK)
					}
				}

				// if this was the last admin
				if e.Session.Admin() && adminCount == 0 {
					now := time.Now()
					ws.lastAdminLeftAt = &now
				}

				// if this was the last user
				if !e.Session.Admin() && membersCount-adminCount == 0 {
					now := time.Now()
					ws.lastUserLeftAt = &now
				}
			case types.SESSION_HOST_SET:
				// TODO: Unused.
			case types.SESSION_HOST_CLEARED:
				// TODO: Unused.
			}
		}
	}()

	go func() {
		for {
			_, ok := <-ws.desktop.GetClipboardUpdatedChannel()
			if !ok {
				ws.logger.Info().Msg("clipboard update channel closed")
				return
			}

			session, ok := ws.sessions.GetHost()
			if !ok {
				return
			}

			err := session.Send(message.Clipboard{
				Event: event.CONTROL_CLIPBOARD,
				Text:  ws.desktop.ReadClipboard(),
			})

			ws.logger.Err(err).Msg("sync clipboard")
		}
	}()

	// watch for file changes and send file list if file transfer is enabled
	if ws.conf.FileTransferEnabled {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			ws.logger.Err(err).Msg("unable to start file transfer dir watcher")
			return
		}

		go func() {
			for {
				select {
				case e, ok := <-watcher.Events:
					if !ok {
						ws.logger.Info().Msg("file transfer dir watcher closed")
						return
					}
					if e.Has(fsnotify.Create) || e.Has(fsnotify.Remove) || e.Has(fsnotify.Rename) {
						ws.logger.Debug().Str("event", e.String()).Msg("file transfer dir watcher event")
						ws.handler.FileTransferRefresh(nil)
					}
				case err := <-watcher.Errors:
					ws.logger.Err(err).Msg("error in file transfer dir watcher")
				}
			}
		}()

		if err := watcher.Add(ws.conf.FileTransferPath); err != nil {
			ws.logger.Err(err).Msg("unable to add file transfer path to watcher")
		}
	}
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
		address:    r.RemoteAddr,
		connection: connection,
	}

	ok, reason := ws.handler.Connected(admin, socket.Address())
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

		Banned: ws.state.AllBanned(),
		Locked: ws.state.AllLocked(),

		ServerStartedAt: ws.serverStartedAt,
		LastAdminLeftAt: ws.lastAdminLeftAt,
		LastUserLeftAt:  ws.lastUserLeftAt,

		ControlProtection: ws.conf.ControlProtection,
		ImplicitControl:   ws.webrtc.ImplicitControl(),
	}
}

func (ws *WebSocketHandler) IsLocked(resource string) bool {
	return ws.state.IsLocked(resource)
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

//
// File transfer
//

func (ws *WebSocketHandler) CanTransferFiles(password string) (bool, error) {
	if !ws.conf.FileTransferEnabled {
		return false, nil
	}

	isAdmin, err := ws.IsAdmin(password)
	if err != nil {
		return false, err
	}

	return isAdmin || !ws.state.IsLocked("file_transfer"), nil
}

func (ws *WebSocketHandler) FileTransferPath(filename string) string {
	return ws.state.FileTransferPath(filename)
}

func (ws *WebSocketHandler) FileTransferEnabled() bool {
	return ws.conf.FileTransferEnabled
}
