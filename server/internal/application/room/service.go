package room

import (
	"github.com/m1k1o/neko/server/internal/control"
	"github.com/m1k1o/neko/server/pkg/types"
)

type SessionSnapshot struct {
	ID      string
	Profile types.MemberProfile
	State   types.SessionState
}

// Snapshot is the transport-neutral room state needed when a client joins.
type Snapshot struct {
	SessionID         string
	HostID            string
	HasHost           bool
	ControlEpoch      uint64
	Sessions          []SessionSnapshot
	ScreenSize        types.ScreenSize
	Settings          types.Settings
	TouchEvents       bool
	ScreencastEnabled bool
	VideoIDs          []string
}

type Service struct {
	sessions types.SessionManager
	desktop  types.DesktopManager
	capture  types.CaptureManager
	control  *control.Service
}

func NewService(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	controlService *control.Service,
) *Service {
	return &Service{sessions: sessions, desktop: desktop, capture: capture, control: controlService}
}

func (s *Service) Snapshot(session types.Session) Snapshot {
	status := s.control.Status()

	sessions := make([]SessionSnapshot, 0)
	for _, current := range s.sessions.List() {
		sessions = append(sessions, SessionSnapshot{
			ID:      current.ID(),
			Profile: current.Profile(),
			State:   current.State(),
		})
	}

	return Snapshot{
		SessionID:         session.ID(),
		HostID:            status.HostID,
		HasHost:           status.HasHost,
		ControlEpoch:      status.Epoch,
		Sessions:          sessions,
		ScreenSize:        s.desktop.GetScreenSize(),
		Settings:          s.sessions.Settings(),
		TouchEvents:       s.desktop.HasTouchSupport(),
		ScreencastEnabled: s.capture.Screencast().Enabled(),
		VideoIDs:          s.capture.Video().IDs(),
	}
}
