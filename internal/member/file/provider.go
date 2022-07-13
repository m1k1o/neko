package file

import (
	"encoding/json"
	"io"
	"os"

	"github.com/demodesk/neko/pkg/types"
)

func New(config Config) types.MemberProvider {
	return &MemberProviderCtx{
		config: config,
	}
}

type MemberProviderCtx struct {
	config Config
}

func (provider *MemberProviderCtx) Connect() error {
	return nil
}

func (provider *MemberProviderCtx) Disconnect() error {
	return nil
}

func (provider *MemberProviderCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	// id will be also username
	id := username

	entry, err := provider.getEntry(id)
	if err != nil {
		return "", types.MemberProfile{}, err
	}

	// TODO: Use hash function.
	if entry.Password != password {
		return "", types.MemberProfile{}, types.ErrMemberInvalidPassword
	}

	return id, entry.Profile, nil
}

func (provider *MemberProviderCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	// id will be also username
	id := username

	entries, err := provider.deserialize()
	if err != nil {
		return "", err
	}

	_, ok := entries[id]
	if ok {
		return "", types.ErrMemberAlreadyExists
	}

	entries[id] = MemberEntry{
		// TODO: Use hash function.
		Password: password,
		Profile:  profile,
	}

	return id, provider.serialize(entries)
}

func (provider *MemberProviderCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	entries, err := provider.deserialize()
	if err != nil {
		return err
	}

	entry, ok := entries[id]
	if !ok {
		return types.ErrMemberDoesNotExist
	}

	entry.Profile = profile
	entries[id] = entry

	return provider.serialize(entries)
}

func (provider *MemberProviderCtx) UpdatePassword(id string, password string) error {
	entries, err := provider.deserialize()
	if err != nil {
		return err
	}

	entry, ok := entries[id]
	if !ok {
		return types.ErrMemberDoesNotExist
	}

	// TODO: Use hash function.
	entry.Password = password
	entries[id] = entry

	return provider.serialize(entries)
}

func (provider *MemberProviderCtx) Select(id string) (types.MemberProfile, error) {
	entry, err := provider.getEntry(id)
	if err != nil {
		return types.MemberProfile{}, err
	}

	return entry.Profile, nil
}

func (provider *MemberProviderCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	profiles := map[string]types.MemberProfile{}

	entries, err := provider.deserialize()
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

func (provider *MemberProviderCtx) Delete(id string) error {
	entries, err := provider.deserialize()
	if err != nil {
		return err
	}

	_, ok := entries[id]
	if !ok {
		return types.ErrMemberDoesNotExist
	}

	delete(entries, id)

	return provider.serialize(entries)
}

func (provider *MemberProviderCtx) deserialize() (map[string]MemberEntry, error) {
	file, err := os.OpenFile(provider.config.Path, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	raw, err := io.ReadAll(file)
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

func (provider *MemberProviderCtx) getEntry(id string) (MemberEntry, error) {
	entries, err := provider.deserialize()
	if err != nil {
		return MemberEntry{}, err
	}

	entry, ok := entries[id]
	if !ok {
		return MemberEntry{}, types.ErrMemberDoesNotExist
	}

	return entry, nil
}

func (provider *MemberProviderCtx) serialize(data map[string]MemberEntry) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(provider.config.Path, raw, os.ModePerm)
}
