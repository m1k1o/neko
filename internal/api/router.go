package api

import (
	"net/http"

	"github.com/go-chi/chi"

	"demodesk/neko/internal/api/members"
	"demodesk/neko/internal/api/room"
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type ApiManagerCtx struct {
	sessions types.SessionManager
	desktop  types.DesktopManager
	capture  types.CaptureManager
	routers  map[string]func(chi.Router)
}

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	conf *config.Server,
) *ApiManagerCtx {

	return &ApiManagerCtx{
		sessions: sessions,
		desktop:  desktop,
		capture:  capture,
		routers:  make(map[string]func(chi.Router)),
	}
}

func (api *ApiManagerCtx) Route(r chi.Router) {
	r.Post("/login", api.Login)

	// Authenticated area
	r.Group(func(r chi.Router) {
		r.Use(api.Authenticate)

		r.Post("/logout", api.Logout)
		r.Get("/whoami", api.Whoami)

		membersHandler := members.New(api.sessions)
		r.Route("/members", membersHandler.Route)

		roomHandler := room.New(api.sessions, api.desktop, api.capture)
		r.Route("/room", roomHandler.Route)

		for path, router := range api.routers {
			r.Route(path, router)
		}
	})
}

func (api *ApiManagerCtx) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := api.sessions.Authenticate(r)
		if err != nil {
			utils.HttpUnauthorized(w, err)
		} else {
			next.ServeHTTP(w, auth.SetSession(r, session))
		}
	})
}

func (api *ApiManagerCtx) AddRouter(path string, router func(chi.Router)) {
	api.routers[path] = router
}
