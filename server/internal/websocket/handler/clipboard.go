package handler

import (
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

func (h *MessageHandlerCtx) clipboardSet(session types.Session, payload *message.ClipboardData) error {
	return h.desktopApp.SetClipboard(session, types.ClipboardText{
		Text: payload.Text,
		// TODO: Send HTML?
	})
}
