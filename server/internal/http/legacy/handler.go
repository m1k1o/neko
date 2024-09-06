package legacy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"m1k1o/neko/internal/api"
	"m1k1o/neko/internal/api/room"
	oldEvent "m1k1o/neko/internal/http/legacy/event"
	oldMessage "m1k1o/neko/internal/http/legacy/message"
	oldTypes "m1k1o/neko/internal/http/legacy/types"

	"m1k1o/neko/pkg/types"
	"m1k1o/neko/pkg/types/event"
	"m1k1o/neko/pkg/types/message"
	"m1k1o/neko/pkg/utils"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// DefaultUpgrader specifies the parameters for upgrading an HTTP
	// connection to a WebSocket connection.
	DefaultUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// DefaultDialer is a dialer with all fields set to the default zero values.
	DefaultDialer = websocket.DefaultDialer
)

type LegacyHandler struct {
	logger     zerolog.Logger
	serverAddr string
	startedAt  time.Time
}

func New() *LegacyHandler {
	// Init

	return &LegacyHandler{
		logger:     log.With().Str("module", "legacy").Logger(),
		serverAddr: "127.0.0.1:8080",
		startedAt:  time.Now(),
	}
}

func (h *LegacyHandler) Route(r types.Router) {
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) error {
		s := newSession(h.logger, h.serverAddr)

		// create a new websocket connection
		connClient, err := DefaultUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return utils.HttpError(http.StatusInternalServerError).
				WithInternalErr(err).
				Msg("couldn't upgrade connection to websocket")
		}
		defer connClient.Close()
		s.connClient = connClient

		// create a new session
		username := r.URL.Query().Get("username")
		password := r.URL.Query().Get("password")
		err = s.create(username, password)
		if err != nil {
			h.logger.Error().Err(err).Msg("couldn't create a new session")

			s.toClient(&oldMessage.SystemMessage{
				Event:   oldEvent.SYSTEM_DISCONNECT,
				Title:   "couldn't create a new session",
				Message: err.Error(),
			})

			// we can't return HTTP error here because the connection is already upgraded
			return nil
		}
		defer s.destroy()

		// dial to the remote backend
		connBackend, _, err := DefaultDialer.Dial("ws://"+h.serverAddr+"/api/ws?token="+s.token, nil)
		if err != nil {
			h.logger.Error().Err(err).Msg("couldn't dial to the remote backend")

			s.toClient(&oldMessage.SystemMessage{
				Event:   oldEvent.SYSTEM_DISCONNECT,
				Title:   "couldn't dial to the remote backend",
				Message: err.Error(),
			})

			// we can't return HTTP error here because the connection is already upgraded
			return nil
		}
		defer connBackend.Close()
		s.connBackend = connBackend

		// request signal
		if err = s.toBackend(event.SIGNAL_REQUEST, message.SignalRequest{}); err != nil {
			h.logger.Error().Err(err).Msg("couldn't request signal")

			s.toClient(&oldMessage.SystemMessage{
				Event:   oldEvent.SYSTEM_DISCONNECT,
				Title:   "couldn't request signal",
				Message: err.Error(),
			})

			// we can't return HTTP error here because the connection is already upgraded
			return nil
		}

		// copy messages between the client and the backend
		errClient := make(chan error, 1)
		errBackend := make(chan error, 1)
		replicateWebsocketConn := func(dst, src *websocket.Conn, errc chan error, rewriteTextMessage func([]byte) error) {
			for {
				msgType, msg, err := src.ReadMessage()
				if err != nil {
					m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
					if e, ok := err.(*websocket.CloseError); ok {
						if e.Code != websocket.CloseNoStatusReceived {
							m = websocket.FormatCloseMessage(e.Code, e.Text)
						}
					}
					errc <- err
					dst.WriteMessage(websocket.CloseMessage, m)
					break
				}
				if msgType == websocket.TextMessage {
					err = rewriteTextMessage(msg)

					if err == nil {
						continue
					}

					if errors.Is(err, ErrBackendRespone) {
						h.logger.Error().Err(err).Msg("backend response error")

						s.toClient(&oldMessage.SystemMessage{
							Event:   oldEvent.SYSTEM_ERROR,
							Title:   "backend response error",
							Message: err.Error(),
						})
						continue
					} else if errors.Is(err, ErrWebsocketSend) {
						errc <- err
						break
					} else {
						h.logger.Error().Err(err).Msg("couldn't rewrite text message")
					}
				}
			}
		}

		// backend -> client
		go replicateWebsocketConn(connClient, connBackend, errClient, s.wsToClient)

		// client -> backend
		go replicateWebsocketConn(connBackend, connClient, errBackend, s.wsToBackend)

		var message string
		select {
		case err = <-errClient:
			message = "websocketproxy: Error when copying from backend to client: %v"
		case err = <-errBackend:
			message = "websocketproxy: Error when copying from client to backend: %v"
		}

		if e, ok := err.(*websocket.CloseError); !ok || e.Code == websocket.CloseAbnormalClosure {
			h.logger.Error().Err(err).Msg(message)
		}

		return nil
	})

	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) error {
		s := newSession(h.logger, h.serverAddr)

		// create a new session
		username := r.URL.Query().Get("usr")
		password := r.URL.Query().Get("pwd")
		err := s.create(username, password)
		if err != nil {
			return utils.HttpForbidden(err.Error())
		}
		defer s.destroy()

		if !s.isAdmin {
			return utils.HttpUnauthorized().Msg("bad authorization")
		}

		w.Header().Set("Content-Type", "application/json")

		// get all sessions
		sessions := []api.SessionDataPayload{}
		err = s.apiReq(http.MethodGet, "/api/sessions", nil, &sessions)
		if err != nil {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}

		// get current control status
		control := room.ControlStatusPayload{}
		err = s.apiReq(http.MethodGet, "/api/room/control", nil, &control)
		if err != nil {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}

		// get settings
		settings := types.Settings{}
		err = s.apiReq(http.MethodGet, "/api/room/settings", nil, &settings)
		if err != nil {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}

		var stats oldTypes.Stats

		// create empty array so that it's not null in json
		stats.Members = []*oldTypes.Member{}

		for _, session := range sessions {
			if session.State.IsConnected {
				stats.Connections++
				member, err := profileToMember(session.ID, session.Profile)
				if err != nil {
					return utils.HttpInternalServerError().WithInternalErr(err)
				}
				// append members
				stats.Members = append(stats.Members, member)
			} else if session.State.NotConnectedSince != nil {
				//
				// TODO: This wont work if the user is removed after the session is closed
				//
				// populate last admin left time
				if session.Profile.IsAdmin && (stats.LastAdminLeftAt == nil || (*session.State.NotConnectedSince).After(*stats.LastAdminLeftAt)) {
					stats.LastAdminLeftAt = session.State.NotConnectedSince
				}
				// populate last user left time
				if !session.Profile.IsAdmin && (stats.LastUserLeftAt == nil || (*session.State.NotConnectedSince).After(*stats.LastUserLeftAt)) {
					stats.LastUserLeftAt = session.State.NotConnectedSince
				}
			}
		}

		locks, err := s.settingsToLocks(settings)
		if err != nil {
			return err
		}

		stats.Host = control.HostId
		// TODO: stats.Banned, not implemented yet
		stats.Locked = locks
		stats.ServerStartedAt = h.startedAt
		stats.ControlProtection = settings.ControlProtection
		stats.ImplicitControl = settings.ImplicitHosting

		return json.NewEncoder(w).Encode(stats)
	})

	/*
		r.Get("/screenshot.jpg", func(w http.ResponseWriter, r *http.Request) error {
			password := r.URL.Query().Get("pwd")
			isAdmin, err := webSocketHandler.IsAdmin(password)
			if err != nil {
				return utils.HttpForbidden(err)
			}

			if !isAdmin {
				return utils.HttpUnauthorized().Msg("bad authorization")
			}

			if webSocketHandler.IsLocked("login") {
				return utils.HttpError(http.StatusLocked).Msg("room is locked")
			}

			quality, err := strconv.Atoi(r.URL.Query().Get("quality"))
			if err != nil {
				quality = 90
			}

			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Content-Type", "image/jpeg")

			img := desktop.GetScreenshotImage()
			if err := jpeg.Encode(w, img, &jpeg.Options{Quality: quality}); err != nil {
				return utils.HttpInternalServerError().WithInternalErr(err)
			}

			return nil
		})

		// allow downloading and uploading files
		if webSocketHandler.FileTransferEnabled() {
			r.Get("/file", func(w http.ResponseWriter, r *http.Request) error {
				return nil
			})

			r.Post("/file", func(w http.ResponseWriter, r *http.Request) error {
				return nil
			})
		}
	*/

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) error {
		_, err := w.Write([]byte("true"))
		return err
	})
}
