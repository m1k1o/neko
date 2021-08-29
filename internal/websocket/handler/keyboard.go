package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) keyboardMap(session types.Session, payload *message.KeyboardMap) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.IsHost() {
		logger.Debug().Msg("is not the host")
		return nil
	}

	return h.desktop.SetKeyboardMap(types.KeyboardMap{
		Layout:  payload.Layout,
		Variant: payload.Variant,
	})
}

func (h *MessageHandlerCtx) keyboardModifiers(session types.Session, payload *message.KeyboardModifiers) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.IsHost() {
		logger.Debug().Msg("is not the host")
		return nil
	}

	h.desktop.SetKeyboardModifiers(types.KeyboardModifiers{
		NumLock:  payload.NumLock,
		CapsLock: payload.CapsLock,
	})

	return nil
}
