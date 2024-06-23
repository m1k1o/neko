package multiuser

import "github.com/demodesk/neko/pkg/types"

type Config struct {
	AdminPassword string
	UserPassword  string
	AdminProfile  types.MemberProfile
	UserProfile   types.MemberProfile
}
