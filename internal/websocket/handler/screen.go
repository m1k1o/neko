package handler

import (
	"errors"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) screenSet(session types.Session, payload *message.ScreenSize) error {
	if !session.Profile().IsAdmin {
		return errors.New("is not the admin")
	}

	if err := h.desktop.SetScreenSize(types.ScreenSize{
		Width:  payload.Width,
		Height: payload.Height,
		Rate:   payload.Rate,
	}); err != nil {
		return err
	}

	h.sessions.Broadcast(
		event.SCREEN_UPDATED,
		message.ScreenSize{
			Width:  payload.Width,
			Height: payload.Height,
			Rate:   payload.Rate,
		}, nil)

	return nil
}
