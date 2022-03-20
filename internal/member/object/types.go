package object

import (
	"gitlab.com/demodesk/neko/server/pkg/types"
)

type MemberEntry struct {
	Password string              `json:"password"`
	Profile  types.MemberProfile `json:"profile"`
}

type Config struct {
	AdminPassword string
	UserPassword  string
}
