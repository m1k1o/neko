package legacy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/m1k1o/neko/server/internal/api"
	oldEvent "github.com/m1k1o/neko/server/internal/http/legacy/event"
	oldMessage "github.com/m1k1o/neko/server/internal/http/legacy/message"
	oldTypes "github.com/m1k1o/neko/server/internal/http/legacy/types"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
	"github.com/m1k1o/neko/server/pkg/utils"

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
)

type LegacyHandler struct {
	logger     zerolog.Logger
	serverAddr string
	pathPrefix string
	bannedIPs  map[string]struct{}
	sessionIPs map[string]string
	wsDialer   *websocket.Dialer
}

func New(serverAddr, pathPrefix string) *LegacyHandler {
	// Init

	return &LegacyHandler{
		logger:     log.With().Str("module", "legacy").Logger(),
		serverAddr: serverAddr,
		pathPrefix: pathPrefix,
		bannedIPs:  make(map[string]struct{}),
		sessionIPs: make(map[string]string),
		wsDialer: &websocket.Dialer{
			Proxy:            nil, // disable proxy for local requests
			HandshakeTimeout: 45 * time.Second,
		},
	}
}

func (h *LegacyHandler) Route(r types.Router) {
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) error {
		s := h.newSession(r)

		// create a new websocket connection
		connClient, err := DefaultUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return utils.HttpError(http.StatusInternalServerError).
				WithInternalErr(err).
				Msg("couldn't upgrade connection to websocket")
		}
		defer connClient.Close()
		s.connClient = connClient

		if h.isBanned(r) {
			s.toClient(&oldMessage.SystemMessage{
				Event:   oldEvent.SYSTEM_DISCONNECT,
				Title:   "banned ip",
				Message: "you are banned",
			})
		}

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
		connBackend, _, err := h.wsDialer.Dial("ws://"+h.serverAddr+path.Join(s.pathPrefix, "/api/ws")+"?token="+url.QueryEscape(s.token), nil)
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
					errc <- fmt.Errorf("src read message error: %w", err)
					dst.WriteMessage(websocket.CloseMessage, m)
					break
				}

				// handle text messages
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
							Message: strings.ReplaceAll(err.Error(), ErrBackendRespone.Error()+": ", ""),
						})
						continue
					}

					if errors.Is(err, ErrWebsocketSend) {
						errc <- fmt.Errorf("dst write message error: %w", err)
						break
					}

					h.logger.Error().Err(err).Msg("couldn't rewrite text message")
					continue
				}

				// forward ping pong messages
				if msgType == websocket.PingMessage ||
					msgType == websocket.PongMessage {
					err = dst.WriteMessage(msgType, msg)
					if err != nil {
						errc <- err
						break
					}
					continue
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
			message = "websocketproxy: Error when copying from backend to client"
		case err = <-errBackend:
			message = "websocketproxy: Error when copying from client to backend"
		}

		if e, ok := err.(*websocket.CloseError); !ok || e.Code == websocket.CloseAbnormalClosure {
			h.logger.Error().Err(err).Msg(message)
		}

		return nil
	})

	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) error {
		if h.isBanned(r) {
			return utils.HttpForbidden("banned ip")
		}

		s := h.newSession(r)

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

		// get stats
		newStats := types.Stats{}
		err = s.apiReq(http.MethodGet, "/api/stats", nil, &newStats)
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
			}
		}

		locks, err := s.settingsToLocks(settings)
		if err != nil {
			return err
		}

		stats.Host = newStats.HostId
		// TODO: stats.Banned, not implemented yet
		stats.Locked = locks
		stats.ServerStartedAt = newStats.ServerStartedAt
		stats.LastAdminLeftAt = newStats.LastAdminLeftAt
		stats.LastUserLeftAt = newStats.LastUserLeftAt
		stats.ControlProtection = settings.ControlProtection
		stats.ImplicitControl = settings.ImplicitHosting

		return json.NewEncoder(w).Encode(stats)
	})

	r.Get("/screenshot.jpg", func(w http.ResponseWriter, r *http.Request) error {
		if h.isBanned(r) {
			return utils.HttpForbidden("banned ip")
		}

		s := h.newSession(r)

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

		quality := r.URL.Query().Get("quality")

		// get the screenshot
		body, headers, err := s.req(http.MethodGet, "/api/room/screen/shot.jpg?quality="+url.QueryEscape(quality), nil, nil)
		if err != nil {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}

		// copy headers
		w.Header().Set("Content-Length", headers.Get("Content-Length"))
		w.Header().Set("Content-Type", headers.Get("Content-Type"))

		// copy the body to the response writer
		_, err = io.Copy(w, body)
		return err
	})

	// allow downloading and uploading files
	r.Get("/file", func(w http.ResponseWriter, r *http.Request) error {
		if h.isBanned(r) {
			return utils.HttpForbidden("banned ip")
		}

		s := h.newSession(r)

		// create a new session
		username := r.URL.Query().Get("usr")
		password := r.URL.Query().Get("pwd")
		err := s.create(username, password)
		if err != nil {
			return utils.HttpForbidden(err.Error())
		}
		defer s.destroy()

		filename := r.URL.Query().Get("filename")

		body, headers, err := s.req(http.MethodGet, "/api/filetransfer?filename="+url.QueryEscape(filename), r.Header, nil)
		if err != nil {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}

		// copy headers
		w.Header().Set("Content-Length", headers.Get("Content-Length"))
		w.Header().Set("Content-Type", headers.Get("Content-Type"))

		// copy the body to the response writer
		_, err = io.Copy(w, body)
		return err
	})

	r.Post("/file", func(w http.ResponseWriter, r *http.Request) error {
		if h.isBanned(r) {
			return utils.HttpForbidden("banned ip")
		}

		s := h.newSession(r)

		// create a new session
		username := r.URL.Query().Get("usr")
		password := r.URL.Query().Get("pwd")
		err := s.create(username, password)
		if err != nil {
			return utils.HttpForbidden(err.Error())
		}
		defer s.destroy()

		body, _, err := s.req(http.MethodPost, "/api/filetransfer", r.Header, r.Body)
		if err != nil {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}

		// copy the body to the response writer
		_, err = io.Copy(w, body)
		return err
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) error {
		_, err := w.Write([]byte("true"))
		return err
	})
}

func (h *LegacyHandler) ban(sessionId string) error {
	// find session by id
	ip, ok := h.sessionIPs[sessionId]
	if !ok {
		return fmt.Errorf("session not found")
	}

	h.bannedIPs[ip] = struct{}{}
	return nil
}

func (h *LegacyHandler) isBanned(r *http.Request) bool {
	ip := getIp(r)
	_, ok := h.bannedIPs[ip]
	return ok
}

func getIp(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		if e, ok := err.(*net.AddrError); ok && e.Err == "missing port in address" {
			return r.RemoteAddr
		}
		return ""
	}

	return ip
}
