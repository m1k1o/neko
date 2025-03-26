package file

import (
	"encoding/json"
	"testing"

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
