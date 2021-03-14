package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"demodesk/neko/internal/types"
)

func New(config Config) types.MemberManager {
	return &MemberManagerCtx{
		config: config,
		mu:     sync.Mutex{},
	}
}

type MemberManagerCtx struct {
	config Config
	mu     sync.Mutex
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

	entry, err := manager.getEntry(id)
	if err != nil {
		return "", types.MemberProfile{}, err
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

	entries, err := manager.deserialize()
	if err != nil {
		return "", err
	}

	_, ok := entries[id]
	if ok {
		return "", fmt.Errorf("Member ID already exists.")
	}

	entries[id] = MemberEntry{
		// TODO: Use hash function.
		Password: password,
		Profile: profile,
	}

	return id, manager.serialize(entries)
}

func (manager *MemberManagerCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entries, err := manager.deserialize()
	if err != nil {
		return err
	}

	entry, ok := entries[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	entry.Profile = profile
	entries[id] = entry

	return manager.serialize(entries)
}

func (manager *MemberManagerCtx) UpdatePassword(id string, password string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entries, err := manager.deserialize()
	if err != nil {
		return err
	}

	entry, ok := entries[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	// TODO: Use hash function.
	entry.Password = password
	entries[id] = entry

	return manager.serialize(entries)
}

func (manager *MemberManagerCtx) Select(id string) (types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, err := manager.getEntry(id)
	if err != nil {
		return types.MemberProfile{}, err
	}

	return entry.Profile, nil
}

func (manager *MemberManagerCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	profiles := map[string]types.MemberProfile{}

	entries, err := manager.deserialize()
	if err != nil {
		return profiles, err
	}

	i := 0
	for id, entry := range entries {
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

	entries, err := manager.deserialize()
	if err != nil {
		return err
	}

	_, ok := entries[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	delete(entries, id)

	return manager.serialize(entries)
}

func (manager *MemberManagerCtx) deserialize() (map[string]MemberEntry, error) {
	file, err := os.OpenFile(manager.config.File, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	raw, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if len(raw) == 0 {
		return map[string]MemberEntry{}, nil
	}

	var entries map[string]MemberEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func (manager *MemberManagerCtx) getEntry(id string) (MemberEntry, error) {
	entries, err := manager.deserialize()
	if err != nil {
		return MemberEntry{}, err
	}

	entry, ok := entries[id]
	if !ok {
		return MemberEntry{}, fmt.Errorf("Member ID does not exist.")
	}

	return entry, nil
}

func (manager *MemberManagerCtx) serialize(data map[string]MemberEntry) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(manager.config.File, raw, os.ModePerm)
}
