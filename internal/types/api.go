package types

import (
	"github.com/go-chi/chi"
)

type ApiManager interface {
	Route(r chi.Router)
}
