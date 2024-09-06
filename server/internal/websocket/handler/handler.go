package handler

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
	"github.com/demodesk/neko/pkg/utils"
)

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	webrtc types.WebRTCManager,
) *MessageHandlerCtx {
	return &MessageHandlerCtx{
		logger:   log.With().Str("module", "websocket").Str("submodule", "handler").Logger(),
		sessions: sessions,
		desktop:  desktop,
		capture:  capture,
		webrtc:   webrtc,
	}
}

type MessageHandlerCtx struct {
	logger   zerolog.Logger
	sessions types.SessionManager
	webrtc   types.WebRTCManager
	desktop  types.DesktopManager
	capture  types.CaptureManager
}

func (h *MessageHandlerCtx) Message(session types.Session, data types.WebSocketMessage) bool {
	var err error
	switch data.Event {
	// System Events
	case event.SYSTEM_LOGS:
		payload := &message.SystemLogs{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.systemLogs(session, payload)
		})

	// Signal Events
	case event.SIGNAL_REQUEST:
		payload := &message.SignalRequest{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.signalRequest(session, payload)
		})
	case event.SIGNAL_RESTART:
		err = h.signalRestart(session)
	case event.SIGNAL_OFFER:
		payload := &message.SignalDescription{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.signalOffer(session, payload)
		})
	case event.SIGNAL_ANSWER:
		payload := &message.SignalDescription{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.signalAnswer(session, payload)
		})
	case event.SIGNAL_CANDIDATE:
		payload := &message.SignalCandidate{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.signalCandidate(session, payload)
		})
	case event.SIGNAL_VIDEO:
		payload := &message.SignalVideo{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.signalVideo(session, payload)
		})
	case event.SIGNAL_AUDIO:
		payload := &message.SignalAudio{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.signalAudio(session, payload)
		})

	// Control Events
	case event.CONTROL_RELEASE:
		err = h.controlRelease(session)
	case event.CONTROL_REQUEST:
		err = h.controlRequest(session)
	case event.CONTROL_MOVE:
		payload := &message.ControlPos{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlMove(session, payload)
		})
	case event.CONTROL_SCROLL:
		payload := &message.ControlScroll{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlScroll(session, payload)
		})
	case event.CONTROL_BUTTONPRESS:
		payload := &message.ControlButton{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlButtonPress(session, payload)
		})
	case event.CONTROL_BUTTONDOWN:
		payload := &message.ControlButton{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlButtonDown(session, payload)
		})
	case event.CONTROL_BUTTONUP:
		payload := &message.ControlButton{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlButtonUp(session, payload)
		})
	case event.CONTROL_KEYPRESS:
		payload := &message.ControlKey{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlKeyPress(session, payload)
		})
	case event.CONTROL_KEYDOWN:
		payload := &message.ControlKey{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlKeyDown(session, payload)
		})
	case event.CONTROL_KEYUP:
		payload := &message.ControlKey{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlKeyUp(session, payload)
		})
	// touch
	case event.CONTROL_TOUCHBEGIN:
		payload := &message.ControlTouch{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlTouchBegin(session, payload)
		})
	case event.CONTROL_TOUCHUPDATE:
		payload := &message.ControlTouch{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlTouchUpdate(session, payload)
		})
	case event.CONTROL_TOUCHEND:
		payload := &message.ControlTouch{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlTouchEnd(session, payload)
		})
	// actions
	case event.CONTROL_CUT:
		err = h.controlCut(session)
	case event.CONTROL_COPY:
		err = h.controlCopy(session)
	case event.CONTROL_PASTE:
		payload := &message.ClipboardData{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.controlPaste(session, payload)
		})
	case event.CONTROL_SELECT_ALL:
		err = h.controlSelectAll(session)

	// Screen Events
	case event.SCREEN_SET:
		payload := &message.ScreenSize{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.screenSet(session, payload)
		})

	// Clipboard Events
	case event.CLIPBOARD_SET:
		payload := &message.ClipboardData{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.clipboardSet(session, payload)
		})

	// Keyboard Events
	case event.KEYBOARD_MAP:
		payload := &message.KeyboardMap{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.keyboardMap(session, payload)
		})
	case event.KEYBOARD_MODIFIERS:
		payload := &message.KeyboardModifiers{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.keyboardModifiers(session, payload)
		})

	// Send Events
	case event.SEND_UNICAST:
		payload := &message.SendUnicast{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.sendUnicast(session, payload)
		})
	case event.SEND_BROADCAST:
		payload := &message.SendBroadcast{}
		err = utils.Unmarshal(payload, data.Payload, func() error {
			return h.sendBroadcast(session, payload)
		})
	default:
		return false
	}

	if err != nil {
		h.logger.Warn().Err(err).
			Str("event", data.Event).
			Str("session_id", session.ID()).
			Msg("message handler has failed")
	}

	return true
}
