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
	}
}

type MessageHandlerCtx struct {
	logger    zerolog.Logger
	sessions  types.SessionManager
	webrtc    types.WebRTCManager
	desktop   types.DesktopManager
	capture   types.CaptureManager
}

func (h *MessageHandlerCtx) Message(session types.Session, raw []byte) error {
	header := message.Message{}
	if err := json.Unmarshal(raw, &header); err != nil {
		return err
	}

	var err error
	switch header.Event {
	// Signal Events
	case event.SIGNAL_REQUEST:
		err = h.signalRequest(session)
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
	case event.KEYBOARD_MODIFIERS:
		payload := &message.KeyboardLayout{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.keyboardLayout(session, payload)
		})
	case event.KEYBOARD_LAYOUT:
		payload := &message.KeyboardModifiers{}
		err = utils.Unmarshal(payload, raw, func() error {
			return h.keyboardModifiers(session, payload)
		})
	default:
		return errors.Errorf("unknown message event %s", header.Event)
	}

	return errors.Wrapf(err, "%s failed", header.Event)
}
