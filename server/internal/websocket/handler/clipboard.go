package handler

import (
	// "errors"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

func (h *MessageHandlerCtx) clipboardSet(session types.Session, payload *message.ClipboardData) error {
	// Disabled: profile check
	// if !session.Profile().CanAccessClipboard {
	// 	return errors.New("cannot access clipboard")
	// }

	// Disabled: is host check
	// if !session.IsHost() {
	// 	return errors.New("is not the host")
	// }

	return h.desktop.ClipboardSetText(types.ClipboardText{
		Text: payload.Text,
		// TODO: Send HTML?
	})
}
