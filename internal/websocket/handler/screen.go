package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) screenSet(session types.Session, payload *message.ScreenResolution) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if err := h.desktop.ChangeScreenSize(payload.Width, payload.Height, payload.Rate); err != nil {
		h.logger.Warn().Err(err).Msgf("unable to change screen size")
		return nil
	}

	return h.sessions.Broadcast(
		message.ScreenResolution{
			Event:  event.SCREEN_RESOLUTION,
			ID:     session.ID(),
			Width:  payload.Width,
			Height: payload.Height,
			Rate:   payload.Rate,
		}, nil)
}

func (h *MessageHandlerCtx) screenResolution(session types.Session) error {
	size := h.desktop.GetScreenSize()
	if size == nil {
		h.logger.Debug().Msg("could not get screen size")
		return nil
	}

	return session.Send(
		message.ScreenResolution{
			Event:  event.SCREEN_RESOLUTION,
			Width:  size.Width,
			Height: size.Height,
			Rate:   int(size.Rate),
		})
}

func (h *MessageHandlerCtx) screenConfigurations(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	return session.Send(
		message.ScreenConfigurations{
			Event:          event.SCREEN_CONFIGURATIONS,
			Configurations: h.desktop.ScreenConfigurations(),
		})
}
