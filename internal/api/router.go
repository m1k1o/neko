package api

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	
	"demodesk/neko/internal/api/member"
	"demodesk/neko/internal/api/room"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/config"
)

type API struct {
	sessions   types.SessionManager
	remote     types.RemoteManager
	broadcast  types.BroadcastManager
	websocket  types.WebSocketHandler
}

var AdminToken *jwtauth.JWTAuth
var UserToken *jwtauth.JWTAuth

func New(
	sessions types.SessionManager,
	remote types.RemoteManager,
	broadcast types.BroadcastManager,
	websocket types.WebSocketHandler,
	conf *config.Server,
) *API {
	AdminToken = jwtauth.New("HS256", []byte(conf.AdminToken), nil)
	UserToken = jwtauth.New("HS256", []byte(conf.UserToken), nil)

	return &API{
		sessions:   sessions,
		remote:     remote,
		broadcast:  broadcast,
		websocket:  websocket,
	}
}

func (a *API) Mount(r *chi.Mux) {
	memberHandler := member.New(a.sessions, a.websocket)
	r.Mount("/member", memberHandler.Router(UsersOnly, AdminsOnly))

	roomHandler := room.New(a.sessions, a.remote, a.broadcast, a.websocket)
	r.Mount("/room", roomHandler.Router(UsersOnly, AdminsOnly))
}

func UsersOnly(r chi.Router, protectedRoutes func(r chi.Router)) {
	r.Group(func(r chi.Router) {
		// Verify JWT tokens
		r.Use(jwtauth.Verifier(UserToken))
		r.Use(jwtauth.Verifier(AdminToken))

		// Handle valid / invalid tokens.
		r.Use(jwtauth.Authenticator)

		protectedRoutes(r)
	})
}

func AdminsOnly(r chi.Router, protectedRoutes func(r chi.Router)) {
	r.Group(func(r chi.Router) {
		// Verify JWT token
		r.Use(jwtauth.Verifier(AdminToken))

		// Handle valid / invalid tokens.
		r.Use(jwtauth.Authenticator)

		protectedRoutes(r)
	})
}
