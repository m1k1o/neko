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

func (manager *MemberManagerCtx) Insert(id string, profile types.MemberProfile) error {
	return nil
}

func (manager *MemberManagerCtx) Update(id string, profile types.MemberProfile) error {
	return nil
}

func (manager *MemberManagerCtx) Delete(id string) error {
	return nil
}

func (manager *MemberManagerCtx) Select() (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, nil
}
