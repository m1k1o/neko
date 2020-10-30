package websocket

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/websocket/broadcast"
)

func (h *MessageHandler) screenSet(id string, session types.Session, payload *message.ScreenResolution) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if err := h.remote.ChangeResolution(payload.Width, payload.Height, payload.Rate); err != nil {
		h.logger.Warn().Err(err).Msgf("unable to change screen size")
		return err
	}

	if err := broadcast.ScreenConfiguration(h.sessions, id, payload.Width, payload.Height, payload.Rate); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.SCREEN_RESOLUTION)
		return err
	}

	return nil
}

func (h *MessageHandler) screenResolution(id string, session types.Session) error {
	if size := h.remote.GetScreenSize(); size != nil {
		if err := session.Send(message.ScreenResolution{
			Event:  event.SCREEN_RESOLUTION,
			Width:  size.Width,
			Height: size.Height,
			Rate:   int(size.Rate),
		}); err != nil {
			h.logger.Warn().Err(err).Msgf("sending event %s has failed", event.SCREEN_RESOLUTION)
			return err
		}
	}

	return nil
}

func (h *MessageHandler) screenConfigurations(id string, session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if err := session.Send(message.ScreenConfigurations{
		Event:          event.SCREEN_CONFIGURATIONS,
		Configurations: h.remote.ScreenConfigurations(),
	}); err != nil {
		h.logger.Warn().Err(err).Msgf("sending event %s has failed", event.SCREEN_CONFIGURATIONS)
		return err
	}

	return nil
}
