package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/demodesk/neko/internal/api/members"
	"github.com/demodesk/neko/internal/api/room"
	"github.com/demodesk/neko/internal/api/sessions"
	"github.com/demodesk/neko/pkg/auth"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

type ApiManagerCtx struct {
	sessions types.SessionManager
	members  types.MemberManager
	desktop  types.DesktopManager
	capture  types.CaptureManager
	routers  map[string]func(types.Router)
}

func New(
	sessions types.SessionManager,
	members types.MemberManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
) *ApiManagerCtx {

	return &ApiManagerCtx{
		sessions: sessions,
		members:  members,
		desktop:  desktop,
		capture:  capture,
		routers:  make(map[string]func(types.Router)),
	}
}

func (api *ApiManagerCtx) Route(r types.Router) {
	r.Post("/login", api.Login)

	// Authenticated area
	r.Group(func(r types.Router) {
		r.Use(api.Authenticate)

		r.Post("/logout", api.Logout)
		r.Get("/whoami", api.Whoami)

		sessionsHandler := sessions.New(api.sessions)
		r.Route("/sessions", sessionsHandler.Route)

		membersHandler := members.New(api.members)
		r.Route("/members", membersHandler.Route)
		r.Route("/members_bulk", membersHandler.RouteBulk)

		roomHandler := room.New(api.sessions, api.desktop, api.capture)
		r.Route("/room", roomHandler.Route)

		for path, router := range api.routers {
			r.Route(path, router)
		}
	})
}

func (api *ApiManagerCtx) Authenticate(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	session, err := api.sessions.Authenticate(r)
	if err != nil {
		if api.sessions.CookieEnabled() {
			api.sessions.CookieClearToken(w, r)
		}

		if errors.Is(err, types.ErrSessionLoginDisabled) {
			return nil, utils.HttpForbidden("login is disabled for this session")
		}

		return nil, utils.HttpUnauthorized().WithInternalErr(err)
	}

	return auth.SetSession(r, session), nil
}

func (api *ApiManagerCtx) AddRouter(path string, router func(types.Router)) {
	api.routers[path] = router
}
