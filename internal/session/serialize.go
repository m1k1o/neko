package session

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/demodesk/neko/pkg/types"
)

func (manager *SessionManagerCtx) save() {
	if manager.config.File == "" {
		return
	}

	// serialize sessions
	sessions := make([]types.SessionProfile, 0, len(manager.sessions))
	for _, session := range manager.sessions {
		sessions = append(sessions, types.SessionProfile{
			Id:      session.id,
			Token:   session.token,
			Profile: session.profile,
		})
	}

	// convert to json
	data, err := json.Marshal(sessions)
	if err != nil {
		manager.logger.Error().Err(err).Msg("failed to marshal sessions")
		return
	}

	// write to file
	err = os.WriteFile(manager.config.File, data, 0644)
	if err != nil {
		manager.logger.Error().Err(err).
			Str("file", manager.config.File).
			Msg("failed to write sessions to a file")
	}
}

func (manager *SessionManagerCtx) load() {
	if manager.config.File == "" {
		return
	}

	// read file
	data, err := os.ReadFile(manager.config.File)
	if err != nil {
		// if file does not exist
		if errors.Is(err, os.ErrNotExist) {
			manager.logger.Info().
				Str("file", manager.config.File).
				Msg("sessions file does not exist")
			return
		}
		manager.logger.Error().Err(err).
			Str("file", manager.config.File).
			Msg("failed to read sessions from a file")
		return
	}

	// if file is empty
	if len(data) == 0 {
		manager.logger.Info().
			Str("file", manager.config.File).
			Msg("sessions file is empty")
		return
	}

	// deserialize sessions
	sessions := make([]types.SessionProfile, 0)
	err = json.Unmarshal(data, &sessions)
	if err != nil {
		manager.logger.Error().Err(err).Msg("failed to unmarshal sessions")
		return
	}

	// create sessions
	manager.sessionsMu.Lock()
	for _, session := range sessions {
		manager.tokens[session.Token] = session.Id
		manager.sessions[session.Id] = &SessionCtx{
			id:      session.Id,
			token:   session.Token,
			manager: manager,
			logger:  manager.logger.With().Str("session_id", session.Id).Logger(),
			profile: session.Profile,
		}
	}
	manager.sessionsMu.Unlock()

	manager.logger.Info().
		Int("sessions", len(sessions)).
		Str("file", manager.config.File).
		Msg("loaded sessions from a file")
}
