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

	h.desktop.SetKeyboardLayout(payload.Layout)
	return nil
}

func (h *MessageHandlerCtx) keyboardModifiers(session types.Session, payload *message.KeyboardModifiers) error {
	if !session.IsHost() {
		h.logger.Debug().Str("id", session.ID()).Msg("is not the host")
		return nil
	}

	var NumLock = 0
	if payload.NumLock == nil {
		NumLock = -1
	} else if *payload.NumLock {
		NumLock = 1
	}

	var CapsLock = 0
	if payload.CapsLock == nil {
		CapsLock = -1
	} else if *payload.CapsLock {
		CapsLock = 1
	}

	var ScrollLock = 0
	if payload.ScrollLock == nil {
		ScrollLock = -1
	} else if *payload.ScrollLock {
		ScrollLock = 1
	}

	h.desktop.SetKeyboardModifiers(NumLock, CapsLock, ScrollLock)
	return nil
}
