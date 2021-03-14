package dummy

import (
	"fmt"

	"demodesk/neko/internal/types"
)

func New() types.MemberManager {
	return &MemberManagerCtx{}
}

type MemberManagerCtx struct{}

func (manager *MemberManagerCtx) Connect() error {
	return nil
}

func (manager *MemberManagerCtx) Disconnect() error {
	return nil
}

func (manager *MemberManagerCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
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

func (manager *MemberManagerCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	return "", fmt.Errorf("Not implemented.")
}

func (manager *MemberManagerCtx) Select(id string) (types.MemberProfile, error) {
	return types.MemberProfile{}, fmt.Errorf("Not implemented.")
}

func (manager *MemberManagerCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, fmt.Errorf("Not implemented.")
}

func (manager *MemberManagerCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	return fmt.Errorf("Not implemented.")
}

func (manager *MemberManagerCtx) UpdatePassword(id string, passwrod string) error {
	return fmt.Errorf("Not implemented.")
}

func (manager *MemberManagerCtx) Delete(id string) error {
	return fmt.Errorf("Not implemented.")
}
