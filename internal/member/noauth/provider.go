package noauth

import (
	"errors"
	"fmt"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

func New() types.MemberProvider {
	return &MemberProviderCtx{
		profile: types.MemberProfile{
			IsAdmin:               true,
			CanLogin:              true,
			CanConnect:            true,
			CanWatch:              true,
			CanHost:               true,
			CanShareMedia:         true,
			CanAccessClipboard:    true,
			SendsInactiveCursor:   true,
			CanSeeInactiveCursors: true,
		},
	}
}

type MemberProviderCtx struct {
	profile types.MemberProfile
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

	provider.profile.Name = username
	return id, provider.profile, nil
}

func (provider *MemberProviderCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	return "", errors.New("new user is created on first login in noauth mode")
}

func (provider *MemberProviderCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	return errors.New("cannot update user profile in noauth mode")
}

func (provider *MemberProviderCtx) UpdatePassword(id string, password string) error {
	return errors.New("password can only be modified in config while in noauth mode")
}

func (provider *MemberProviderCtx) Select(id string) (types.MemberProfile, error) {
	return types.MemberProfile{}, errors.New("cannot select user in noauth mode")
}

func (provider *MemberProviderCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, errors.New("cannot select users in noauth mode")
}

func (provider *MemberProviderCtx) Delete(id string) error {
	return errors.New("cannot delete user in noauth mode")
}
