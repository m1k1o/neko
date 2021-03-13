package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) clipboardSet(session types.Session, payload *message.ClipboardData) error {
	if !session.Profile().CanAccessClipboard {
		h.logger.Debug().Str("session_id", session.ID()).Msg("cannot access clipboard")
		return nil
	}

	if !session.IsHost() {
		h.logger.Debug().Str("session_id", session.ID()).Msg("is not the host")
		return nil
	}

	return h.desktop.ClipboardSetText(types.ClipboardText{
		Text: payload.Text,
		// TODO: Send HTML?
	})
}
