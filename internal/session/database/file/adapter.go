package file

import (
    "encoding/json"
    "io/ioutil"
	"os"
	"fmt"
	"sync"

	"demodesk/neko/internal/types"
)

func New(file string) types.MembersDatabase {
	return &MembersDatabaseCtx{
		file: file,
		mu:   sync.Mutex{},
	}
}

type MembersDatabaseCtx struct {
	file string
	mu   sync.Mutex
}

func (manager *MembersDatabaseCtx) Connect() error {
	return nil
}

func (manager *MembersDatabaseCtx) Disconnect() error {
	return nil
}

func (manager *MembersDatabaseCtx) Insert(id string, profile types.MemberProfile) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	profiles, err := manager.deserialize()
	if err != nil {
		return err
	}

	_, ok := profiles[id]
	if ok {
		return fmt.Errorf("Member ID already exists.")
	}

	profiles[id] = profile

	return manager.serialize(profiles)
}

func (manager *MembersDatabaseCtx) Update(id string, profile types.MemberProfile) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	profiles, err := manager.deserialize()
	if err != nil {
		return err
	}

	_, ok := profiles[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	profiles[id] = profile

	return manager.serialize(profiles)
}

func (manager *MembersDatabaseCtx) Delete(id string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	profiles, err := manager.deserialize()
	if err != nil {
		return err
	}

	_, ok := profiles[id]
	if !ok {
		return fmt.Errorf("Member ID does not exist.")
	}

	delete(profiles, id)

	return manager.serialize(profiles)
}

func (manager *MembersDatabaseCtx) Select() (map[string]types.MemberProfile, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	profiles, err := manager.deserialize()
	return profiles, err
}

func (manager *MembersDatabaseCtx) deserialize() (map[string]types.MemberProfile, error) {
	file, err := os.OpenFile(manager.file, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

    raw, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if len(raw) == 0 {
		return map[string]types.MemberProfile{}, nil
	}

	var profiles map[string]types.MemberProfile
	if err := json.Unmarshal([]byte(raw), &profiles); err != nil {
		return nil, err
	}

	return profiles, nil
}

func (manager *MembersDatabaseCtx) serialize(data map[string]types.MemberProfile) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(manager.file, raw, os.ModePerm)
}
