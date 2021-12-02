package websocket

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
	"m1k1o/neko/internal/utils"
)

type MessageHandler struct {
	logger    zerolog.Logger
	sessions  types.SessionManager
	webrtc    types.WebRTCManager
	remote    types.RemoteManager
	broadcast types.BroadcastManager
	banned    map[string]string // IP -> session ID (that banned it)
	locked    map[string]string // resource name -> session ID (that locked it)
}

func (h *MessageHandler) Connected(admin bool, socket *WebSocket) (bool, string) {
	address := socket.Address()
	if address == "" {
		h.logger.Debug().Msg("no remote address")
	} else {
		_, ok := h.banned[address]
		if ok {
			h.logger.Debug().Str("address", address).Msg("banned")
			return false, "banned"
		}
	}

	_, ok := h.locked["login"]
	if ok && !admin {
		h.logger.Debug().Msg("server locked")
		return false, "locked"
	}

	return true, ""
}

func (h *MessageHandler) Disconnected(id string) {
	h.sessions.Destroy(id)
}

func (h *MessageHandler) Message(id string, raw []byte) error {
	header := message.Message{}
	if err := json.Unmarshal(raw, &header); err != nil {
		return err
	}

	session, ok := h.sessions.Get(id)
	if !ok {
		return errors.Errorf("unknown session id %s", id)
	}

	switch header.Event {
	// Signal Events
	case event.SIGNAL_OFFER:
		payload := &message.SignalOffer{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.signalRemoteOffer(id, session, payload)
			}), "%s failed", header.Event)
	case event.SIGNAL_ANSWER:
		payload := &message.SignalAnswer{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.signalRemoteAnswer(id, session, payload)
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

	// Boradcast Events
	case event.BORADCAST_CREATE:
		payload := &message.BroadcastCreate{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.boradcastCreate(session, payload)
			}), "%s failed", header.Event)
	case event.BORADCAST_DESTROY:
		return errors.Wrapf(h.boradcastDestroy(session), "%s failed", header.Event)

	// Admin Events
	case event.ADMIN_LOCK:
		payload := &message.AdminLock{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.adminLock(id, session, payload)
			}), "%s failed", header.Event)
	case event.ADMIN_UNLOCK:
		payload := &message.AdminLock{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.adminUnlock(id, session, payload)
			}), "%s failed", header.Event)
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
