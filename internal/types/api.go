package types

import (
	"github.com/go-chi/chi"
)

type ApiManager interface {
	Route(r chi.Router)
	AddRouter(path string, router func(chi.Router))
}
