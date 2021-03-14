package member

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/member/dummy"
	"demodesk/neko/internal/member/file"
	"demodesk/neko/internal/member/object"
	"demodesk/neko/internal/types"
)

func New(sessions types.SessionManager, config *config.Member) *MemberManagerCtx {
	manager := &MemberManagerCtx{
		logger:   log.With().Str("module", "member").Logger(),
		sessions: sessions,
		config:   config,
	}

	switch config.Provider {
	case "file":
		manager.provider = file.New(file.Config{
			File: config.FilePath,
		})
	case "object":
		manager.provider = object.New(object.Config{
			AdminPassword: config.AdminPassword,
			UserPassword:  config.Password,
		})
	case "dummy":
		fallthrough
	default:
		manager.provider = dummy.New()
	}
	
	return manager
}

type MemberManagerCtx struct {
	logger    zerolog.Logger
	sessions  types.SessionManager
	config    *config.Member
	mu        sync.Mutex
	provider  types.MemberProvider
}

func (manager *MemberManagerCtx) Connect() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.provider.Connect()
}

func (manager *MemberManagerCtx) Disconnect() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.provider.Disconnect()
}

func (manager *MemberManagerCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.provider.Authenticate(username, password)
}

func (manager *MemberManagerCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.provider.Insert(username, password, profile)
}

func (manager *MemberManagerCtx) Select(id string) (types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.provider.Select(id)
}

func (manager *MemberManagerCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.provider.SelectAll(limit, offset)
}

func (manager *MemberManagerCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// update corresponding session, if exists
	if _, ok := manager.sessions.Get(id); ok {
		if err := manager.sessions.Update(id, profile); err != nil {
			manager.logger.Err(err).Msg("error while updating session")
		}
	}

	return manager.provider.UpdateProfile(id, profile)
}

func (manager *MemberManagerCtx) UpdatePassword(id string, password string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.provider.UpdatePassword(id, password)
}

func (manager *MemberManagerCtx) Delete(id string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// destroy corresponding session, if exists
	if _, ok := manager.sessions.Get(id); ok {
		if err := manager.sessions.Delete(id); err != nil {
			manager.logger.Err(err).Msg("error while deleting session")
		}
	}

	return manager.provider.Delete(id)
}

//
// member -> session
//

func (manager *MemberManagerCtx) Login(username string, password string) (types.Session, string, error) {
	id, profile, err := manager.provider.Authenticate(username, password)
	if err != nil {
		return nil, "", err
	}

	return manager.sessions.Create(id, profile)
}

func (manager *MemberManagerCtx) Logout(id string) error {
	if _, ok := manager.sessions.Get(id); !ok {
		return fmt.Errorf("session not found")
	}

	return manager.sessions.Delete(id)
}
