package object

import (
	"fmt"
	"sync"

	"demodesk/neko/internal/types"
)

func New() types.MemberManager {
	return &MemberManagerCtx{
		entries: make(map[string]MemberEntry),
		mu:      sync.Mutex{},
	}
}

type MemberManagerCtx struct {
	entries map[string]MemberEntry
	mu      sync.Mutex
}

type MemberEntry struct {
	Password string              `json:"password"`
	Profile  types.MemberProfile `json:"profile"`
}

func (manager *MemberManagerCtx) Connect() error {
	return nil
}

func (manager *MemberManagerCtx) Disconnect() error {
	return nil
}

func (manager *MemberManagerCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// id will be also username
	id := username

	entry, ok := manager.entries[id]
	if !ok {
		return "", types.MemberProfile{}, fmt.Errorf("Member ID does not exist.")
	}

	// TODO: Use hash function.
	if entry.Password != password {
		return "", types.MemberProfile{}, fmt.Errorf("Invalid password.")
	}

	return id, entry.Profile, nil
}

func (manager *MemberManagerCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// id will be also username
	id := username

	entry, ok := manager.entries[id]
	if ok {
		return "", fmt.Errorf("Member ID already exists.")
	}

	// TODO: Use hash function.
	entry.Password = password
	entry.Profile = profile
	manager.entries[id] = entry

	return id, nil
}

func (manager *MemberManagerCtx) Update(id string, profile types.MemberProfile) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, ok := manager.entries[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	entry.Profile = profile
	manager.entries[id] = entry

	return nil
}

func (manager *MemberManagerCtx) UpdatePassword(id string, password string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, ok := manager.entries[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	// TODO: Use hash function.
	entry.Password = password
	manager.entries[id] = entry

	return nil
}

func (manager *MemberManagerCtx) Select(id string) (types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, ok := manager.entries[id]
	if ok {
		return types.MemberProfile{}, fmt.Errorf("Member ID already exists.")
	}

	return entry.Profile, nil
}

func (manager *MemberManagerCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	profiles := map[string]types.MemberProfile{}

	i := 0
	for id, entry := range manager.entries {
		if i < offset || i > offset + limit {
			continue
		}

		profiles[id] = entry.Profile
		i = i + 1
	}

	return profiles, nil
}

func (manager *MemberManagerCtx) Delete(id string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	_, ok := manager.entries[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	delete(manager.entries, id)

	return nil
}
