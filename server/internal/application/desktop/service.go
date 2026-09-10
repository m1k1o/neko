package desktop

import (
	"errors"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
	"github.com/m1k1o/neko/server/pkg/utils"
)

var (
	ErrNotAdmin           = errors.New("is not the admin")
	ErrNotHost            = errors.New("is not the host")
	ErrClipboardForbidden = errors.New("cannot access clipboard")
)

type Service struct {
	desktop  types.DesktopManager
	sessions types.SessionManager
	capture  types.CaptureManager
}

func NewService(
	desktop types.DesktopManager,
	sessions types.SessionManager,
	capture types.CaptureManager,
) *Service {
	return &Service{desktop: desktop, sessions: sessions, capture: capture}
}

func (s *Service) ScreenSize() types.ScreenSize {
	return s.desktop.GetScreenSize()
}

func (s *Service) ScreenConfigurations() []types.ScreenSize {
	return s.desktop.ScreenConfigurations()
}

func (s *Service) Screenshot(quality int) ([]byte, error) {
	return utils.CreateJPGImage(s.desktop.GetScreenshotImage(), quality)
}

func (s *Service) ScreencastEnabled() bool {
	return s.capture.Screencast().Enabled()
}

func (s *Service) ScreencastImage() ([]byte, error) {
	return s.capture.Screencast().Image()
}

func (s *Service) SetScreenSize(session types.Session, size types.ScreenSize) (types.ScreenSize, error) {
	if !session.Profile().IsAdmin {
		return types.ScreenSize{}, ErrNotAdmin
	}
	result, err := s.desktop.SetScreenSize(size)
	if err != nil {
		return types.ScreenSize{}, err
	}
	s.sessions.Broadcast(event.SCREEN_UPDATED, message.ScreenSizeUpdate{ID: session.ID(), ScreenSize: result})
	return result, nil
}

func (s *Service) SetKeyboardMap(session types.Session, keyboardMap types.KeyboardMap) error {
	if !session.IsHost() {
		return ErrNotHost
	}
	return s.desktop.SetKeyboardMap(keyboardMap)
}

func (s *Service) GetKeyboardMap() (*types.KeyboardMap, error) {
	return s.desktop.GetKeyboardMap()
}

func (s *Service) SetKeyboardModifiers(session types.Session, modifiers types.KeyboardModifiers) error {
	if !session.IsHost() {
		return ErrNotHost
	}
	s.desktop.SetKeyboardModifiers(modifiers)
	return nil
}

func (s *Service) GetKeyboardModifiers() types.KeyboardModifiers {
	return s.desktop.GetKeyboardModifiers()
}

func (s *Service) SetClipboard(session types.Session, clipboard types.ClipboardText) error {
	if !session.Profile().CanAccessClipboard {
		return ErrClipboardForbidden
	}
	if !session.IsHost() {
		return ErrNotHost
	}
	return s.desktop.ClipboardSetText(clipboard)
}

func (s *Service) GetClipboard() (*types.ClipboardText, error) {
	return s.desktop.ClipboardGetText()
}

func (s *Service) ClipboardImage() ([]byte, error) {
	return s.desktop.ClipboardGetBinary("image/png")
}

func (s *Service) ClipboardBinary(mime string) ([]byte, error) {
	return s.desktop.ClipboardGetBinary(mime)
}

func (s *Service) SetClipboardBinary(mime string, data []byte) error {
	return s.desktop.ClipboardSetBinary(mime, data)
}

func (s *Service) ClipboardTargets() ([]string, error) {
	return s.desktop.ClipboardGetTargets()
}

func (s *Service) BroadcastStatus() (bool, string) {
	broadcast := s.capture.Broadcast()
	return broadcast.Started(), broadcast.Url()
}

func (s *Service) StartBroadcast(url string) error {
	return s.capture.Broadcast().Start(url)
}

func (s *Service) StopBroadcast() {
	s.capture.Broadcast().Stop()
}

func (s *Service) BroadcastStatusChanged() {
	active, url := s.BroadcastStatus()
	s.sessions.AdminBroadcast(event.BROADCAST_STATUS, message.BroadcastStatus{IsActive: active, URL: url})
}
