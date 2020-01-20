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

func (h *MessageHandler) SocketConnected(id string, socket *websocket.Conn) error {
	return nil
}

func (h *MessageHandler) SocketDisconnected(id string) error {
	return h.sessions.Destroy(id)
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
	case event.SIGNAL_PROVIDE:
		payload := message.Signal{}
		return errors.Wrapf(
			utils.Unmarshal(&payload, raw, func() error {
				return h.webrtc.CreatePeer(id, payload.SDP)
			}), "%s failed", header.Event)
	case event.IDENTITY_DETAILS:
		payload := &message.IdentityDetails{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.identityDetails(id, session, payload)
			}), "%s failed", header.Event)
	case event.CONTROL_RELEASE:
		return errors.Wrapf(h.controlRelease(id, session), "%s failed", header.Event)
	case event.CONTROL_REQUEST:
		return errors.Wrapf(h.controlRequest(id, session), "%s failed", header.Event)
	default:
		return errors.Errorf("unknown message event %s", header.Event)
	}
}
