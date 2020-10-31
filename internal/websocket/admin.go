package websocket

import (
	"strings"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandler) adminLock(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if h.locked {
		h.logger.Debug().Msg("server already locked...")
		return nil
	}

	h.locked = true

	if err := h.sessions.Broadcast(
		message.Admin{
			Event: event.ADMIN_LOCK,
			ID:    session.ID(),
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_LOCK)
		return err
	}

	return nil
}

func (h *MessageHandler) adminUnlock(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if !h.locked {
		h.logger.Debug().Msg("server not locked...")
		return nil
	}

	h.locked = false

	if err := h.sessions.Broadcast(
		message.Admin{
			Event: event.ADMIN_UNLOCK,
			ID:    session.ID(),
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_UNLOCK)
		return err
	}

	return nil
}

func (h *MessageHandler) adminControl(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	host := h.sessions.GetHost()
	h.sessions.SetHost(session)

	if host != nil {
		if err := h.sessions.Broadcast(
			message.AdminTarget{
				Event:  event.ADMIN_CONTROL,
				ID:     session.ID(),
				Target: host.ID(),
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_CONTROL)
			return err
		}
	} else {
		if err := h.sessions.Broadcast(
			message.Admin{
				Event: event.ADMIN_CONTROL,
				ID:    session.ID(),
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_CONTROL)
			return err
		}
	}

	return nil
}

func (h *MessageHandler) adminRelease(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	host := h.sessions.GetHost()
	h.sessions.ClearHost()

	if host != nil {
		if err := h.sessions.Broadcast(
			message.AdminTarget{
				Event:  event.ADMIN_RELEASE,
				ID:     session.ID(),
				Target: host.ID(),
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_RELEASE)
			return err
		}
	} else {
		if err := h.sessions.Broadcast(
			message.Admin{
				Event: event.ADMIN_RELEASE,
				ID:    session.ID(),
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_RELEASE)
			return err
		}
	}

	return nil
}

func (h *MessageHandler) adminGive(session types.Session, payload *message.Admin) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		h.logger.Debug().Str("id", target.ID()).Msg("user does not exist")
		return nil
	}

	// set host
	h.sessions.SetHost(target)

	// let everyone know
	if err := h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.CONTROL_GIVE,
			ID:     session.ID(),
			Target: target.ID(),
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_LOCKED)
		return err
	}

	return nil
}

func (h *MessageHandler) adminMute(session types.Session, payload *message.Admin) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		h.logger.Debug().Str("id", payload.ID).Msg("can't find session id")
		return nil
	}

	if target.Admin() {
		h.logger.Debug().Msg("target is an admin, baling")
		return nil
	}

	target.SetMuted(true)

	if err := h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.ADMIN_MUTE,
			Target: target.ID(),
			ID:     session.ID(),
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_UNMUTE)
		return err
	}

	return nil
}

func (h *MessageHandler) adminUnmute(session types.Session, payload *message.Admin) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		h.logger.Debug().Str("id", payload.ID).Msg("can't find target session")
		return nil
	}

	target.SetMuted(false)

	if err := h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.ADMIN_UNMUTE,
			Target: target.ID(),
			ID:     session.ID(),
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_UNMUTE)
		return err
	}

	return nil
}

func (h *MessageHandler) adminKick(session types.Session, payload *message.Admin) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		h.logger.Debug().Str("id", payload.ID).Msg("can't find session id")
		return nil
	}

	if target.Admin() {
		h.logger.Debug().Msg("target is an admin, baling")
		return nil
	}

	if err := target.Disconnect("kicked"); err != nil {
		return err
	}

	if err := h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.ADMIN_KICK,
			Target: target.ID(),
			ID:     session.ID(),
		}, []string{payload.ID}); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_KICK)
		return err
	}

	return nil
}

func (h *MessageHandler) adminBan(session types.Session, payload *message.Admin) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		h.logger.Debug().Str("id", payload.ID).Msg("can't find session id")
		return nil
	}

	if target.Admin() {
		h.logger.Debug().Msg("target is an admin, baling")
		return nil
	}

	remote := target.Address()
	if remote == "" {
		h.logger.Debug().Msg("no remote address, baling")
		return nil
	}

	address := strings.SplitN(remote, ":", -1)
	if len(address[0]) < 1 {
		h.logger.Debug().Str("address", remote).Msg("no remote address, baling")
		return nil
	}

	h.logger.Debug().Str("address", remote).Msg("adding address to banned")

	h.banned[address[0]] = true

	if err := target.Disconnect("banned"); err != nil {
		return err
	}

	if err := h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.ADMIN_BAN,
			Target: target.ID(),
			ID:     session.ID(),
		}, []string{payload.ID}); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_BAN)
		return err
	}

	return nil
}
