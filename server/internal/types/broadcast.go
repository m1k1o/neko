package types

type BroadcastManager interface {
	Shutdown() error
	Start() error
	Stop()
	IsActive() bool
	Create(url string) error
	Destroy()
	GetUrl() string
}
