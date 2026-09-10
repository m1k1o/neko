package handler

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

func (h *MessageHandlerCtx) systemInit(session types.Session) error {
	snapshot := h.room.Snapshot(session)

	sessions := map[string]message.SessionData{}
	for _, current := range snapshot.Sessions {
		sessions[current.ID] = message.SessionData{
			ID:      current.ID,
			Profile: current.Profile,
			State:   current.State,
		}
	}

	session.Send(
		event.SYSTEM_INIT,
		message.SystemInit{
			SessionId: snapshot.SessionID,
			ControlHost: message.ControlHost{
				HasHost: snapshot.HasHost,
				HostID:  snapshot.HostID,
				Epoch:   snapshot.ControlEpoch,
			},
			ScreenSize:        snapshot.ScreenSize,
			Sessions:          sessions,
			Settings:          snapshot.Settings,
			TouchEvents:       snapshot.TouchEvents,
			ScreencastEnabled: snapshot.ScreencastEnabled,
			WebRTC: message.SystemWebRTC{
				Videos: snapshot.VideoIDs,
			},
		})

	return nil
}

func (h *MessageHandlerCtx) systemAdmin(session types.Session) error {
	active, url := h.desktopApp.BroadcastStatus()
	session.Send(
		event.SYSTEM_ADMIN,
		message.SystemAdmin{
			BroadcastStatus: message.BroadcastStatus{
				IsActive: active,
				URL:      url,
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
