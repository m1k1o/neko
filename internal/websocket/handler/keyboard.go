package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) keyboardMap(session types.Session, payload *message.KeyboardMap) error {
	if !session.IsHost() {
		h.logger.Debug().Str("session_id", session.ID()).Msg("is not the host")
		return nil
	}

	return h.desktop.SetKeyboardMap(types.KeyboardMap{
		Layout:  payload.Layout,
		Variant: payload.Variant,
	})
}

func (h *MessageHandlerCtx) keyboardModifiers(session types.Session, payload *message.KeyboardModifiers) error {
	if !session.IsHost() {
		h.logger.Debug().Str("session_id", session.ID()).Msg("is not the host")
		return nil
	}

	h.desktop.SetKeyboardModifiers(types.KeyboardModifiers{
		NumLock:  payload.NumLock,
		CapsLock: payload.CapsLock,
	})
	return nil
}
