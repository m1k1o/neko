package file

import (
	"encoding/json"
	"maps"
	"path/filepath"
	"slices"
	"testing"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

// Ensure that hashes are the same after encoding and decoding using json
func TestMemberProviderCtx_hash(t *testing.T) {
	provider := &MemberProviderCtx{
		config: Config{
			Hash: true,
		},
	}

	// generate random strings
	passwords := []string{}
	for i := 0; i < 10; i++ {
		password, err := utils.NewUID(32)
		if err != nil {
			t.Errorf("utils.NewUID() returned error: %s", err)
		}
		passwords = append(passwords, password)
	}

	for _, password := range passwords {
		hashedPassword := provider.hash(password)

		// json encode password hash
		hashedPasswordJSON, err := json.Marshal(hashedPassword)
		if err != nil {
			t.Errorf("json.Marshal() returned error: %s", err)
		}

		// json decode password hash json
		var hashedPasswordStr string
		err = json.Unmarshal(hashedPasswordJSON, &hashedPasswordStr)
		if err != nil {
			t.Errorf("json.Unmarshal() returned error: %s", err)
		}

		if hashedPasswordStr != hashedPassword {
			t.Errorf("hashedPasswordStr: %s != hashedPassword: %s", hashedPasswordStr, hashedPassword)
		}
	}
}

func TestMemberProviderCtx_SelectAll(t *testing.T) {
	provider := &MemberProviderCtx{
		config: Config{
			Path: filepath.Join(t.TempDir(), "members.json"),
		},
	}

	for _, id := range []string{"dave", "alice", "erin", "bob", "carol"} {
		if _, err := provider.Insert(id, "password", types.MemberProfile{Name: id}); err != nil {
			t.Fatalf("provider.Insert() returned error: %s", err)
		}
	}

	pages := [][]string{{"alice", "bob"}, {"carol", "dave"}, {"erin"}}
	for i, want := range pages {
		profiles, err := provider.SelectAll(2, i*2)
		if err != nil {
			t.Fatalf("provider.SelectAll() returned error: %s", err)
		}

		got := slices.Sorted(maps.Keys(profiles))
		if !slices.Equal(got, want) {
			t.Errorf("page %d: got %v, want %v", i, got, want)
		}
	}
}
