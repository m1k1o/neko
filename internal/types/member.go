package types

type MemberProfile struct {
	Name               string `json:"name"`
	IsAdmin            bool   `json:"is_admin"`
	CanLogin           bool   `json:"can_login"`
	CanConnect         bool   `json:"can_connect"`
	CanWatch           bool   `json:"can_watch"`
	CanHost            bool   `json:"can_host"`
	CanAccessClipboard bool   `json:"can_access_clipboard"`
}

type MemberProvider interface {
	Connect() error
	Disconnect() error

	Authenticate(username string, password string) (string, MemberProfile, error)

	Insert(username string, password string, profile MemberProfile) (string, error)
	Select(id string) (MemberProfile, error)
	SelectAll(limit int, offset int) (map[string]MemberProfile, error)
	UpdateProfile(id string, profile MemberProfile) error
	UpdatePassword(id string, password string) error
	Delete(id string) error
}

type MemberManager interface {
	MemberProvider

	Login(username string, password string) (Session, string, error)
	Logout(id string) error
}