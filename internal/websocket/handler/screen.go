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

	data := types.ScreenSize(*payload)
	if err := h.desktop.SetScreenSize(data); err != nil {
		return err
	}

	h.sessions.Broadcast(event.SCREEN_UPDATED, payload, nil)
	return nil
}
