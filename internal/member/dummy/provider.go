package dummy

import (
	"errors"

	"github.com/demodesk/neko/pkg/types"
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
	provider.profile.Name = username
	return username, provider.profile, nil
}

func (provider *MemberProviderCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	return "", errors.New("not implemented")
}

func (provider *MemberProviderCtx) Select(id string) (types.MemberProfile, error) {
	provider.profile.Name = id
	return provider.profile, nil
}

func (provider *MemberProviderCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, nil
}

func (provider *MemberProviderCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	return nil
}

func (provider *MemberProviderCtx) UpdatePassword(id string, password string) error {
	return errors.New("not implemented")
}

func (provider *MemberProviderCtx) Delete(id string) error {
	return nil
}
