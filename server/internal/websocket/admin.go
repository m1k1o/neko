package websocket

import (
	"n.eko.moe/neko/internal/event"
	"n.eko.moe/neko/internal/message"
	"n.eko.moe/neko/internal/session"
)

func (h *MessageHandler) adminLock(id string, session *session.Session) error {
	if !session.Admin || !h.locked {
		return nil
	}

	h.locked = true

	if err := h.sessions.Brodcast(
		message.Admin{
			Event: event.ADMIN_LOCK,
			ID:    id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_LOCK)
		return err
	}

	return nil
}

func (h *MessageHandler) adminUnlock(id string, session *session.Session) error {
	if !session.Admin || !h.locked {
		return nil
	}

	h.locked = false

	if err := h.sessions.Brodcast(
		message.Admin{
			Event: event.ADMIN_UNLOCK,
			ID:    id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_UNLOCK)
		return err
	}

	return nil
}

func (h *MessageHandler) adminControl(id string, session *session.Session) error {
	if !session.Admin {
		return nil
	}

	host, ok := h.sessions.GetHost()

	h.sessions.SetHost(id)

	if ok {
		if err := h.sessions.Brodcast(
			message.AdminTarget{
				Event:  event.ADMIN_CONTROL,
				ID:     id,
				Target: host.ID,
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_CONTROL)
			return err
		}
	} else {
		if err := h.sessions.Brodcast(
			message.Admin{
				Event: event.ADMIN_CONTROL,
				ID:    id,
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_CONTROL)
			return err
		}
	}

	return nil
}

func (h *MessageHandler) adminRelease(id string, session *session.Session) error {
	if !session.Admin {
		return nil
	}

	host, ok := h.sessions.GetHost()

	h.sessions.ClearHost()

	if ok {
		if err := h.sessions.Brodcast(
			message.AdminTarget{
				Event:  event.ADMIN_RELEASE,
				ID:     id,
				Target: host.ID,
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_RELEASE)
			return err
		}
	} else {
		if err := h.sessions.Brodcast(
			message.Admin{
				Event: event.ADMIN_RELEASE,
				ID:    id,
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_RELEASE)
			return err
		}
	}

	return nil
}

func (h *MessageHandler) adminBan(id string, session *session.Session, payload *message.Admin) error {
	if !session.Admin {
		return nil
	}

	target, ok := h.sessions.Get(id)
	if !ok {
		return nil
	}

	if target.Admin {
		return nil
	}

	address := target.RemoteAddr()
	if address == nil {
		return nil
	}

	h.banned[*address] = true

	if err := session.Kick(message.Disconnect{
		Event:   event.SYSTEM_DISCONNECT,
		Message: "You have been banned",
	}); err != nil {
		return err
	}

	if err := h.sessions.Brodcast(
		message.AdminTarget{
			Event:  event.ADMIN_BAN,
			Target: target.ID,
			ID:     id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_BAN)
		return err
	}

	return nil
}

func (h *MessageHandler) adminKick(id string, session *session.Session, payload *message.Admin) error {
	if !session.Admin {
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		return nil
	}

	if target.Admin {
		return nil
	}

	if err := target.Kick(message.Disconnect{
		Event:   event.SYSTEM_DISCONNECT,
		Message: "You have been kicked",
	}); err != nil {
		return err
	}

	if err := h.sessions.Brodcast(
		message.AdminTarget{
			Event:  event.ADMIN_KICK,
			Target: target.ID,
			ID:     id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_KICK)
		return err
	}

	return nil
}

func (h *MessageHandler) adminMute(id string, session *session.Session, payload *message.Admin) error {
	if !session.Admin {
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		return nil
	}

	if target.Admin {
		return nil
	}

	target.Muted = true

	if err := h.sessions.Brodcast(
		message.AdminTarget{
			Event:  event.ADMIN_MUTE,
			Target: target.ID,
			ID:     id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_UNMUTE)
		return err
	}

	return nil
}

func (h *MessageHandler) adminUnmute(id string, session *session.Session, payload *message.Admin) error {
	if !session.Admin {
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		return nil
	}

	target.Muted = false

	if err := h.sessions.Brodcast(
		message.AdminTarget{
			Event:  event.ADMIN_UNMUTE,
			Target: target.ID,
			ID:     id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_UNMUTE)
		return err
	}

	return nil
}
