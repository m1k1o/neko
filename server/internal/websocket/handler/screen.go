package handler

import (
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
)

func (h *MessageHandler) screenSet(id string, session types.Session, payload *message.ScreenResolution) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if err := h.desktop.SetScreenSize(types.ScreenSize{
		Width:  payload.Width,
		Height: payload.Height,
		Rate:   payload.Rate,
	}); err != nil {
		h.logger.Warn().Err(err).Msgf("unable to change screen size")
		return err
	}

	if err := h.sessions.Broadcast(
		message.ScreenResolution{
			Event:  event.SCREEN_RESOLUTION,
			ID:     id,
			Width:  payload.Width,
			Height: payload.Height,
			Rate:   payload.Rate,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.SCREEN_RESOLUTION)
		return err
	}

	return nil
}

func (h *MessageHandler) screenResolution(id string, session types.Session) error {
	if size := h.desktop.GetScreenSize(); size != nil {
		if err := session.Send(message.ScreenResolution{
			Event:  event.SCREEN_RESOLUTION,
			Width:  size.Width,
			Height: size.Height,
			Rate:   size.Rate,
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
		Configurations: h.desktop.ScreenConfigurations(),
	}); err != nil {
		h.logger.Warn().Err(err).Msgf("sending event %s has failed", event.SCREEN_CONFIGURATIONS)
		return err
	}

	return nil
}
