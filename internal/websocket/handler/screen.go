package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) screenSet(session types.Session, payload *message.ScreenSize) error {
	if !session.IsAdmin() {
		h.logger.Debug().Msg("member not admin")
		return nil
	}

	if err := h.desktop.ChangeScreenSize(payload.Width, payload.Height, payload.Rate); err != nil {
		return err
	}

	h.sessions.Broadcast(
		message.ScreenSize{
			Event:  event.SCREEN_UPDATED,
			Width:  payload.Width,
			Height: payload.Height,
			Rate:   payload.Rate,
		}, nil)

	return nil
}
