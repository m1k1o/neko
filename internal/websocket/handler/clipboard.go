package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) clipboardSet(session types.Session, payload *message.ClipboardData) error {
	if !session.IsHost() {
		h.logger.Debug().Str("id", session.ID()).Msg("is not the host")
		return nil
	}

	h.desktop.WriteClipboard(payload.Text)
	return nil
}
