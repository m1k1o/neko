package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/message"
)


func (h *MessageHandlerCtx) keyboardLayout(session types.Session, payload *message.KeyboardLayout) error {
	if !session.IsHost() {
		h.logger.Debug().Str("id", session.ID()).Msg("is not the host")
		return nil
	}

	return h.desktop.SetKeyboardLayout(payload.Layout)
}

func (h *MessageHandlerCtx) keyboardModifiers(session types.Session, payload *message.KeyboardModifiers) error {
	if !session.IsHost() {
		h.logger.Debug().Str("id", session.ID()).Msg("is not the host")
		return nil
	}

	h.desktop.SetKeyboardModifiers(types.KeyboardModifiers{
		NumLock: payload.NumLock,
		CapsLock: payload.CapsLock,
	})
	return nil
}
