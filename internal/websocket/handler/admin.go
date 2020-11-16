package handler

import (
	"strings"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) adminLock(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if h.locked {
		h.logger.Debug().Msg("server already locked...")
		return nil
	}

	h.locked = true

	return h.sessions.Broadcast(
		message.Admin{
			Event: event.ADMIN_LOCK,
			ID:    session.ID(),
		}, nil)
}

func (h *MessageHandlerCtx) adminUnlock(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if !h.locked {
		h.logger.Debug().Msg("server not locked...")
		return nil
	}

	h.locked = false

	return h.sessions.Broadcast(
		message.Admin{
			Event: event.ADMIN_UNLOCK,
			ID:    session.ID(),
		}, nil)
}

func (h *MessageHandlerCtx) adminControl(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	host := h.sessions.GetHost()
	h.sessions.SetHost(session)

	if host != nil {
		return h.sessions.Broadcast(
			message.AdminTarget{
				Event:  event.ADMIN_CONTROL,
				ID:     session.ID(),
				Target: host.ID(),
			}, nil)
	}

	return h.sessions.Broadcast(
		message.Admin{
			Event: event.ADMIN_CONTROL,
			ID:    session.ID(),
		}, nil)
}

func (h *MessageHandlerCtx) adminRelease(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	host := h.sessions.GetHost()
	h.sessions.ClearHost()

	if host != nil {
		return h.sessions.Broadcast(
			message.AdminTarget{
				Event:  event.ADMIN_RELEASE,
				ID:     session.ID(),
				Target: host.ID(),
			}, nil)
	}

	return h.sessions.Broadcast(
		message.Admin{
			Event: event.ADMIN_RELEASE,
			ID:    session.ID(),
		}, nil)
}

func (h *MessageHandlerCtx) adminGive(session types.Session, payload *message.Admin) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		h.logger.Debug().Str("id", payload.ID).Msg("can't find target session")
		return nil
	}

	h.sessions.SetHost(target)

	return h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.CONTROL_GIVE,
			ID:     session.ID(),
			Target: target.ID(),
		}, nil)
}

func (h *MessageHandlerCtx) adminKick(session types.Session, payload *message.Admin) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		h.logger.Debug().Str("id", payload.ID).Msg("can't find target session")
		return nil
	}

	if target.Admin() {
		h.logger.Debug().Msg("target is an admin, baling")
		return nil
	}

	if err := target.Disconnect("kicked"); err != nil {
		return err
	}

	return h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.ADMIN_KICK,
			Target: target.ID(),
			ID:     session.ID(),
		}, []string{payload.ID})
}

func (h *MessageHandlerCtx) adminBan(session types.Session, payload *message.Admin) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		h.logger.Debug().Str("id", payload.ID).Msg("can't find target session")
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

	return h.sessions.Broadcast(
		message.AdminTarget{
			Event:  event.ADMIN_BAN,
			Target: target.ID(),
			ID:     session.ID(),
		}, []string{payload.ID})
}
