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

	data := types.ScreenSize(*payload)
	if err := h.desktop.SetScreenSize(data); err != nil {
		return err
	}

	h.sessions.Broadcast(event.SCREEN_UPDATED, payload, nil)
	return nil
}
