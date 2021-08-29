package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) screenSet(session types.Session, payload *message.ScreenSize) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.Profile().IsAdmin {
		logger.Debug().Msg("is not the admin")
		return nil
	}

	if err := h.desktop.SetScreenSize(types.ScreenSize{
		Width:  payload.Width,
		Height: payload.Height,
		Rate:   payload.Rate,
	}); err != nil {
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
