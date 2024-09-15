package file

import (
	"m1k1o/neko/pkg/types"
)

type MemberEntry struct {
	Password string              `json:"password"`
	Profile  types.MemberProfile `json:"profile"`
}

type Config struct {
	Path string
	Hash bool
}
