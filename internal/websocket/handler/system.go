package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
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
			Profile: session.GetProfile(),
			State:   session.GetState(),
		}
	}

	var cursorImage *message.CursorImage
	cur := h.desktop.GetCursorImage()
	uri, err := utils.GetCursorImageURI(cur)
	if err == nil {
		cursorImage = &message.CursorImage{
			Event:  event.CURSOR_IMAGE,
			Uri:    uri,
			Width:  cur.Width,
			Height: cur.Height,
			X:      cur.Xhot,
			Y:      cur.Yhot,
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
				Rate:   size.Rate,
			},
			Members:         members,
			ImplicitHosting: h.sessions.ImplicitHosting(),
			CursorImage:     cursorImage,
		})
}

func (h *MessageHandlerCtx) systemAdmin(session types.Session) error {
	screenSizesList := []message.ScreenSize{}
	for _, size := range h.desktop.ScreenConfigurations() {
		for _, fps := range size.Rates {
			screenSizesList = append(screenSizesList, message.ScreenSize{
				Width:  size.Width,
				Height: size.Height,
				Rate:   fps,
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
