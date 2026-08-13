package oauth

import (
	"github.com/m1k1o/neko/server/pkg/types"
)

type Config struct {
	Enabled          bool
	AutoRedirect     bool
	Name             string
	AdminEmails      []string
	ClientID         string
	ClientSecret     string
	IssuerURL        string
	AuthorizationURL string
	TokenURL         string
	UserInfoURL      string
	RedirectURL      string
	Scopes           []string
	SubjectField     string
	UsernameField    string
	AvatarField      string
	SuccessRedirect  string
	AdminProfile     types.MemberProfile
	UserProfile      types.MemberProfile
}
