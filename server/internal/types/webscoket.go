package types

type WebScoket interface {
	Address() *string
	Send(v interface{}) error
	Destroy() error
}
