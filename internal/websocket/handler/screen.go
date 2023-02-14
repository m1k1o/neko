package handler

import (
	"errors"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
)

func (h *MessageHandlerCtx) screenSet(session types.Session, payload *message.ScreenSize) error {
	if !session.Profile().IsAdmin {
		return errors.New("is not the admin")
	}

	size, err := h.desktop.SetScreenSize(types.ScreenSize{
		Width:  payload.Width,
		Height: payload.Height,
		Rate:   payload.Rate,
	})

	if err != nil {
		return err
	}

	h.sessions.Broadcast(event.SCREEN_UPDATED, message.ScreenSize{
		Width:  size.Width,
		Height: size.Height,
		Rate:   size.Rate,
	})
	return nil
}
