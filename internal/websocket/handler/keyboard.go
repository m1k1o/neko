package handler

import (
	"errors"

	"gitlab.com/demodesk/neko/server/internal/types"
	"gitlab.com/demodesk/neko/server/internal/types/message"
)

func (h *MessageHandlerCtx) keyboardMap(session types.Session, payload *message.KeyboardMap) error {
	if !session.IsHost() {
		return errors.New("is not the host")
	}

	return h.desktop.SetKeyboardMap(types.KeyboardMap{
		Layout:  payload.Layout,
		Variant: payload.Variant,
	})
}

func (h *MessageHandlerCtx) keyboardModifiers(session types.Session, payload *message.KeyboardModifiers) error {
	if !session.IsHost() {
		return errors.New("is not the host")
	}

	h.desktop.SetKeyboardModifiers(types.KeyboardModifiers{
		NumLock:  payload.NumLock,
		CapsLock: payload.CapsLock,
	})

	return nil
}
