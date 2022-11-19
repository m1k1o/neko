package handler

import (
	"strings"

	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
)

func (h *MessageHandler) adminLock(id string, session types.Session, payload *message.AdminLock) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if h.state.IsLocked(payload.Resource) {
		h.logger.Debug().Str("resource", payload.Resource).Msg("resource already locked...")
		return nil
	}

	// allow only known resources
	switch payload.Resource {
	case "login":
	case "control":
	case "file_transfer":
	default:
		h.logger.Debug().Msg("unknown lock resource")
		return nil
	}

	// TODO: Handle locks in sessions as flags.
	if payload.Resource == "control" {
		h.sessions.SetControlLocked(true)
	}

	h.state.Lock(payload.Resource, id)

	if err := h.sessions.Broadcast(
		message.AdminLock{
			Event:    event.ADMIN_LOCK,
			ID:       id,
			Resource: payload.Resource,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_LOCK)
		return err
	}

	return nil
}

func (h *MessageHandler) adminUnlock(id string, session types.Session, payload *message.AdminLock) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if !h.state.IsLocked(payload.Resource) {
		h.logger.Debug().Str("resource", payload.Resource).Msg("resource not locked...")
		return nil
	}

	// TODO: Handle locks in sessions as flags.
	if payload.Resource == "control" {
		h.sessions.SetControlLocked(false)
	}

	h.state.Unlock(payload.Resource)

	if err := h.sessions.Broadcast(
		message.AdminLock{
			Event:    event.ADMIN_UNLOCK,
			ID:       id,
			Resource: payload.Resource,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_UNLOCK)
		return err
	}

	return nil
}

func (h *MessageHandler) adminControl(id string, session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	host, ok := h.sessions.GetHost()

	err := h.sessions.SetHost(id)
	if err != nil {
		return err
	}

	if ok {
		if err := h.sessions.Broadcast(
			message.AdminTarget{
				Event:  event.ADMIN_CONTROL,
				ID:     id,
				Target: host.ID(),
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_CONTROL)
			return err
		}
	} else {
		if err := h.sessions.Broadcast(
			message.Admin{
				Event: event.ADMIN_CONTROL,
				ID:    id,
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_CONTROL)
			return err
		}
	}

	return nil
}

func (h *MessageHandler) AdminRelease(id string, session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	host, ok := h.sessions.GetHost()

	h.sessions.ClearHost()

	if ok {
		if err := h.sessions.Broadcast(
			message.AdminTarget{
				Event:  event.ADMIN_RELEASE,
				ID:     id,
				Target: host.ID(),
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_RELEASE)
			return err
		}
	} else {
		if err := h.sessions.Broadcast(
			message.Admin{
				Event: event.ADMIN_RELEASE,
				ID:    id,
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_RELEASE)
			return err
		}
	}

	return nil
}

func (h *MessageHandler) adminGive(id string, session types.Session, payload *message.Admin) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if !h.sessions.Has(payload.ID) {
		h.logger.Debug().Str("id", payload.ID).Msg("user does not exist")
		return nil
	}

	// set host
	err := h.sessions.SetHost(payload.ID)
	if err != nil {
		return err
	}

	// let everyone know
	if err := h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.CONTROL_GIVE,
			ID:     id,
			Target: payload.ID,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_LOCKED)
		return err
	}

	return nil
}

func (h *MessageHandler) adminMute(id string, session types.Session, payload *message.Admin) error {
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
			ID:     id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_UNMUTE)
		return err
	}

	return nil
}

func (h *MessageHandler) adminUnmute(id string, session types.Session, payload *message.Admin) error {
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
			ID:     id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_UNMUTE)
		return err
	}

	return nil
}

func (h *MessageHandler) adminKick(id string, session types.Session, payload *message.Admin) error {
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

	if err := target.Kick("kicked"); err != nil {
		return err
	}

	if err := h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.ADMIN_KICK,
			Target: target.ID(),
			ID:     id,
		}, []string{payload.ID}); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_KICK)
		return err
	}

	return nil
}

func (h *MessageHandler) adminBan(id string, session types.Session, payload *message.Admin) error {
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
	h.state.Ban(address[0], id)

	if err := target.Kick("banned"); err != nil {
		return err
	}

	if err := h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.ADMIN_BAN,
			Target: target.ID(),
			ID:     id,
		}, []string{payload.ID}); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.ADMIN_BAN)
		return err
	}

	return nil
}
