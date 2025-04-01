package file

import (
	"github.com/m1k1o/neko/server/pkg/types"
)

type MemberEntry struct {
	Password string              `json:"password"`
	Profile  types.MemberProfile `json:"profile"`
}

type Config struct {
	Path string
	Hash bool
}
