package member

import (
	"errors"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/internal/config"
	"github.com/demodesk/neko/internal/member/dummy"
	"github.com/demodesk/neko/internal/member/file"
	"github.com/demodesk/neko/internal/member/object"
	"github.com/demodesk/neko/pkg/types"
)

func New(sessions types.SessionManager, config *config.Member) *MemberManagerCtx {
	manager := &MemberManagerCtx{
		logger:   log.With().Str("module", "member").Logger(),
		sessions: sessions,
		config:   config,
	}

	switch config.Provider {
	case "file":
		manager.provider = file.New(config.File)
	case "object":
		manager.provider = object.New(config.Object)
	case "dummy":
		fallthrough
	default:
		manager.provider = dummy.New()
	}

	return manager
}

type MemberManagerCtx struct {
	logger     zerolog.Logger
	sessions   types.SessionManager
	config     *config.Member
	providerMu sync.Mutex
	provider   types.MemberProvider
	loginMu    sync.Mutex
}

func (manager *MemberManagerCtx) Connect() error {
	manager.providerMu.Lock()
	defer manager.providerMu.Unlock()

	return manager.provider.Connect()
}

func (manager *MemberManagerCtx) Disconnect() error {
	manager.providerMu.Lock()
	defer manager.providerMu.Unlock()

	return manager.provider.Disconnect()
}

func (manager *MemberManagerCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	manager.providerMu.Lock()
	defer manager.providerMu.Unlock()

	return manager.provider.Authenticate(username, password)
}

func (manager *MemberManagerCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	manager.providerMu.Lock()
	defer manager.providerMu.Unlock()

	return manager.provider.Insert(username, password, profile)
}

func (manager *MemberManagerCtx) Select(id string) (types.MemberProfile, error) {
	manager.providerMu.Lock()
	defer manager.providerMu.Unlock()

	return manager.provider.Select(id)
}

func (manager *MemberManagerCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	manager.providerMu.Lock()
	defer manager.providerMu.Unlock()

	return manager.provider.SelectAll(limit, offset)
}

func (manager *MemberManagerCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	manager.providerMu.Lock()
	defer manager.providerMu.Unlock()

	// update corresponding session, if exists
	err := manager.sessions.Update(id, profile)
	if err != nil && !errors.Is(err, types.ErrSessionNotFound) {
		manager.logger.Err(err).Msg("error while updating session")
	}

	return manager.provider.UpdateProfile(id, profile)
}

func (manager *MemberManagerCtx) UpdatePassword(id string, password string) error {
	manager.providerMu.Lock()
	defer manager.providerMu.Unlock()

	return manager.provider.UpdatePassword(id, password)
}

func (manager *MemberManagerCtx) Delete(id string) error {
	manager.providerMu.Lock()
	defer manager.providerMu.Unlock()

	// destroy corresponding session, if exists
	err := manager.sessions.Delete(id)
	if err != nil && !errors.Is(err, types.ErrSessionNotFound) {
		manager.logger.Err(err).Msg("error while deleting session")
	}

	return manager.provider.Delete(id)
}

//
// member -> session
//

func (manager *MemberManagerCtx) Login(username string, password string) (types.Session, string, error) {
	manager.loginMu.Lock()
	defer manager.loginMu.Unlock()

	id, profile, err := manager.provider.Authenticate(username, password)
	if err != nil {
		return nil, "", err
	}

	session, ok := manager.sessions.Get(id)
	if ok {
		if session.State().IsConnected {
			return nil, "", types.ErrSessionAlreadyConnected
		}

		// TODO: Replace session.
		if err := manager.sessions.Delete(id); err != nil {
			return nil, "", err
		}
	}

	return manager.sessions.Create(id, profile)
}

func (manager *MemberManagerCtx) Logout(id string) error {
	manager.loginMu.Lock()
	defer manager.loginMu.Unlock()

	return manager.sessions.Delete(id)
}
