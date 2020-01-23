package websocket

import (
	"encoding/json"
	"strings"

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
	banned   map[string]bool
	locked   bool
}

func (h *MessageHandler) SocketConnected(id string, socket *websocket.Conn) (bool, string, error) {
	remote := socket.RemoteAddr().String()
	if remote != "" {
		address := strings.SplitN(remote, ":", -1)
		if len(address[0]) < 1 {
			h.logger.Debug().Str("address", remote).Msg("no remote address, baling")
		} else {

			ok, banned := h.banned[address[0]]
			if ok && banned {
				h.logger.Debug().Str("address", remote).Msg("banned")
				return false, "This IP has been banned", nil
			}
		}
	}

	if h.locked {
		h.logger.Debug().Str("address", remote).Msg("locked")
		return false, "Server is currently locked", nil
	}
	return true, "", nil
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
	// Signal Events
	case event.SIGNAL_PROVIDE:
		payload := message.Signal{}
		return errors.Wrapf(
			utils.Unmarshal(&payload, raw, func() error {
				return h.webrtc.CreatePeer(id, payload.SDP)
			}), "%s failed", header.Event)

	// Identity Events
	case event.IDENTITY_DETAILS:
		payload := &message.IdentityDetails{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.identityDetails(id, session, payload)
			}), "%s failed", header.Event)

	// Control Events
	case event.CONTROL_RELEASE:
		return errors.Wrapf(h.controlRelease(id, session), "%s failed", header.Event)
	case event.CONTROL_REQUEST:
		return errors.Wrapf(h.controlRequest(id, session), "%s failed", header.Event)
	case event.CONTROL_GIVE:
		payload := &message.Control{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.controlGive(id, session, payload)
			}), "%s failed", header.Event)

	// Chat Events
	case event.CHAT_MESSAGE:
		payload := &message.ChatRecieve{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.chat(id, session, payload)
			}), "%s failed", header.Event)
	case event.CHAT_EMOTE:
		payload := &message.EmoteRecieve{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.chatEmote(id, session, payload)
			}), "%s failed", header.Event)

	// Admin Events
	case event.ADMIN_LOCK:
		return errors.Wrapf(h.adminLock(id, session), "%s failed", header.Event)
	case event.ADMIN_UNLOCK:
		return errors.Wrapf(h.adminUnlock(id, session), "%s failed", header.Event)
	case event.ADMIN_CONTROL:
		return errors.Wrapf(h.adminControl(id, session), "%s failed", header.Event)
	case event.ADMIN_RELEASE:
		return errors.Wrapf(h.adminRelease(id, session), "%s failed", header.Event)
	case event.ADMIN_GIVE:
		payload := &message.Admin{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.adminGive(id, session, payload)
			}), "%s failed", header.Event)
	case event.ADMIN_BAN:
		payload := &message.Admin{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.adminBan(id, session, payload)
			}), "%s failed", header.Event)
	case event.ADMIN_KICK:
		payload := &message.Admin{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.adminKick(id, session, payload)
			}), "%s failed", header.Event)
	case event.ADMIN_MUTE:
		payload := &message.Admin{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.adminMute(id, session, payload)
			}), "%s failed", header.Event)
	case event.ADMIN_UNMUTE:
		payload := &message.Admin{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.adminUnmute(id, session, payload)
			}), "%s failed", header.Event)
	default:
		return errors.Errorf("unknown message event %s", header.Event)
	}
}
