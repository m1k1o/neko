package multiuser

import "m1k1o/neko/pkg/types"

type Config struct {
	AdminPassword string
	UserPassword  string
	AdminProfile  types.MemberProfile
	UserProfile   types.MemberProfile
}
