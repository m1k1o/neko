package object

import (
	"github.com/demodesk/neko/pkg/types"
)

func New(config Config) types.MemberProvider {
	return &MemberProviderCtx{
		config:  config,
		entries: make(map[string]*MemberEntry),
	}
}

type MemberProviderCtx struct {
	config  Config
	entries map[string]*MemberEntry
}

func (provider *MemberProviderCtx) Connect() error {
	var err error

	if provider.config.AdminPassword != "" {
		// create default admin account at startup
		_, err = provider.Insert("admin", provider.config.AdminPassword, types.MemberProfile{
			Name:                  "Administrator",
			IsAdmin:               true,
			CanLogin:              true,
			CanConnect:            true,
			CanWatch:              true,
			CanHost:               true,
			CanShareMedia:         true,
			CanAccessClipboard:    true,
			SendsInactiveCursor:   true,
			CanSeeInactiveCursors: true,
		})
	}

	if provider.config.UserPassword != "" {
		// create default user account at startup
		_, err = provider.Insert("user", provider.config.UserPassword, types.MemberProfile{
			Name:                  "User",
			IsAdmin:               false,
			CanLogin:              true,
			CanConnect:            true,
			CanWatch:              true,
			CanHost:               true,
			CanShareMedia:         true,
			CanAccessClipboard:    true,
			SendsInactiveCursor:   true,
			CanSeeInactiveCursors: false,
		})
	}

	return err
}

func (provider *MemberProviderCtx) Disconnect() error {
	return nil
}

func (provider *MemberProviderCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	// id will be also username
	id := username

	entry, ok := provider.entries[id]
	if !ok {
		return "", types.MemberProfile{}, types.ErrMemberDoesNotExist
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

	_, ok := provider.entries[id]
	if ok {
		return "", types.ErrMemberAlreadyExists
	}

	provider.entries[id] = &MemberEntry{
		// TODO: Use hash function.
		Password: password,
		Profile:  profile,
	}

	return id, nil
}

func (provider *MemberProviderCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	entry, ok := provider.entries[id]
	if !ok {
		return types.ErrMemberDoesNotExist
	}

	entry.Profile = profile

	return nil
}

func (provider *MemberProviderCtx) UpdatePassword(id string, password string) error {
	entry, ok := provider.entries[id]
	if !ok {
		return types.ErrMemberDoesNotExist
	}

	// TODO: Use hash function.
	entry.Password = password

	return nil
}

func (provider *MemberProviderCtx) Select(id string) (types.MemberProfile, error) {
	entry, ok := provider.entries[id]
	if !ok {
		return types.MemberProfile{}, types.ErrMemberDoesNotExist
	}

	return entry.Profile, nil
}

func (provider *MemberProviderCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	profiles := make(map[string]types.MemberProfile)

	i := 0
	for id, entry := range provider.entries {
		if i >= offset && (limit == 0 || i < offset+limit) {
			profiles[id] = entry.Profile
		}

		i = i + 1
	}

	return profiles, nil
}

func (provider *MemberProviderCtx) Delete(id string) error {
	_, ok := provider.entries[id]
	if !ok {
		return types.ErrMemberDoesNotExist
	}

	delete(provider.entries, id)

	return nil
}
