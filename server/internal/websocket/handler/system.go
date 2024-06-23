package handler

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
)

func (h *MessageHandlerCtx) systemInit(session types.Session) error {
	host, hasHost := h.sessions.GetHost()

	var hostID string
	if hasHost {
		hostID = host.ID()
	}

	controlHost := message.ControlHost{
		HasHost: hasHost,
		HostID:  hostID,
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

	session.Send(
		event.SYSTEM_INIT,
		message.SystemInit{
			SessionId:         session.ID(),
			ControlHost:       controlHost,
			ScreenSize:        h.desktop.GetScreenSize(),
			Sessions:          sessions,
			Settings:          h.sessions.Settings(),
			TouchEvents:       h.desktop.HasTouchSupport(),
			ScreencastEnabled: h.capture.Screencast().Enabled(),
			WebRTC: message.SystemWebRTC{
				Videos: h.capture.Video().IDs(),
			},
		})

	return nil
}

func (h *MessageHandlerCtx) systemAdmin(session types.Session) error {
	configurations := h.desktop.ScreenConfigurations()

	list := make([]types.ScreenSize, 0, len(configurations))
	for _, conf := range configurations {
		list = append(list, types.ScreenSize{
			Width:  conf.Width,
			Height: conf.Height,
			Rate:   conf.Rate,
		})
	}

	broadcast := h.capture.Broadcast()
	session.Send(
		event.SYSTEM_ADMIN,
		message.SystemAdmin{
			ScreenSizesList: list, // TODO: remove
			BroadcastStatus: message.BroadcastStatus{
				IsActive: broadcast.Started(),
				URL:      broadcast.Url(),
			},
		})

	return nil
}

func (h *MessageHandlerCtx) systemLogs(session types.Session, payload *message.SystemLogs) error {
	for _, msg := range *payload {
		level, _ := zerolog.ParseLevel(msg.Level)

		if level < zerolog.DebugLevel || level > zerolog.ErrorLevel {
			level = zerolog.NoLevel
		}

		// do not use handler logger context
		log.WithLevel(level).
			Fields(msg.Fields).
			Str("module", "client").
			Str("session_id", session.ID()).
			Msg(msg.Message)
	}

	return nil
}
