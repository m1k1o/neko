package control

import (
	"errors"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

var (
	ErrNotAllowed  = errors.New("session is not allowed to control")
	ErrNotHost     = errors.New("session is not the control holder")
	ErrAlreadyHost = errors.New("session already holds control")
	ErrUnavailable = errors.New("control lease is unavailable")
)

type RequestResult struct {
	Granted bool
	Queued  bool
}

type Status struct {
	HasHost bool
	HostID  string
	Epoch   uint64
}

// Service is the application boundary for control ownership. REST and
// WebSocket adapters use the same permission, queue notification, and reset
// behavior instead of duplicating it in transport handlers.
type Service struct {
	sessions types.SessionManager
	desktop  types.DesktopManager
}

func NewService(sessions types.SessionManager, desktop types.DesktopManager) *Service {
	return &Service{sessions: sessions, desktop: desktop}
}

func (s *Service) Status() Status {
	host, ok := s.sessions.GetHost()
	status := Status{HasHost: ok, Epoch: s.sessions.ControlEpoch()}
	if ok {
		status.HostID = host.ID()
	}
	return status
}

func (s *Service) Request(session types.Session) (RequestResult, error) {
	if !session.Profile().CanHost || session.PrivateModeEnabled() {
		return RequestResult{}, ErrNotAllowed
	}
	if session.IsHost() {
		return RequestResult{}, ErrAlreadyHost
	}
	if s.sessions.Settings().LockedControls && !session.Profile().IsAdmin {
		return RequestResult{}, ErrNotAllowed
	}

	if s.sessions.Settings().ImplicitHosting {
		session.SetAsHost()
		return RequestResult{Granted: true}, nil
	}

	granted, queued := s.sessions.RequestControl(session)
	if granted {
		return RequestResult{Granted: true}, nil
	}
	if !queued {
		return RequestResult{}, ErrUnavailable
	}

	host, ok := s.sessions.GetHost()
	if !ok {
		return RequestResult{}, ErrUnavailable
	}
	host.Send(event.CONTROL_REQUEST, message.SessionID{ID: session.ID()})
	return RequestResult{Queued: true}, nil
}

func (s *Service) Release(session types.Session) error {
	if !session.Profile().CanHost || session.PrivateModeEnabled() {
		return ErrNotAllowed
	}
	if !session.IsHost() {
		return ErrNotHost
	}

	s.desktop.ResetKeys()
	return s.sessions.ReleaseControl(session)
}

// Renew keeps an active control lease alive while the holder remains
// connected but temporarily idle. The epoch is supplied by the client so a
// stale session cannot renew a lease that has already changed owners.
func (s *Service) Renew(session types.Session, epoch uint64) error {
	if !session.Profile().CanHost || session.PrivateModeEnabled() {
		return ErrNotAllowed
	}

	if err := s.sessions.RenewControl(session, epoch); errors.Is(err, ErrNotHolder) {
		return ErrNotHost
	} else {
		return err
	}
}

// Take grants control to the administrator session and invalidates any
// queued requests. Authorization is enforced by the transport route.
func (s *Service) Take(session types.Session) {
	session.SetAsHost()
}

// Give transfers control to target after applying the target's host policy.
// Authorization for the caller is enforced by the transport route.
func (s *Service) Give(session, target types.Session) error {
	if !target.Profile().CanHost {
		return ErrNotAllowed
	}

	target.SetAsHostBy(session)
	return nil
}

// Reset force-releases control and clears pressed desktop keys. The session
// is kept as the event actor so all adapters expose the same audit semantics.
func (s *Service) Reset(session types.Session) {
	if _, hasHost := s.sessions.GetHost(); !hasHost {
		return
	}

	s.desktop.ResetKeys()
	session.ClearHost()
}

// Disconnect clears desktop input state before a control holder leaves. The
// session manager remains responsible for the rest of session lifecycle.
func (s *Service) Disconnect(session types.Session) {
	if !session.IsHost() {
		return
	}
	s.desktop.ResetKeys()
	session.ClearHost()
}
