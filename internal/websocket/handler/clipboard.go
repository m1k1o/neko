package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) clipboardSet(session types.Session, payload *message.ClipboardData) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.Profile().CanAccessClipboard {
		logger.Debug().Msg("cannot access clipboard")
		return nil
	}

	if !session.IsHost() {
		logger.Debug().Msg("is not the host")
		return nil
	}

	return h.desktop.ClipboardSetText(types.ClipboardText{
		Text: payload.Text,
		// TODO: Send HTML?
	})
}
