package filetransfer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/demodesk/neko/pkg/auth"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const MULTIPART_FORM_MAX_MEMORY = 32 << 20

func NewManager(
	sessions types.SessionManager,
	config *Config,
) *Manager {
	logger := log.With().Str("module", "filetransfer").Logger()

	return &Manager{
		logger:   logger,
		config:   config,
		sessions: sessions,
		shutdown: make(chan struct{}),
	}
}

type Manager struct {
	logger   zerolog.Logger
	config   *Config
	sessions types.SessionManager
	shutdown chan struct{}
	mu       sync.RWMutex
	fileList []Item
}

func (m *Manager) isEnabledForSession(session types.Session) (bool, error) {
	settings := Settings{
		Enabled: true, // defaults to true
	}
	err := m.sessions.Settings().Plugins.Unmarshal(PluginName, &settings)
	if err != nil && !errors.Is(err, types.ErrPluginSettingsNotFound) {
		return false, fmt.Errorf("unable to unmarshal %s plugin settings from global settings: %w", PluginName, err)
	}

	profile := Settings{
		Enabled: true, // defaults to true
	}

	err = session.Profile().Plugins.Unmarshal(PluginName, &profile)
	if err != nil && !errors.Is(err, types.ErrPluginSettingsNotFound) {
		return false, fmt.Errorf("unable to unmarshal %s plugin settings from profile: %w", PluginName, err)
	}

	return m.config.Enabled && (settings.Enabled || session.Profile().IsAdmin) && profile.Enabled, nil
}

func (m *Manager) refresh() (error, bool) {
	// if file transfer is disabled, return immediately without refreshing
	if !m.config.Enabled {
		return nil, false
	}

	files, err := ListFiles(m.config.RootDir)
	if err != nil {
		return err, false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// check if file list has changed (todo: use hash instead of comparing all fields)
	changed := false
	if len(files) == len(m.fileList) {
		for i, file := range files {
			if file.Name != m.fileList[i].Name || file.Size != m.fileList[i].Size {
				changed = true
				break
			}
		}
	} else {
		changed = true
	}

	m.fileList = files
	return nil, changed
}

func (m *Manager) broadcastUpdate() {
	m.mu.RLock()
	fileList := m.fileList
	m.mu.RUnlock()

	m.sessions.Broadcast(FILETRANSFER_UPDATE, Message{
		Enabled: m.config.Enabled,
		RootDir: m.config.RootDir,
		Files:   fileList,
	})
}

func (m *Manager) sendUpdate(session types.Session) {
	m.mu.RLock()
	fileList := m.fileList
	m.mu.RUnlock()

	session.Send(FILETRANSFER_UPDATE, Message{
		Enabled: m.config.Enabled,
		RootDir: m.config.RootDir,
		Files:   fileList,
	})
}

func (m *Manager) Start() error {
	// send init message once a user connects
	m.sessions.OnConnected(func(session types.Session) {
		m.sendUpdate(session)
	})

	// if file transfer is disabled, return immediately without starting the watcher
	if !m.config.Enabled {
		return nil
	}

	if _, err := os.Stat(m.config.RootDir); os.IsNotExist(err) {
		err = os.Mkdir(m.config.RootDir, os.ModePerm)
		m.logger.Err(err).Msg("creating file transfer directory")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("unable to start file transfer dir watcher: %w", err)
	}

	go func() {
		defer watcher.Close()

		// periodically refresh file list
		ticker := time.NewTicker(m.config.RefreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-m.shutdown:
				m.logger.Info().Msg("shutting down file transfer manager")
				return
			case <-ticker.C:
				err, changed := m.refresh()
				if err != nil {
					m.logger.Err(err).Msg("unable to refresh file transfer list")
				}
				if changed {
					m.broadcastUpdate()
				}
			case e, ok := <-watcher.Events:
				if !ok {
					m.logger.Info().Msg("file transfer dir watcher closed")
					return
				}

				if e.Has(fsnotify.Create) || e.Has(fsnotify.Remove) || e.Has(fsnotify.Rename) {
					m.logger.Debug().Str("event", e.String()).Msg("file transfer dir watcher event")

					err, changed := m.refresh()
					if err != nil {
						m.logger.Err(err).Msg("unable to refresh file transfer list")
					}

					if changed {
						m.broadcastUpdate()
					}
				}
			case err := <-watcher.Errors:
				m.logger.Err(err).Msg("error in file transfer dir watcher")
			}
		}
	}()

	if err := watcher.Add(m.config.RootDir); err != nil {
		return fmt.Errorf("unable to watch file transfer dir: %w", err)
	}

	// initial refresh
	err, changed := m.refresh()
	if err != nil {
		return fmt.Errorf("unable to refresh file transfer list: %w", err)
	}
	if changed {
		m.broadcastUpdate()
	}

	return nil
}

func (m *Manager) Shutdown() error {
	close(m.shutdown)
	return nil
}

func (m *Manager) Route(r types.Router) {
	r.With(auth.AdminsOnly).Get("/", m.downloadFileHandler)
	r.With(auth.AdminsOnly).Post("/", m.uploadFileHandler)
}

func (m *Manager) WebSocketHandler(session types.Session, msg types.WebSocketMessage) bool {
	switch msg.Event {
	case FILETRANSFER_UPDATE:
		err, changed := m.refresh()
		if err != nil {
			m.logger.Err(err).Msg("unable to refresh file transfer list")
		}

		if changed {
			// broadcast update message to all clients
			m.broadcastUpdate()
		} else {
			// send update message to this client only
			m.sendUpdate(session)
		}
		return true
	}

	// not handled by this plugin
	return false
}

func (m *Manager) downloadFileHandler(w http.ResponseWriter, r *http.Request) error {
	session, ok := auth.GetSession(r)
	if !ok {
		return utils.HttpUnauthorized("session not found")
	}

	enabled, err := m.isEnabledForSession(session)
	if err != nil {
		return utils.HttpInternalServerError().
			WithInternalErr(err).
			Msg("error checking file transfer permissions")
	}

	if !enabled {
		return utils.HttpForbidden("file transfer is disabled")
	}

	filename := r.URL.Query().Get("filename")
	badChars, err := regexp.MatchString(`(?m)\.\.(?:\/|$)`, filename)
	if filename == "" || badChars || err != nil {
		return utils.HttpBadRequest().
			WithInternalErr(err).
			Msg("bad filename")
	}

	// ensure filename is clean and only contains the basename
	filename = filepath.Clean(filename)
	filename = filepath.Base(filename)
	filePath := filepath.Join(m.config.RootDir, filename)

	http.ServeFile(w, r, filePath)
	return nil
}

func (m *Manager) uploadFileHandler(w http.ResponseWriter, r *http.Request) error {
	session, ok := auth.GetSession(r)
	if !ok {
		return utils.HttpUnauthorized("session not found")
	}

	enabled, err := m.isEnabledForSession(session)
	if err != nil {
		return utils.HttpInternalServerError().
			WithInternalErr(err).
			Msg("error checking file transfer permissions")
	}

	if !enabled {
		return utils.HttpForbidden("file transfer is disabled")
	}

	err = r.ParseMultipartForm(MULTIPART_FORM_MAX_MEMORY)
	if err != nil || r.MultipartForm == nil {
		return utils.HttpBadRequest().
			WithInternalErr(err).
			Msg("error parsing form")
	}

	defer func() {
		err = r.MultipartForm.RemoveAll()
		if err != nil {
			m.logger.Warn().Err(err).Msg("failed to clean up multipart form")
		}
	}()

	for _, formheader := range r.MultipartForm.File["files"] {
		// ensure filename is clean and only contains the basename
		filename := filepath.Clean(formheader.Filename)
		filename = filepath.Base(filename)
		filePath := filepath.Join(m.config.RootDir, filename)

		formfile, err := formheader.Open()
		if err != nil {
			return utils.HttpBadRequest().
				WithInternalErr(err).
				Msg("error opening formdata file")
		}
		defer formfile.Close()

		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return utils.HttpInternalServerError().
				WithInternalErr(err).
				Msg("error opening file for writing")
		}
		defer f.Close()

		_, err = io.Copy(f, formfile)
		if err != nil {
			return utils.HttpInternalServerError().
				WithInternalErr(err).
				Msg("error writing file")
		}
	}

	return nil
}
