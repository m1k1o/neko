package handler

import (
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
)

func (h *MessageHandler) boradcastCreate(session types.Session, payload *message.BroadcastCreate) error {
	broadcast := h.capture.Broadcast()

	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if payload.URL == "" {
		return session.Send(
			message.SystemMessage{
				Event:   event.SYSTEM_ERROR,
				Title:   "Error while starting broadcast",
				Message: "missing broadcast URL",
			})
	}

	if broadcast.Started() {
		return session.Send(
			message.SystemMessage{
				Event:   event.SYSTEM_ERROR,
				Title:   "Error while starting broadcast",
				Message: "server is already broadcasting",
			})
	}

	if err := broadcast.Start(payload.URL); err != nil {
		if err := session.Send(
			message.SystemMessage{
				Event:   event.SYSTEM_ERROR,
				Title:   "Error while starting broadcast",
				Message: err.Error(),
			}); err != nil {
			h.logger.Warn().Err(err).Msgf("sending event %s has failed", event.SYSTEM_ERROR)
			return err
		}
	}

	if err := h.boradcastStatus(nil); err != nil {
		return err
	}

	return nil
}

func (h *MessageHandler) boradcastDestroy(session types.Session) error {
	broadcast := h.capture.Broadcast()

	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if !broadcast.Started() {
		return session.Send(
			message.SystemMessage{
				Event:   event.SYSTEM_ERROR,
				Title:   "Error while stopping broadcast",
				Message: "server is not broadcasting",
			})
	}

	broadcast.Stop()

	if err := h.boradcastStatus(nil); err != nil {
		return err
	}

	return nil
}

func (h *MessageHandler) boradcastStatus(session types.Session) error {
	broadcast := h.capture.Broadcast()

	msg := message.BroadcastStatus{
		Event:    event.BORADCAST_STATUS,
		IsActive: broadcast.Started(),
		URL:      broadcast.Url(),
	}

	// if no session, broadcast change
	if session == nil {
		if err := h.sessions.AdminBroadcast(msg, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.BORADCAST_STATUS)
			return err
		}

		return nil
	}

	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if err := session.Send(msg); err != nil {
		h.logger.Warn().Err(err).Msgf("sending event %s has failed", event.BORADCAST_STATUS)
		return err
	}

	return nil
}
