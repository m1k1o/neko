package handler

import (
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
)

func (h *MessageHandler) SessionCreated(id string, session types.Session) error {
	// send sdp and id over to client
	if err := h.signalProvide(id, session); err != nil {
		return err
	}

	// send initialization information
	if err := session.Send(message.SystemInit{
		Event:           event.SYSTEM_INIT,
		ImplicitHosting: h.webrtc.ImplicitControl(),
		Locks:           h.state.AllLocked(),
		FileTransfer:    h.state.FileTransferEnabled(),
	}); err != nil {
		h.logger.Warn().Str("id", id).Err(err).Msgf("sending event %s has failed", event.SYSTEM_INIT)
		return err
	}

	if session.Admin() {
		// send screen configurations if admin
		if err := h.screenConfigurations(id, session); err != nil {
			return err
		}

		// send broadcast status if admin
		if err := h.boradcastStatus(session); err != nil {
			return err
		}
	}

	// send file list if file transfer is enabled
	if h.state.FileTransferEnabled() && (session.Admin() || !h.state.IsLocked("file_transfer")) {
		if err := h.FileTransferRefresh(session); err != nil {
			return err
		}
	}

	return nil
}

func (h *MessageHandler) SessionConnected(id string, session types.Session) error {
	// send list of members to session
	if err := session.Send(message.MembersList{
		Event:   event.MEMBER_LIST,
		Members: h.sessions.Members(),
	}); err != nil {
		h.logger.Warn().Str("id", id).Err(err).Msgf("sending event %s has failed", event.MEMBER_LIST)
		return err
	}

	// send screen current resolution
	if err := h.screenResolution(id, session); err != nil {
		return err
	}

	// tell session there is a host
	host, ok := h.sessions.GetHost()
	if ok {
		if err := session.Send(message.Control{
			Event: event.CONTROL_LOCKED,
			ID:    host.ID(),
		}); err != nil {
			h.logger.Warn().Str("id", id).Err(err).Msgf("sending event %s has failed", event.CONTROL_LOCKED)
			return err
		}
	}

	// let everyone know there is a new session
	if err := h.sessions.Broadcast(
		message.Member{
			Event:  event.MEMBER_CONNECTED,
			Member: session.Member(),
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_RELEASE)
		return err
	}

	return nil
}

func (h *MessageHandler) SessionDestroyed(id string) error {
	// clear host if exists
	if h.sessions.IsHost(id) {
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
