package dummy

import (
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
	return username, types.MemberProfile{}, nil
}

func (manager *MemberManagerCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	return "", nil
}

func (manager *MemberManagerCtx) Select(id string) (types.MemberProfile, error) {
	return types.MemberProfile{}, nil
}

func (manager *MemberManagerCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, nil
}

func (manager *MemberManagerCtx) Update(id string, profile types.MemberProfile) error {
	return nil
}

func (manager *MemberManagerCtx) UpdatePassword(id string, passwrod string) error {
	return nil
}

func (manager *MemberManagerCtx) Delete(id string) error {
	return nil
}
