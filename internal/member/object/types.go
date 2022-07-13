package object

import (
	"github.com/demodesk/neko/pkg/types"
)

type MemberEntry struct {
	Password string              `json:"password"`
	Profile  types.MemberProfile `json:"profile"`
}

type Config struct {
	AdminPassword string
	UserPassword  string
}
