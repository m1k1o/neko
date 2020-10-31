package api

import (
	"net/http"

	"github.com/go-chi/chi"

	"demodesk/neko/internal/api/member"
	"demodesk/neko/internal/api/room"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/config"
	"demodesk/neko/internal/api/utils"
)

type API struct {
	sessions   types.SessionManager
	remote     types.RemoteManager
	broadcast  types.BroadcastManager
	websocket  types.WebSocketHandler
}

var AdminToken []byte
var UserToken []byte

func New(
	sessions types.SessionManager,
	remote types.RemoteManager,
	broadcast types.BroadcastManager,
	websocket types.WebSocketHandler,
	conf *config.Server,
) *API {
	AdminToken = []byte(conf.AdminToken)
	UserToken = []byte(conf.UserToken)

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

func UsersOnly(next http.Handler) http.Handler {
	return utils.AuthMiddleware(next, UserToken, AdminToken)
}

func AdminsOnly(next http.Handler) http.Handler {
	return utils.AuthMiddleware(next, AdminToken)
}
