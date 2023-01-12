package multiuser

import (
	"errors"
	"fmt"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

func New(config Config) types.MemberProvider {
	return &MemberProviderCtx{
		config: config,
	}
}

type MemberProviderCtx struct {
	config Config
}

func (provider *MemberProviderCtx) Connect() error {
	return nil
}

func (provider *MemberProviderCtx) Disconnect() error {
	return nil
}

func (provider *MemberProviderCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	// generate random token
	token, err := utils.NewUID(5)
	if err != nil {
		return "", types.MemberProfile{}, err
	}

	// id is username with token
	id := fmt.Sprintf("%s-%s", username, token)

	// if logged in as administrator
	if provider.config.AdminPassword == password {
		return id, types.MemberProfile{
			Name:                  username,
			IsAdmin:               true,
			CanLogin:              true,
			CanConnect:            true,
			CanWatch:              true,
			CanHost:               true,
			CanShareMedia:         true,
			CanAccessClipboard:    true,
			SendsInactiveCursor:   true,
			CanSeeInactiveCursors: true,
		}, nil
	}

	// if logged in as user
	if provider.config.UserPassword == password {
		return id, types.MemberProfile{
			Name:                  username,
			IsAdmin:               false,
			CanLogin:              true,
			CanConnect:            true,
			CanWatch:              true,
			CanHost:               true,
			CanShareMedia:         true,
			CanAccessClipboard:    true,
			SendsInactiveCursor:   true,
			CanSeeInactiveCursors: false,
		}, nil
	}

	return "", types.MemberProfile{}, types.ErrMemberInvalidPassword
}

func (provider *MemberProviderCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	return "", errors.New("new user is created on first login in multiuser mode")
}

func (provider *MemberProviderCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	return nil
}

func (provider *MemberProviderCtx) UpdatePassword(id string, password string) error {
	return errors.New("password can only be modified in config while in multiuser mode")
}

func (provider *MemberProviderCtx) Select(id string) (types.MemberProfile, error) {
	return types.MemberProfile{}, errors.New("cannot select user in multiuser mode")
}

func (provider *MemberProviderCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, nil
}

func (provider *MemberProviderCtx) Delete(id string) error {
	return errors.New("cannot delete user in multiuser mode")
}
