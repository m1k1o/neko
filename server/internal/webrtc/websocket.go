package webrtc

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"n.eko.moe/neko/internal/nanoid"
)

const (
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 60 * time.Second
)

func (manager *WebRTCManager) Upgrade(w http.ResponseWriter, r *http.Request) error {
	manager.logger.
		Info().
		Msg("Attempting to upgrade ws")

	socket, err := manager.upgrader.Upgrade(w, r, nil)
	if err != nil {
		manager.logger.Error().Err(err).Msg("Failed to upgrade websocket!")
		return err
	}

	sessionID, ok := manager.authenticate(r)
	if ok != true {
		manager.logger.Warn().Msg("Authenticatetion failed")
		if err = socket.Close(); err != nil {
			return err
		}
		return nil
	}

	session := &session{
		id:     sessionID,
		socket: socket,
		mu:     sync.Mutex{},
	}

	manager.logger.
		Info().
		Str("ID", sessionID).
		Str("RemoteAddr", socket.RemoteAddr().String()).
		Msg("Created Session")

	manager.sessions[sessionID] = session

	defer func() {
		manager.destroy(session)
	}()

	if err = manager.onConnected(session); err != nil {
		manager.logger.Error().Err(err).Msg("onConnected failed!")
		return nil
	}

	manager.handleWS(session)

	return nil
}

func (manager *WebRTCManager) authenticate(r *http.Request) (sessionID string, ok bool) {

	passwords, ok := r.URL.Query()["password"]
	if !ok || len(passwords[0]) < 1 {
		return "", false
	}

	if passwords[0] != manager.password {
		manager.logger.Warn().Str("Password", passwords[0]).Msg("Wrong password: ")
		return "", false
	}

	id, err := nanoid.NewIDSize(32)
	if err != nil {
		return "", false
	}
	return id, true
}

func (manager *WebRTCManager) onConnected(session *session) error {
	if err := session.send(messageIdentityProvide{
		message: message{Event: "identity/provide"},
		ID:      session.id,
	}); err != nil {
		return err
	}
	return nil
}

func (manager *WebRTCManager) onMessage(session *session, raw []byte) error {
	message := message{}
	if err := json.Unmarshal(raw, &message); err != nil {
		return err
	}

	switch message.Event {
	case "sdp/provide":
		return errors.Wrap(manager.createPeer(session, raw), "sdp/provide failed")
	case "control/release":
		return errors.Wrap(manager.controlRelease(session), "control/release failed")
	case "control/request":
		return errors.Wrap(manager.controlRequest(session), "control/request failed")
	default:
		manager.logger.Warn().Msgf("Unknown client method %s", message.Event)
	}

	return nil
}

func (manager *WebRTCManager) handleWS(session *session) {
	bytes := make(chan []byte)
	cancel := make(chan struct{})
	ticker := time.NewTicker(pingPeriod)

	go func() {
		defer func() {
			ticker.Stop()
			manager.logger.Info().Str("RemoteAddr", session.socket.RemoteAddr().String()).Msg("Handle WS ending")
			manager.destroy(session)
		}()

		for {
			_, raw, err := session.socket.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					manager.logger.Warn().Err(err).Msg("ReadMessage error")
				}
				break
			}
			bytes <- raw
		}
	}()

	for {
		select {
		case raw := <-bytes:
			manager.logger.Info().
				Str("ID", session.id).
				Str("Message", string(raw)).
				Msg("Reading from Websocket")
			if err := manager.onMessage(session, raw); err != nil {
				manager.logger.Error().Err(err).Msg("onClientMessage has failed")
				return
			}
		case <-cancel:
			return
		case _ = <-ticker.C:
			if err := session.socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (manager *WebRTCManager) destroy(session *session) {
	if manager.controller == session.id {
		manager.controller = ""
		manager.clearKeys()
		for id, sess := range manager.sessions {
			if id != session.id {
				if err := sess.send(message{Event: "control/released"}); err != nil {
					manager.logger.Error().Err(err).Msg("session.send has failed")
				}
			}
		}
	}

	if err := session.destroy(); err != nil {
		manager.logger.Error().Err(err).Msg("session.destroy has failed")
	}

	delete(manager.sessions, session.id)
}

func (manager *WebRTCManager) controlRelease(session *session) error {
	if manager.controller == session.id {
		manager.controller = ""
		manager.clearKeys()

		if err := session.send(message{Event: "control/release"}); err != nil {
			return err
		}

		for id, sess := range manager.sessions {
			if id != session.id {
				if err := sess.send(message{Event: "control/released"}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (manager *WebRTCManager) controlRequest(session *session) error {
	if manager.controller == "" {
		manager.controller = session.id

		if err := session.send(message{Event: "control/give"}); err != nil {
			return err
		}

		for id, sess := range manager.sessions {
			if id != session.id {
				if err := sess.send(message{Event: "control/given"}); err != nil {
					return err
				}
			}
		}
	} else {
		if err := session.send(message{Event: "control/locked"}); err != nil {
			return err
		}

		controller, ok := manager.sessions[manager.controller]
		if ok {
			controller.send(message{Event: "control/requesting"})
		}
	}
	return nil
}
