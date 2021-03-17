package api

import (
	"net/http"

	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type SessionLoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SessionDataPayload struct {
	ID      string              `json:"id"`
	Profile types.MemberProfile `json:"profile"`
	State   types.SessionState  `json:"state"`
}

func (api *ApiManagerCtx) Login(w http.ResponseWriter, r *http.Request) {
	data := &SessionLoginPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	session, token, err := api.members.Login(data.Username, data.Password)
	if err != nil {
		utils.HttpUnauthorized(w, err)
		return
	}

	api.sessions.CookieSetToken(w, token)

	utils.HttpSuccess(w, SessionDataPayload{
		ID:      session.ID(),
		Profile: session.Profile(),
		State:   session.State(),
	})
}

func (api *ApiManagerCtx) Logout(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)

	err := api.members.Logout(session.ID())
	if err != nil {
		utils.HttpUnauthorized(w, err)
		return
	}

	api.sessions.CookieClearToken(w)

	utils.HttpSuccess(w, true)
}

func (api *ApiManagerCtx) Whoami(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)

	utils.HttpSuccess(w, SessionDataPayload{
		ID:      session.ID(),
		Profile: session.Profile(),
		State:   session.State(),
	})
}
