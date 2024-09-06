package handler

import (
	"errors"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/message"
)

func (h *MessageHandlerCtx) keyboardMap(session types.Session, payload *message.KeyboardMap) error {
	if !session.IsHost() {
		return errors.New("is not the host")
	}

	return h.desktop.SetKeyboardMap(payload.KeyboardMap)
}

func (h *MessageHandlerCtx) keyboardModifiers(session types.Session, payload *message.KeyboardModifiers) error {
	if !session.IsHost() {
		return errors.New("is not the host")
	}

	h.desktop.SetKeyboardModifiers(payload.KeyboardModifiers)
	return nil
}
