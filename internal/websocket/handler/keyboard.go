package handler

import (
	"errors"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/message"
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
