package handler

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
)

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	webrtc types.WebRTCManager,
) *MessageHandlerCtx {
	logger := log.With().Str("module", "handler").Logger()

	return &MessageHandlerCtx{
		logger:    logger,
		sessions:  sessions,
		desktop:   desktop,
		capture:   capture,
		webrtc:    webrtc,
		banned:    make(map[string]bool),
		locked:    false,
	}
}

type MessageHandlerCtx struct {
	logger    zerolog.Logger
	sessions  types.SessionManager
	webrtc    types.WebRTCManager
	desktop   types.DesktopManager
	capture   types.CaptureManager
	banned    map[string]bool
	locked    bool
}

func (h *MessageHandlerCtx) Connected(session types.Session, socket types.WebSocket) (bool, string) {
	address := socket.Address()
	if address != "" {
		ok, banned := h.banned[address]
		if ok && banned {
			h.logger.Debug().Str("address", address).Msg("banned")
			return false, "banned"
		}
	} else {
		h.logger.Debug().Msg("no remote address")
	}

	if h.locked && !session.Admin(){
		h.logger.Debug().Msg("server locked")
		return false, "locked"
	}

	return true, ""
}

func (h *MessageHandlerCtx) Disconnected(id string) error {
	// TODO: Refactor.
	if h.locked && len(h.sessions.Admins()) == 0 {
		h.locked = false
	}

	return h.sessions.Destroy(id)
}

func (h *MessageHandlerCtx) Message(session types.Session, raw []byte) error {
	header := message.Message{}
	if err := json.Unmarshal(raw, &header); err != nil {
		return err
	}

	var err error
	switch header.Event {
	// Signal Events
	case event.SIGNAL_ANSWER:
		payload := &message.SignalAnswer{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.signalAnswer(session, payload)
		})

	// Control Events
	case event.CONTROL_RELEASE:
		err = h.controlRelease(session)
	case event.CONTROL_REQUEST:
		err = h.controlRequest(session)
	case event.CONTROL_GIVE:
		payload := &message.Control{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.controlGive(session, payload)
		})
	case event.CONTROL_CLIPBOARD:
		payload := &message.Clipboard{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.controlClipboard(session, payload)
		})
	case event.CONTROL_KEYBOARD:
		payload := &message.Keyboard{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.controlKeyboard(session, payload)
		})

	// Screen Events
	case event.SCREEN_RESOLUTION:
		err = h.screenSize(session)
	case event.SCREEN_CONFIGURATIONS:
		err = h.screenConfigurations(session)
	case event.SCREEN_SET:
		payload := &message.ScreenResolution{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.screenSizeChange(session, payload)
		})

	// Boradcast Events
	case event.BORADCAST_CREATE:
		payload := &message.BroadcastCreate{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.boradcastCreate(session, payload)
		})
	case event.BORADCAST_DESTROY:
		err = h.boradcastDestroy(session)

	// Admin Events
	case event.ADMIN_LOCK:
		err = h.adminLock(session)
	case event.ADMIN_UNLOCK:
		err = h.adminUnlock(session)
	case event.ADMIN_CONTROL:
		err = h.adminControl(session)
	case event.ADMIN_RELEASE:
		err = h.adminRelease(session)
	case event.ADMIN_GIVE:
		payload := &message.Admin{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.adminGive(session, payload)
		})
	case event.ADMIN_BAN:
		payload := &message.Admin{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.adminBan(session, payload)
		})
	case event.ADMIN_KICK:
		payload := &message.Admin{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.adminKick(session, payload)
		})
	default:
		return errors.Errorf("unknown message event %s", header.Event)
	}

	return errors.Wrapf(err, "%s failed", header.Event)
}
