package handler

import (
	"encoding/json"

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

func (h *MessageHandlerCtx) Message(session types.Session, raw []byte) bool {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	header := message.Message{}
	if err := json.Unmarshal(raw, &header); err != nil {
		logger.Error().Err(err).Msg("message parsing has failed")
		return false
	}

	var err error
	switch header.Event {
	// Signal Events
	case event.SIGNAL_REQUEST:
		payload := &message.SignalVideo{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.signalRequest(session, payload)
		})
	case event.SIGNAL_RESTART:
		err = h.signalRestart(session)
	case event.SIGNAL_ANSWER:
		payload := &message.SignalAnswer{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.signalAnswer(session, payload)
		})
	case event.SIGNAL_CANDIDATE:
		payload := &message.SignalCandidate{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.signalCandidate(session, payload)
		})
	case event.SIGNAL_VIDEO:
		payload := &message.SignalVideo{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.signalVideo(session, payload)
		})

	// Control Events
	case event.CONTROL_RELEASE:
		err = h.controlRelease(session)
	case event.CONTROL_REQUEST:
		err = h.controlRequest(session)

	// Screen Events
	case event.SCREEN_SET:
		payload := &message.ScreenSize{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.screenSet(session, payload)
		})

	// Clipboard Events
	case event.CLIPBOARD_SET:
		payload := &message.ClipboardData{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.clipboardSet(session, payload)
		})

	// Keyboard Events
	case event.KEYBOARD_MAP:
		payload := &message.KeyboardMap{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.keyboardMap(session, payload)
		})
	case event.KEYBOARD_MODIFIERS:
		payload := &message.KeyboardModifiers{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.keyboardModifiers(session, payload)
		})

	// Send Events
	case event.SEND_UNICAST:
		payload := &message.SendUnicast{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.sendUnicast(session, payload)
		})
	case event.SEND_BROADCAST:
		payload := &message.SendBroadcast{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.sendBroadcast(session, payload)
		})
	default:
		logger.Warn().Str("event", header.Event).Msg("unknown message event")
		return false
	}

	if err != nil {
		logger.Error().Err(err).Str("event", header.Event).Msg("message handler has failed")
	}

	return true
}
