package types

type ApiManager interface {
	Route(r Router)
	AddRouter(path string, router func(Router))
}
