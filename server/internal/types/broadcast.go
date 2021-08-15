package types

type BroadcastManager interface {
	Start() error
	Stop()
	IsActive() bool
	Create(url string) error
	Destroy()
	GetUrl() string
}
