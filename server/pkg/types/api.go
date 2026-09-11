package types

import "net/http"

type ApiManager interface {
	Route(r Router)
	AddRouter(path string, router func(Router))
	IsAuthenticated(r *http.Request) bool
}
