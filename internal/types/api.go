package types

import (
	"github.com/go-chi/chi"
)

type ApiManager interface {
	Mount(r *chi.Mux)
}
