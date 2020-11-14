package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"

	"demodesk/neko/internal/api/member"
	"demodesk/neko/internal/api/room"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/config"
)

type ApiManagerCtx struct {
	sessions  types.SessionManager
	desktop   types.DesktopManager
	capture   types.CaptureManager
}

const (
    keySessionCtx int = iota
)

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	conf *config.Server,
) *ApiManagerCtx {

	return &ApiManagerCtx{
		sessions:   sessions,
		desktop:    desktop,
		capture:    capture,
	}
}

func (api *ApiManagerCtx) Route(r chi.Router) {
	r.Use(api.Authenticate)

	memberHandler := member.New(api.sessions)
	r.Route("/member", memberHandler.Route)

	roomHandler := room.New(api.sessions, api.desktop, api.capture)
	r.Route("/room", roomHandler.Route)

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)
		utils.HttpBadRequest(w, "Hi `" + session.ID() + "`, you are authenticated.")
	})
}

func (api *ApiManagerCtx) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := api.sessions.Authenticate(r)
		if err != nil {
			utils.HttpNotAuthenticated(w, err)
		} else {
			ctx := context.WithValue(r.Context(), keySessionCtx, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func GetSession(r *http.Request) types.Session {
	return r.Context().Value(keySessionCtx).(types.Session)
}
