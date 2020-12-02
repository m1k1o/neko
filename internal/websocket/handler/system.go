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
		h.logger.Debug().Msg("could not get screen size")
		return nil
	}

	members := map[string]message.MemberData{}
	for _, session := range h.sessions.Members() {
		members[session.ID()] = message.MemberData{
			ID:      session.ID(),
			Name:    session.Name(),
			IsAdmin: session.IsAdmin(),
		}
	}

	return session.Send(
		message.SystemInit{
			Event:           event.SYSTEM_INIT,
			MemberId:        session.ID(),
			ControlHost:     controlHost,
			ScreenSize:      message.ScreenSize{
				Width:  size.Width,
				Height: size.Height,
				Rate:   int(size.Rate),
			},
			Members:         members,
			ImplicitHosting: h.sessions.ImplicitHosting(),
		})
}

func (h *MessageHandlerCtx) systemAdmin(session types.Session) error {
	screenSizesList := []message.ScreenSize{}
	for _, size := range h.desktop.ScreenConfigurations() {
		for _, fps := range size.Rates {
			screenSizesList = append(screenSizesList, message.ScreenSize{
				Width:  size.Width,
				Height: size.Height,
				Rate:   int(fps),
			})
		}
	}

	return session.Send(
		message.SystemAdmin{
			Event:           event.SYSTEM_ADMIN,
			ScreenSizesList: screenSizesList,
			BroadcastStatus: message.BroadcastStatus{
				IsActive: h.capture.BroadcastEnabled(),
				URL:      h.capture.BroadcastUrl(),
			},
		})
}
