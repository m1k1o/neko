package oauth

import (
	"errors"

	"github.com/m1k1o/neko/server/pkg/types"
)

// MemberProviderCtx disables password authentication. OAuth sessions are
// created by MemberManagerCtx.LoginOAuth after the OAuth callback succeeds.
type MemberProviderCtx struct{}

func New() types.MemberProvider { return &MemberProviderCtx{} }

func (provider *MemberProviderCtx) Connect() error { return nil }

func (provider *MemberProviderCtx) Disconnect() error { return nil }

func (provider *MemberProviderCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	return "", types.MemberProfile{}, types.ErrMemberInvalidPassword
}

func (provider *MemberProviderCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	return "", errors.New("OAuth members are created by OAuth login")
}

func (provider *MemberProviderCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	return errors.New("OAuth profile is managed by the identity provider")
}

func (provider *MemberProviderCtx) UpdatePassword(id string, password string) error {
	return errors.New("OAuth provider does not have passwords")
}

func (provider *MemberProviderCtx) Select(id string) (types.MemberProfile, error) {
	return types.MemberProfile{}, errors.New("OAuth members are stored in active sessions")
}

func (provider *MemberProviderCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, nil
}

func (provider *MemberProviderCtx) Delete(id string) error {
	return errors.New("OAuth members are managed by the identity provider")
}
