package websocket

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"n.eko.moe/neko/internal/event"
	"n.eko.moe/neko/internal/message"
	"n.eko.moe/neko/internal/session"
	"n.eko.moe/neko/internal/utils"
	"n.eko.moe/neko/internal/webrtc"
)

type MessageHandler struct {
	logger   zerolog.Logger
	sessions *session.SessionManager
	webrtc   *webrtc.WebRTCManager
}

func (h *MessageHandler) Connected(id string, socket *websocket.Conn) error {
	return nil
}

func (h *MessageHandler) Disconnected(id string) error {
	return h.sessions.Destroy(id)
}

func (h *MessageHandler) Created(id string, session *session.Session) error {
	if err := session.Send(message.IdentityProvide{
		Message: message.Message{Event: event.IDENTITY_PROVIDE},
		ID:      id,
	}); err != nil {
		return err
	}

	return nil
}

func (h *MessageHandler) Destroyed(id string) error {
	if h.sessions.IsHost(id) {
		h.sessions.ClearHost()
		if err := h.sessions.Brodcast(message.Message{Event: event.CONTROL_RELEASED}, []string{id}); err != nil {
			h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.CONTROL_RELEASED)
		}
	}

	return nil
}

func (h *MessageHandler) Message(id string, raw []byte) error {
	header := message.Message{}
	if err := json.Unmarshal(raw, &header); err != nil {
		return err
	}

	session, ok := h.sessions.Get(id)
	if !ok {
		errors.Errorf("unknown session id %s", id)
	}

	switch header.Event {
	case event.SDP_PROVIDE:
		payload := message.SDP{}
		return errors.Wrapf(utils.Unmarshal(&payload, raw, func() error { return h.webrtc.CreatePeer(id, payload.SDP) }), "%s failed", header.Event)
	case event.CONTROL_RELEASE:
		return errors.Wrapf(h.controlRelease(id, session), "%s failed", header.Event)
	case event.CONTROL_REQUEST:
		return errors.Wrapf(h.controlRequest(id, session), "%s failed", header.Event)
	default:
		return errors.Errorf("unknown message event %s", header.Event)
	}
}

func (h *MessageHandler) controlRelease(id string, session *session.Session) error {
	if !h.sessions.IsHost(id) {
		return nil
	}

	h.logger.Debug().Str("id", id).Msgf("host called %s", event.CONTROL_RELEASED)
	h.sessions.ClearHost()

	if err := session.Send(message.Message{Event: event.CONTROL_RELEASE}); err != nil {
		h.logger.Warn().Err(err).Str("id", id).Msgf("sending event %s has failed", event.CONTROL_RELEASE)
		return err
	}

	if err := h.sessions.Brodcast(message.Message{Event: event.CONTROL_RELEASED}, []string{session.ID}); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.CONTROL_RELEASED)
		return err
	}

	return nil
}

func (h *MessageHandler) controlRequest(id string, session *session.Session) error {
	h.logger.Debug().Str("id", id).Msgf("user called %s", event.CONTROL_REQUEST)

	if !h.sessions.HasHost() {
		h.sessions.SetHost(id)

		if err := session.Send(message.Message{Event: event.CONTROL_GIVE}); err != nil {
			h.logger.Warn().Err(err).Str("id", id).Msgf("sending event %s has failed", event.CONTROL_GIVE)
			return err
		}

		if err := h.sessions.Brodcast(message.Message{Event: event.CONTROL_GIVEN}, []string{session.ID}); err != nil {
			h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.CONTROL_GIVEN)
			return err
		}

		return nil
	}

	if err := session.Send(message.Message{Event: event.CONTROL_LOCKED}); err != nil {
		h.logger.Warn().Err(err).Str("id", id).Msgf("sending event %s has failed", event.CONTROL_LOCKED)
		return err
	}

	host, ok := h.sessions.GetHost()
	if ok {
		if err := host.Send(message.Message{Event: event.CONTROL_REQUESTING}); err != nil {
			h.logger.Warn().Err(err).Str("id", id).Msgf("sending event %s has failed", event.CONTROL_REQUESTING)
			return err
		}
	}

	return nil
}
