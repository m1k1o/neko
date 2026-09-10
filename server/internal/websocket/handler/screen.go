package handler

import (
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

func (h *MessageHandlerCtx) screenSet(session types.Session, payload *message.ScreenSize) error {
	_, err := h.desktopApp.SetScreenSize(session, payload.ScreenSize)
	return err
}
