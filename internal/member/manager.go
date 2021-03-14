package member

import (
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/member/dummy"
	"demodesk/neko/internal/member/file"
	"demodesk/neko/internal/member/object"
	"demodesk/neko/internal/types"
)

func New(config *config.Member) *MemberManagerCtx {
	manager := &MemberManagerCtx{
		logger: log.With().Str("module", "member").Logger(),
		config: config,
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
	config    *config.Member
	mu        sync.Mutex
	provider  types.MemberProvider
}

func (manager *MemberManagerCtx) Connect() error {
	return manager.provider.Connect()
}

func (manager *MemberManagerCtx) Disconnect() error {
	return manager.provider.Disconnect()
}

func (manager *MemberManagerCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	return manager.provider.Authenticate(username, password)
}

func (manager *MemberManagerCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	return manager.provider.Insert(username, password, profile)
}

func (manager *MemberManagerCtx) Select(id string) (types.MemberProfile, error) {
	return manager.provider.Select(id)
}

func (manager *MemberManagerCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	return manager.provider.SelectAll(limit, offset)
}

func (manager *MemberManagerCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	return manager.provider.UpdateProfile(id, profile)
}

func (manager *MemberManagerCtx) UpdatePassword(id string, password string) error {
	return manager.provider.UpdatePassword(id, password)
}

func (manager *MemberManagerCtx) Delete(id string) error {
	return manager.provider.Delete(id)
}
