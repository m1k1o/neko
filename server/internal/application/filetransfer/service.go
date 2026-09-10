package filetransfer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/m1k1o/neko/server/pkg/types"
)

var (
	ErrDisabled        = errors.New("file transfer is disabled")
	ErrPermission      = errors.New("file transfer permission denied")
	ErrInvalidFilename = errors.New("bad filename")
)

type Config struct {
	Enabled      bool
	RootDir      string
	UserDownload bool
	UserUpload   bool
	UserDelete   bool
}

type Upload struct {
	Name   string
	Reader io.Reader
}

type Service struct {
	sessions types.SessionManager
	config   Config
}

func NewService(sessions types.SessionManager, config Config) *Service {
	return &Service{sessions: sessions, config: config}
}

func (s *Service) IsEnabledForSession(session types.Session) (bool, error) {
	settings := types.PluginSettings{}
	if err := s.sessions.Settings().Plugins.Unmarshal("filetransfer", &settings); err != nil && !errors.Is(err, types.ErrPluginSettingsNotFound) {
		return false, fmt.Errorf("unable to unmarshal filetransfer settings: %w", err)
	}
	profile := types.PluginSettings{}
	if err := session.Profile().Plugins.Unmarshal("filetransfer", &profile); err != nil && !errors.Is(err, types.ErrPluginSettingsNotFound) {
		return false, fmt.Errorf("unable to unmarshal filetransfer profile settings: %w", err)
	}
	// PluginSettings is intentionally decoded by the caller-specific config
	// in the adapter; this service only applies the global enabled policy.
	return s.config.Enabled && (settingsEnabled(settings) || session.Profile().IsAdmin) && profileEnabled(profile), nil
}

func (s *Service) AuthorizeDownload(session types.Session) error {
	if !s.config.Enabled {
		return ErrDisabled
	}
	if !session.Profile().IsAdmin && !s.config.UserDownload {
		return fmt.Errorf("%w: download is not allowed for non-admin users", ErrPermission)
	}
	return nil
}

func (s *Service) AuthorizeUpload(session types.Session) error {
	if !s.config.Enabled {
		return ErrDisabled
	}
	if !session.Profile().IsAdmin && !s.config.UserUpload {
		return fmt.Errorf("%w: upload is not allowed for non-admin users", ErrPermission)
	}
	return nil
}

func (s *Service) AuthorizeDelete(session types.Session) error {
	if !s.config.Enabled {
		return ErrDisabled
	}
	if !session.Profile().IsAdmin && !s.config.UserDelete {
		return fmt.Errorf("%w: delete is not allowed for non-admin users", ErrPermission)
	}
	return nil
}

func (s *Service) Path(filename string) (string, error) {
	badChars, err := regexp.MatchString(`(?m)\.\.(?:/|$)`, filename)
	if err != nil {
		return "", err
	}
	if filename == "" || badChars {
		return "", ErrInvalidFilename
	}
	return filepath.Join(s.config.RootDir, filepath.Base(filepath.Clean(filename))), nil
}

func (s *Service) UploadFiles(session types.Session, files []Upload) error {
	if err := s.AuthorizeUpload(session); err != nil {
		return err
	}
	for _, file := range files {
		path, err := s.Path(file.Name)
		if err != nil {
			return err
		}
		output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(output, file.Reader)
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func (s *Service) Delete(session types.Session, filename string) error {
	if err := s.AuthorizeDelete(session); err != nil {
		return err
	}
	path, err := s.Path(filename)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func settingsEnabled(settings types.PluginSettings) bool {
	return pluginBool(settings, "enabled", true)
}
func profileEnabled(settings types.PluginSettings) bool { return pluginBool(settings, "enabled", true) }

func pluginBool(settings types.PluginSettings, key string, fallback bool) bool {
	value, ok := settings[key]
	if !ok {
		return fallback
	}
	boolean, ok := value.(bool)
	return ok && boolean
}
