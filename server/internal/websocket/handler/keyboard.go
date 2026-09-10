package handler

import (
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

func (h *MessageHandlerCtx) keyboardMap(session types.Session, payload *message.KeyboardMap) error {
	return h.desktopApp.SetKeyboardMap(session, payload.KeyboardMap)
}

func (h *MessageHandlerCtx) keyboardModifiers(session types.Session, payload *message.KeyboardModifiers) error {
	return h.desktopApp.SetKeyboardModifiers(session, payload.KeyboardModifiers)
}
