package handler

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
	"m1k1o/neko/internal/utils"
	"m1k1o/neko/internal/websocket/state"
)

type MessageHandler struct {
	logger   zerolog.Logger
	sessions types.SessionManager
	desktop  types.DesktopManager
	capture  types.CaptureManager
	webrtc   types.WebRTCManager
	state    *state.State
}

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	webrtc types.WebRTCManager,
	state *state.State,
) *MessageHandler {
	return &MessageHandler{
		logger:   log.With().Str("module", "websocket").Str("submodule", "handler").Logger(),
		sessions: sessions,
		desktop:  desktop,
		capture:  capture,
		webrtc:   webrtc,
		state:    state,
	}
}

func (h *MessageHandler) Connected(admin bool, address string) (bool, string) {
	if address == "" {
		h.logger.Debug().Msg("no remote address")
	} else {
		if h.state.IsBanned(address) {
			h.logger.Debug().Str("address", address).Msg("banned")
			return false, "banned"
		}
	}

	if h.state.IsLocked("login") && !admin {
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
	case event.SIGNAL_CANDIDATE:
		payload := &message.SignalCandidate{}
		return errors.Wrapf(
			utils.Unmarshal(payload, raw, func() error {
				return h.signalRemoteCandidate(id, session, payload)
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

	// File Transfer Events
	case event.FILETRANSFER_REFRESH:
		return errors.Wrapf(h.FileTransferRefresh(session), "%s failed", header.Event)

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
		return errors.Wrapf(h.AdminRelease(id, session), "%s failed", header.Event)
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
