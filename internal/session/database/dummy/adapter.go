package dummy

import (
	"demodesk/neko/internal/types"
)

func New() types.MembersDatabase {
	return &MembersDatabaseCtx{}
}

type MembersDatabaseCtx struct {}

func (manager *MembersDatabaseCtx) Connect() error {
	return nil
}

func (manager *MembersDatabaseCtx) Disconnect() error {
	return nil
}

func (manager *MembersDatabaseCtx) Insert(id string, profile types.MemberProfile) error {
	return nil
}

func (manager *MembersDatabaseCtx) Update(id string, profile types.MemberProfile) error {
	return nil
}

func (manager *MembersDatabaseCtx) Delete(id string) error {
	return nil
}

func (manager *MembersDatabaseCtx) Select() (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, nil
}
