package object

import (
	"fmt"
	"sync"

	"demodesk/neko/internal/types"
)

func New() types.MemberManager {
	return &MemberManagerCtx{
		profiles: make(map[string]types.MemberProfile),
		mu:       sync.Mutex{},
	}
}

type MemberManagerCtx struct {
	profiles map[string]types.MemberProfile
	mu       sync.Mutex
}

func (manager *MemberManagerCtx) Connect() error {
	return nil
}

func (manager *MemberManagerCtx) Disconnect() error {
	return nil
}

func (manager *MemberManagerCtx) Insert(id string, profile types.MemberProfile) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	_, ok := manager.profiles[id]
	if ok {
		return fmt.Errorf("Member ID already exists.")
	}

	manager.profiles[id] = profile
	return nil
}

func (manager *MemberManagerCtx) Update(id string, profile types.MemberProfile) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	_, ok := manager.profiles[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	manager.profiles[id] = profile
	return nil
}

func (manager *MemberManagerCtx) Delete(id string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	_, ok := manager.profiles[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	delete(manager.profiles, id)
	return nil
}

func (manager *MemberManagerCtx) Select() (map[string]types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.profiles, nil
}
