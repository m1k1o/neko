package broadcast

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func ScreenConfiguration(session types.SessionManager, id string, width int, height int, rate int) error {
	return session.Broadcast(message.ScreenResolution{
		Event:  event.SCREEN_RESOLUTION,
		ID:     id,
		Width:  width,
		Height: height,
		Rate:   rate,
	}, nil)
}
