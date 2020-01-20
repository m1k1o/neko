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

	h.sessions.SetHost(id)

	if err := h.sessions.Brodcast(
		message.Admin{
			Event: event.ADMIN_FORCE_CONTROL,
			ID:    id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_FORCE_CONTROL)
		return err
	}

	return nil
}

func (h *MessageHandler) adminRelease(id string, session *session.Session) error {
	if !session.Admin {
		return nil
	}

	h.sessions.ClearHost()

	if err := h.sessions.Brodcast(
		message.Admin{
			Event: event.ADMIN_FORCE_RELEASE,
			ID:    id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_FORCE_RELEASE)
		return err
	}

	return nil
}

func (h *MessageHandler) adminBan(id string, session *session.Session, payload *message.Admin) error {
	if !session.Admin {
		return nil
	}

	session, ok := h.sessions.Get(id)
	if !ok {
		return nil
	}

	address := session.RemoteAddr()
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
		message.AdminSubject{
			Event:   event.ADMIN_BAN,
			Subject: payload.ID,
			ID:      id,
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

	if err := h.sessions.Kick(payload.ID, message.Disconnect{
		Event:   event.SYSTEM_DISCONNECT,
		Message: "You have been banned",
	}); err != nil {
		return err
	}

	if err := h.sessions.Brodcast(
		message.AdminSubject{
			Event:   event.ADMIN_KICK,
			Subject: payload.ID,
			ID:      id,
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

	if err := h.sessions.Mute(payload.ID); err != nil {
		return err
	}

	if err := h.sessions.Brodcast(
		message.AdminSubject{
			Event:   event.ADMIN_MUTE,
			Subject: payload.ID,
			ID:      id,
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

	if err := h.sessions.Unmute(payload.ID); err != nil {
		return err
	}

	if err := h.sessions.Brodcast(
		message.AdminSubject{
			Event:   event.ADMIN_UNMUTE,
			Subject: payload.ID,
			ID:      id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.ADMIN_UNMUTE)
		return err
	}

	return nil
}
