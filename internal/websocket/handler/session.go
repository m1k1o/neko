package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) SessionCreated(session types.Session) error {
	// send sdp and id over to client
	if err := h.signalProvide(session); err != nil {
		return err
	}

	if session.Admin() {
		// send screen configurations if admin
		if err := h.screenConfigurations(session); err != nil {
			return err
		}

		// send broadcast status if admin
		if err := h.boradcastStatus(session); err != nil {
			return err
		}
	}

	return nil
}

func (h *MessageHandlerCtx) SessionConnected(session types.Session) error {
	// TODO: Refactor.
	members := []*message.MembersListEntry{}
	for _, session := range h.sessions.Members() {
		members = append(members, &message.MembersListEntry{
			ID:    session.ID(),
			Name:  session.Name(),
			Admin: session.Admin(),
			Muted: session.Muted(),
		})
	}

	// send list of members to session
	if err := session.Send(message.MembersList{
		Event:    event.MEMBER_LIST,
		Memebers: members,
	}); err != nil {
		h.logger.Warn().Str("id", session.ID()).Err(err).Msgf("sending event %s has failed", event.MEMBER_LIST)
		return err
	}

	// send screen current resolution
	if err := h.screenResolution(session); err != nil {
		return err
	}

	// tell session there is a host
	host := h.sessions.GetHost()
	if host != nil {
		if err := session.Send(message.Control{
			Event: event.CONTROL_LOCKED,
			ID:    host.ID(),
		}); err != nil {
			h.logger.Warn().Str("id", session.ID()).Err(err).Msgf("sending event %s has failed", event.CONTROL_LOCKED)
			return err
		}
	}

	// let everyone know there is a new session
	if err := h.sessions.Broadcast(
		message.Member{
			Event:  event.MEMBER_CONNECTED,
			Member: &message.MembersListEntry{
				ID:    session.ID(),
				Name:  session.Name(),
				Admin: session.Admin(),
				Muted: session.Muted(),
			},
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_RELEASE)
		return err
	}

	return nil
}

func (h *MessageHandlerCtx) SessionDestroyed(id string) error {
	// clear host if exists
	host := h.sessions.GetHost()
	if host != nil && host.ID() == id {
		h.sessions.ClearHost()
		if err := h.sessions.Broadcast(message.Control{
			Event: event.CONTROL_RELEASE,
			ID:    id,
		}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_RELEASE)
		}
	}

	// let everyone know session disconnected
	if err := h.sessions.Broadcast(
		message.MemberDisconnected{
			Event: event.MEMBER_DISCONNECTED,
			ID:    id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.MEMBER_DISCONNECTED)
		return err
	}

	return nil
}
