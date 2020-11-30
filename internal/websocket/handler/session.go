package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) SessionConnected(session types.Session) error {
	// create member list
	members := []*message.MembersListEntry{}
	for _, session := range h.sessions.Members() {
		members = append(members, &message.MembersListEntry{
			ID:    session.ID(),
			Name:  session.Name(),
			Admin: session.Admin(),
		})
	}

	// send list of members to session
	if err := session.Send(
		message.MembersList{
			Event:    event.MEMBER_LIST,
			Memebers: members,
		}); err != nil {
		return err
	}

	// tell session there is a host
	host := h.sessions.GetHost()
	if host != nil {
		if err := session.Send(
			message.ControlHost{
				Event:   event.CONTROL_HOST,
				HasHost: true,
				HostID:  host.ID(),
			}); err != nil {
			return err
		}
	}

	// let everyone know there is a new session
	h.sessions.Broadcast(
		message.Member{
			Event:  event.MEMBER_CONNECTED,
			ID:    session.ID(),
			Name:  session.Name(),
			Admin: session.Admin(),
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) SessionDisconnected(session types.Session) error {
	// clear host if exists
	if session.IsHost() {
		h.sessions.ClearHost()

		h.sessions.Broadcast(
			message.ControlHost{
				Event:   event.CONTROL_HOST,
				HasHost: false,
			}, nil)
	}

	// let everyone know session disconnected
	h.sessions.Broadcast(
		message.MemberDisconnected{
			Event: event.MEMBER_DISCONNECTED,
			ID:    session.ID(),
		}, nil);

	return nil
}
