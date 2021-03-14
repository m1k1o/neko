package dummy

import (
	"fmt"

	"demodesk/neko/internal/types"
)

func New() types.MemberProvider {
	return &MemberProviderCtx{}
}

type MemberProviderCtx struct{}

func (provider *MemberProviderCtx) Connect() error {
	return nil
}

func (provider *MemberProviderCtx) Disconnect() error {
	return nil
}

func (provider *MemberProviderCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	return username, types.MemberProfile{
		Name:               username,
		IsAdmin:            true,
		CanLogin:           true,
		CanConnect:         true,
		CanWatch:           true,
		CanHost:            true,
		CanAccessClipboard: true,
	}, nil
}

func (provider *MemberProviderCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	return "", fmt.Errorf("Not implemented.")
}

func (provider *MemberProviderCtx) Select(id string) (types.MemberProfile, error) {
	return types.MemberProfile{}, fmt.Errorf("Not implemented.")
}

func (provider *MemberProviderCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, fmt.Errorf("Not implemented.")
}

func (provider *MemberProviderCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	return fmt.Errorf("Not implemented.")
}

func (provider *MemberProviderCtx) UpdatePassword(id string, password string) error {
	return fmt.Errorf("Not implemented.")
}

func (provider *MemberProviderCtx) Delete(id string) error {
	return fmt.Errorf("Not implemented.")
}
