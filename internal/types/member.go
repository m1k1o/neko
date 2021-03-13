package types

type MembersDatabase interface {
	Connect() error
	Disconnect() error

	Insert(id string, profile MemberProfile) error
	Update(id string, profile MemberProfile) error
	Delete(id string) error
	Select() (map[string]MemberProfile, error)
}
