package handler

import (
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
)

func (h *MessageHandlerCtx) SessionCreated(session types.Session) error {
	h.sessions.Broadcast(
		event.SESSION_CREATED,
		message.SessionData{
			ID:      session.ID(),
			Profile: session.Profile(),
			State:   session.State(),
		})

	return nil
}

func (h *MessageHandlerCtx) SessionDeleted(session types.Session) error {
	h.sessions.Broadcast(
		event.SESSION_DELETED,
		message.SessionID{
			ID: session.ID(),
		})

	return nil
}

func (h *MessageHandlerCtx) SessionConnected(session types.Session) error {
	if err := h.systemInit(session); err != nil {
		return err
	}

	if session.Profile().IsAdmin {
		if err := h.systemAdmin(session); err != nil {
			return err
		}

		// update settings in atomic way
		h.sessions.UpdateSettingsFunc(session, func(settings *types.Settings) bool {
			// if control protection & locked controls: unlock controls
			if settings.LockedControls && settings.ControlProtection {
				settings.LockedControls = false
				return true // update settings
			}
			return false // do not update settings
		})
	}

	return h.SessionStateChanged(session)
}

func (h *MessageHandlerCtx) SessionDisconnected(session types.Session) error {
	// clear host if exists
	if session.IsHost() {
		h.desktop.ResetKeys()
		session.ClearHost()
	}

	if session.Profile().IsAdmin {
		hasAdmin := false
		h.sessions.Range(func(s types.Session) bool {
			if s.Profile().IsAdmin && s.ID() != session.ID() && s.State().IsConnected {
				hasAdmin = true
				return false
			}
			return true
		})

		// update settings in atomic way
		h.sessions.UpdateSettingsFunc(session, func(settings *types.Settings) bool {
			// if control protection & not locked controls & no admin: lock controls
			if !settings.LockedControls && settings.ControlProtection && !hasAdmin {
				settings.LockedControls = true
				return true // update settings
			}
			return false // do not update settings
		})
	}

	return h.SessionStateChanged(session)
}

func (h *MessageHandlerCtx) SessionProfileChanged(session types.Session, new, old types.MemberProfile) error {
	h.sessions.Broadcast(
		event.SESSION_PROFILE,
		message.MemberProfile{
			ID:            session.ID(),
			MemberProfile: new,
		})

	return nil
}

func (h *MessageHandlerCtx) SessionStateChanged(session types.Session) error {
	h.sessions.Broadcast(
		event.SESSION_STATE,
		message.SessionState{
			ID:           session.ID(),
			SessionState: session.State(),
		})

	return nil
}
