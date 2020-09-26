package types

type BroadcastManager interface {
	Start()
	Stop()
	IsActive() bool
	Create(url string)
	Destroy()
	GetUrl() string
}
