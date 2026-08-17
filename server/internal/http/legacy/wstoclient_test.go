package legacy

import (
	"testing"

	"github.com/m1k1o/neko/server/pkg/types"
)

func TestProfileToMemberPreservesAvatar(t *testing.T) {
	member, err := profileToMember("oauth:user-123", types.MemberProfile{
		Name:   "Ada Lovelace",
		Avatar: "https://example.test/ada.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if member.Name != "Ada Lovelace" || member.Avatar != "https://example.test/ada.png" {
		t.Fatalf("member = %#v", member)
	}
}
