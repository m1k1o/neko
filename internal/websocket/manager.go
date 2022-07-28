package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/internal/websocket/handler"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
	"github.com/demodesk/neko/pkg/utils"
)

// send pings to peer with this period - must be less than pongWait
const pingPeriod = 10 * time.Second

// period for sending inactive cursor messages
const inactiveCursorsPeriod = 750 * time.Millisecond

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	webrtc types.WebRTCManager,
) *WebSocketManagerCtx {
	logger := log.With().Str("module", "websocket").Logger()

	return &WebSocketManagerCtx{
		logger:   logger,
		shutdown: make(chan struct{}),
		sessions: sessions,
		desktop:  desktop,
		handler:  handler.New(sessions, desktop, capture, webrtc),
		handlers: []types.WebSocketHandler{},
	}
}

type WebSocketManagerCtx struct {
	logger   zerolog.Logger
	wg       sync.WaitGroup
	shutdown chan struct{}
	sessions types.SessionManager
	desktop  types.DesktopManager
	handler  *handler.MessageHandlerCtx
	handlers []types.WebSocketHandler

	shutdownInactiveCursors chan struct{}
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

		manager.sessions.Broadcast(event.CONTROL_HOST, payload)

		manager.logger.Info().
			Bool("has_host", payload.HasHost).
			Str("host_id", payload.HostID).
			Msg("session host changed")
	})

	manager.sessions.OnSettingsChanged(func(new types.Settings, old types.Settings) {
		// start inactive cursors
		if new.InactiveCursors && !old.InactiveCursors {
			manager.startInactiveCursors()
		}

		// stop inactive cursors
		if !new.InactiveCursors && old.InactiveCursors {
			manager.stopInactiveCursors()
		}

		manager.sessions.Broadcast(event.SYSTEM_SETTINGS, new)
		manager.logger.Info().
			Interface("new", new).
			Interface("old", old).
			Msg("settings changed")
	})

	manager.desktop.OnClipboardUpdated(func() {
		session := manager.sessions.GetHost()
		if session == nil || !session.Profile().CanAccessClipboard {
			return
		}

		manager.logger.Info().Msg("sync clipboard")

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

	if manager.sessions.Settings().InactiveCursors {
		manager.startInactiveCursors()
	}

	manager.logger.Info().Msg("websocket starting")
}

func (manager *WebSocketManagerCtx) Shutdown() error {
	manager.logger.Info().Msg("shutdown")
	close(manager.shutdown)
	manager.stopInactiveCursors()
	manager.wg.Wait()
	return nil
}

func (manager *WebSocketManagerCtx) AddHandler(handler types.WebSocketHandler) {
	manager.handlers = append(manager.handlers, handler)
}

func (manager *WebSocketManagerCtx) Upgrade(checkOrigin types.CheckOrigin) types.RouterHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		upgrader := websocket.Upgrader{
			CheckOrigin: checkOrigin,
			// Do not return any error while handshake
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {},
		}

		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return utils.HttpBadRequest().WithInternalErr(err)
		}

		// Cannot write HTTP response after connection upgrade
		manager.connect(connection, r)
		return nil
	}
}

func (manager *WebSocketManagerCtx) connect(connection *websocket.Conn, r *http.Request) {
	// create new peer
	peer := newPeer(connection)

	session, err := manager.sessions.Authenticate(r)
	if err != nil {
		manager.logger.Warn().Err(err).Msg("authentication failed")
		peer.Destroy(err.Error())
		return
	}

	// add session id to all log messages
	logger := manager.logger.With().Str("session_id", session.ID()).Logger()
	peer.setSessionID(session.ID())

	if !session.Profile().CanConnect {
		logger.Warn().Msg("connection disabled")
		peer.Destroy("connection disabled")
		return
	}

	if session.State().IsConnected {
		logger.Warn().Msg("already connected")

		if !manager.sessions.Settings().MercifulReconnect {
			peer.Destroy("already connected")
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

	manager.handle(connection, peer, session)
}

func (manager *WebSocketManagerCtx) handle(connection *websocket.Conn, peer types.WebSocketPeer, session types.Session) {
	// add session id to logger context
	logger := manager.logger.With().Str("session_id", session.ID()).Logger()

	bytes := make(chan []byte)
	cancel := make(chan struct{})

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	manager.wg.Add(1)
	go func() {
		defer manager.wg.Done()

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

			if data.Event != event.SYSTEM_LOGS {
				logger.Debug().
					Str("address", connection.RemoteAddr().String()).
					Str("event", data.Event).
					Str("payload", string(data.Payload)).
					Msg("received message from client")
			}

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
		case <-manager.shutdown:
			peer.Destroy("connection shutdown")
			return
		case <-ticker.C:
			if err := peer.Ping(); err != nil {
				logger.Err(err).Msg("ping message has failed")
				return
			}
		}
	}
}

func (manager *WebSocketManagerCtx) startInactiveCursors() {
	if manager.shutdownInactiveCursors != nil {
		manager.logger.Warn().Msg("inactive cursors handler already running")
		return
	}

	manager.logger.Info().Msg("starting inactive cursors handler")
	manager.shutdownInactiveCursors = make(chan struct{})

	manager.wg.Add(1)
	go func() {
		defer manager.wg.Done()

		ticker := time.NewTicker(inactiveCursorsPeriod)
		defer ticker.Stop()

		var currentEmpty bool
		var lastEmpty = false

		for {
			select {
			case <-manager.shutdownInactiveCursors:
				manager.logger.Info().Msg("stopping inactive cursors handler")
				manager.shutdownInactiveCursors = nil

				// remove last cursor entries and send empty message
				_ = manager.sessions.PopCursors()
				manager.sessions.InactiveCursorsBroadcast(event.SESSION_CURSORS, []message.SessionCursors{})
				return
			case <-ticker.C:
				cursorsMap := manager.sessions.PopCursors()

				currentEmpty = len(cursorsMap) == 0
				if currentEmpty && lastEmpty {
					continue
				}
				lastEmpty = currentEmpty

				sessionCursors := []message.SessionCursors{}
				for session, cursors := range cursorsMap {
					sessionCursors = append(
						sessionCursors,
						message.SessionCursors{
							ID:      session.ID(),
							Cursors: cursors,
						},
					)
				}

				manager.sessions.InactiveCursorsBroadcast(event.SESSION_CURSORS, sessionCursors)
			}
		}
	}()
}

func (manager *WebSocketManagerCtx) stopInactiveCursors() {
	if manager.shutdownInactiveCursors != nil {
		close(manager.shutdownInactiveCursors)
	}
}
