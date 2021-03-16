package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) systemInit(session types.Session) error {
	host := h.sessions.GetHost()

	controlHost := message.ControlHost{
		HasHost: host != nil,
	}

	if controlHost.HasHost {
		controlHost.HostID = host.ID()
	}

	size := h.desktop.GetScreenSize()
	if size == nil {
		h.logger.Warn().Msg("could not get screen size")
		return nil
	}

	sessions := map[string]message.SessionData{}
	for _, session := range h.sessions.List() {
		sessionId := session.ID()
		sessions[sessionId] = message.SessionData{
			ID:      sessionId,
			Profile: session.Profile(),
			State:   session.State(),
		}
	}

	return session.Send(
		message.SystemInit{
			Event:       event.SYSTEM_INIT,
			SessionId:   session.ID(),
			ControlHost: controlHost,
			ScreenSize: message.ScreenSize{
				Width:  size.Width,
				Height: size.Height,
				Rate:   size.Rate,
			},
			Sessions:        sessions,
			ImplicitHosting: h.sessions.ImplicitHosting(),
		})
}

func (h *MessageHandlerCtx) systemAdmin(session types.Session) error {
	screenSizesList := []message.ScreenSize{}
	for _, size := range h.desktop.ScreenConfigurations() {
		for _, rate := range size.Rates {
			screenSizesList = append(screenSizesList, message.ScreenSize{
				Width:  size.Width,
				Height: size.Height,
				Rate:   rate,
			})
		}
	}

	broadcast := h.capture.Broadcast()
	return session.Send(
		message.SystemAdmin{
			Event:           event.SYSTEM_ADMIN,
			ScreenSizesList: screenSizesList,
			BroadcastStatus: message.BroadcastStatus{
				IsActive: broadcast.Started(),
				URL:      broadcast.Url(),
			},
		})
}
