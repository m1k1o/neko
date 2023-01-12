package types

import "errors"

var (
	ErrMemberAlreadyExists   = errors.New("member already exists")
	ErrMemberDoesNotExist    = errors.New("member does not exist")
	ErrMemberInvalidPassword = errors.New("invalid password")
)

type MemberProfile struct {
	Name string `json:"name"`

	// permissions
	IsAdmin               bool `json:"is_admin"                 mapstructure:"is_admin"`
	CanLogin              bool `json:"can_login"                mapstructure:"can_login"`
	CanConnect            bool `json:"can_connect"              mapstructure:"can_connect"`
	CanWatch              bool `json:"can_watch"                mapstructure:"can_watch"`
	CanHost               bool `json:"can_host"                 mapstructure:"can_host"`
	CanShareMedia         bool `json:"can_share_media"          mapstructure:"can_share_media"`
	CanAccessClipboard    bool `json:"can_access_clipboard"     mapstructure:"can_access_clipboard"`
	SendsInactiveCursor   bool `json:"sends_inactive_cursor"    mapstructure:"sends_inactive_cursor"`
	CanSeeInactiveCursors bool `json:"can_see_inactive_cursors" mapstructure:"can_see_inactive_cursors"`

	// plugin scope
	Plugins map[string]any `json:"plugins"`
}

type MemberProvider interface {
	Connect() error
	Disconnect() error

	Authenticate(username string, password string) (id string, profile MemberProfile, err error)

	Insert(username string, password string, profile MemberProfile) (id string, err error)
	Select(id string) (profile MemberProfile, err error)
	SelectAll(limit int, offset int) (profiles map[string]MemberProfile, err error)
	UpdateProfile(id string, profile MemberProfile) error
	UpdatePassword(id string, password string) error
	Delete(id string) error
}

type MemberManager interface {
	MemberProvider

	Login(username string, password string) (Session, string, error)
	Logout(id string) error
}
