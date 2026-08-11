package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/m1k1o/neko/server/internal/api/members"
	"github.com/m1k1o/neko/server/internal/api/room"
	"github.com/m1k1o/neko/server/internal/api/sessions"
	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/auth"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type ApiManagerCtx struct {
	sessions types.SessionManager
	members  types.MemberManager
	desktop  types.DesktopManager
	capture  types.CaptureManager
	routers  map[string]func(types.Router)
	oauth    *oauthHandler
}

func New(
	sessions types.SessionManager,
	members types.MemberManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	memberConfig *config.Member,
	serverConfig *config.Server,
) *ApiManagerCtx {
	pathPrefix := "/"
	if serverConfig != nil {
		pathPrefix = serverConfig.PathPrefix
	}

	return &ApiManagerCtx{
		sessions: sessions,
		members:  members,
		desktop:  desktop,
		capture:  capture,
		routers:  make(map[string]func(types.Router)),
		oauth:    newOAuthHandler(memberConfig.OAuth, pathPrefix, serverConfig != nil && serverConfig.Proxy, memberConfig.Provider == "oauth"),
	}
}

func (api *ApiManagerCtx) Route(r types.Router) {
	r.Post("/login", api.Login)
	r.Get("/oauth/config", api.OAuthConfig)
	r.Get("/oauth/login", api.OAuthLogin)
	r.Get("/oauth/callback", api.OAuthCallback)

	// Authenticated area
	r.Group(func(r types.Router) {
		r.Use(api.Authenticate)

		r.Post("/logout", api.Logout)
		r.Get("/whoami", api.Whoami)
		r.Post("/profile", api.UpdateProfile)
		r.Get("/stats", api.Stats)

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

func (api *ApiManagerCtx) IsAuthenticated(r *http.Request) bool {
	_, err := api.sessions.Authenticate(r)
	return err == nil
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
