package websocket

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"n.eko.moe/neko/internal/types"
	"n.eko.moe/neko/internal/types/event"
	"n.eko.moe/neko/internal/types/message"
	"n.eko.moe/neko/internal/utils"
)

type MessageHandler struct {
	logger   zerolog.Logger
	sessions types.SessionManager
	webrtc   types.WebRTCManager
	remote   types.RemoteManager
	banned   map[string]bool
	locked   bool
}

func (h *MessageHandler) Connected(id string, socket *WebSocket) (bool, string, error) {
	address := socket.Address()
	if address == "" {
		h.logger.Debug().Msg("no remote address")
	} else {
		ok, banned := h.banned[address]
		if ok && banned {
			h.logger.Debug().Str("address", address).Msg("banned")
			return false, "banned", nil
		}
	}

	if h.locked {
		session, ok := h.sessions.Get(id)
		if !ok || !session.Admin() {
			h.logger.Debug().Msg("server locked")
			return false, "locked", nil
		}
	}

	return true, "", nil
}

func (h *MessageHandler) Disconnected(id string) error {
	if h.locked && len(h.sessions.Admins()) == 0 {
		h.locked = false
	}

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
	case event.SIGNAL_ANSWER:
		payload := &message.SignalAnswer{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.signalAnswer(id, session, payload)
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
	case event.CONTROL_CLIPBOARD:
		payload := &message.Clipboard{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.controlClipboard(id, session, payload)
			}), "%s failed", header.Event)
	case event.CONTROL_KEYBOARD:
		payload := &message.Keyboard{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.controlKeyboard(id, session, payload)
			}), "%s failed", header.Event)


	// Chat Events
	case event.CHAT_MESSAGE:
		payload := &message.ChatReceive{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.chat(id, session, payload)
			}), "%s failed", header.Event)
	case event.CHAT_EMOTE:
		payload := &message.EmoteReceive{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.chatEmote(id, session, payload)
			}), "%s failed", header.Event)

	// Screen Events
	case event.SCREEN_RESOLUTION:
		return errors.Wrapf(h.screenResolution(id, session), "%s failed", header.Event)
	case event.SCREEN_CONFIGURATIONS:
		return errors.Wrapf(h.screenConfigurations(id, session), "%s failed", header.Event)
	case event.SCREEN_SET:
		payload := &message.ScreenResolution{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.screenSet(id, session, payload)
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
