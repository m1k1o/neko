package member

import (
	"errors"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/internal/member/file"
	"github.com/m1k1o/neko/server/internal/member/multiuser"
	"github.com/m1k1o/neko/server/internal/member/noauth"
	oauthprovider "github.com/m1k1o/neko/server/internal/member/oauth"
	"github.com/m1k1o/neko/server/internal/member/object"
	"github.com/m1k1o/neko/server/pkg/types"
)

func New(sessions types.SessionManager, config *config.Member) *MemberManagerCtx {
	manager := &MemberManagerCtx{
		logger:     log.With().Str("module", "member").Logger(),
		sessions:   sessions,
		config:     config,
		oauthExtra: make(map[string]map[string]any),
	}

	switch config.Provider {
	case "file":
		manager.provider = file.New(config.File)
	case "object":
		manager.provider = object.New(config.Object)
	case "multiuser":
		manager.provider = multiuser.New(config.Multiuser)
	case "oauth":
		manager.provider = oauthprovider.New()
	case "noauth":
		fallthrough
	default:
		manager.provider = noauth.New()
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
	oauthExtra map[string]map[string]any
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

	// get primarily from corresponding session, if exists
	session, ok := manager.sessions.Get(id)
	if ok {
		return session.Profile(), nil
	}

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
	defer func() {
		manager.loginMu.Lock()
		defer manager.loginMu.Unlock()
		delete(manager.oauthExtra, id)
	}()

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

	if !profile.IsAdmin && manager.sessions.Settings().LockedLogins {
		return nil, "", types.ErrSessionLoginsLocked
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

// LoginOAuth creates a stable session identity from an OAuth provider subject.
// The profile data is refreshed on every successful OAuth login.
func (manager *MemberManagerCtx) LoginOAuth(subject, name, avatar, email string, isAdmin bool, extraData map[string]any) (types.Session, string, error) {
	manager.loginMu.Lock()
	defer manager.loginMu.Unlock()

	if subject == "" {
		return nil, "", errors.New("OAuth subject is empty")
	}

	profile := manager.config.OAuth.UserProfile
	if name != "" {
		profile.Name = name
	}
	profile.Avatar = avatar
	if isOAuthAdministrator(email, isAdmin, manager.config.OAuth.AdminEmails) {
		profile.IsAdmin = true
		profile.CanSeeInactiveCursors = true
	}

	if !profile.IsAdmin && manager.sessions.Settings().LockedLogins {
		return nil, "", types.ErrSessionLoginsLocked
	}

	id := "oauth:" + subject
	if session, ok := manager.sessions.Get(id); ok {
		if session.State().IsConnected {
			return nil, "", types.ErrSessionAlreadyConnected
		}

		if err := manager.sessions.Delete(id); err != nil {
			return nil, "", err
		}
	}

	session, token, err := manager.sessions.Create(id, profile)
	if err == nil {
		manager.oauthExtra[id] = cloneOAuthExtraData(extraData)
	}
	return session, token, err
}

func isOAuthAdministrator(email string, isAdmin bool, adminEmails []string) bool {
	if isAdmin {
		return true
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	for _, adminEmail := range adminEmails {
		if strings.EqualFold(email, strings.TrimSpace(adminEmail)) {
			return true
		}
	}
	return false
}

func (manager *MemberManagerCtx) OAuthExtraData(id string) map[string]any {
	manager.loginMu.Lock()
	defer manager.loginMu.Unlock()

	return cloneOAuthExtraData(manager.oauthExtra[id])
}

func cloneOAuthExtraData(extraData map[string]any) map[string]any {
	if len(extraData) == 0 {
		return nil
	}
	copy := make(map[string]any, len(extraData))
	for key, value := range extraData {
		copy[key] = value
	}
	return copy
}

func (manager *MemberManagerCtx) Logout(id string) error {
	manager.loginMu.Lock()
	defer manager.loginMu.Unlock()

	err := manager.sessions.Delete(id)
	if err == nil || errors.Is(err, types.ErrSessionNotFound) {
		delete(manager.oauthExtra, id)
	}
	return err
}
