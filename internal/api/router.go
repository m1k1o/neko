package api

import (
	"net/http"

	"github.com/go-chi/chi"

	"demodesk/neko/internal/api/member"
	"demodesk/neko/internal/api/room"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/api/utils"
)

type ApiManagerCtx struct {
	sessions  types.SessionManager
	desktop   types.DesktopManager
	capture   types.CaptureManager
}

var AdminToken []byte
var UserToken []byte

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
	conf *config.Server,
) *ApiManagerCtx {
	AdminToken = []byte(conf.AdminToken)
	UserToken = []byte(conf.UserToken)

	return &ApiManagerCtx{
		sessions:   sessions,
		desktop:    desktop,
		capture:    capture,
	}
}

func (a *ApiManagerCtx) Mount(r *chi.Mux) {
	memberHandler := member.New(a.sessions)
	r.Mount("/member", memberHandler.Router(UsersOnly, AdminsOnly))

	roomHandler := room.New(a.sessions, a.desktop, a.capture)
	r.Mount("/room", roomHandler.Router(UsersOnly, AdminsOnly))
}

func UsersOnly(next http.Handler) http.Handler {
	return utils.AuthMiddleware(next, UserToken, AdminToken)
}

func AdminsOnly(next http.Handler) http.Handler {
	return utils.AuthMiddleware(next, AdminToken)
}
