package object

import (
	"fmt"
	"sync"

	"demodesk/neko/internal/types"
)

func New(config Config) types.MemberManager {
	return &MemberManagerCtx{
		config:  config,
		entries: make(map[string]*MemberEntry),
		mu:      sync.Mutex{},
	}
}

type MemberManagerCtx struct {
	config  Config
	entries map[string]*MemberEntry
	mu      sync.Mutex
}

func (manager *MemberManagerCtx) Connect() error {
	var err error

	if manager.config.AdminPassword != "" {
		// create default admin account at startup
		_, err = manager.Insert("admin", manager.config.AdminPassword, types.MemberProfile{
			Name:               "Administrator",
			IsAdmin:            true,
			CanLogin:           true,
			CanConnect:         true,
			CanWatch:           true,
			CanHost:            true,
			CanAccessClipboard: true,
		})
	}

	if manager.config.UserPassword != "" {
		// create default user account at startup
		_, err = manager.Insert("user", manager.config.UserPassword, types.MemberProfile{
			Name:               "User",
			IsAdmin:            false,
			CanLogin:           true,
			CanConnect:         true,
			CanWatch:           true,
			CanHost:            true,
			CanAccessClipboard: true,
		})
	}

	return err
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

	_, ok := manager.entries[id]
	if ok {
		return "", fmt.Errorf("Member ID already exists.")
	}

	manager.entries[id] = &MemberEntry{
		// TODO: Use hash function.
		Password: password,
		Profile: profile,
	}

	return id, nil
}

func (manager *MemberManagerCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, ok := manager.entries[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	entry.Profile = profile

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

	return nil
}

func (manager *MemberManagerCtx) Select(id string) (types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, ok := manager.entries[id]
	if !ok {
		return types.MemberProfile{}, fmt.Errorf("Member ID does not exist.")
	}

	return entry.Profile, nil
}

func (manager *MemberManagerCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	profiles := make(map[string]types.MemberProfile)

	i := 0
	for id, entry := range manager.entries {
		if i >= offset && (limit == 0 || i < offset+limit) {
			profiles[id] = entry.Profile
		}

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
