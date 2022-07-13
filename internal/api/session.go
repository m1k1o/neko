package api

import (
	"net/http"

	"github.com/demodesk/neko/pkg/auth"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

type SessionLoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SessionDataPayload struct {
	ID      string              `json:"id"`
	Token   string              `json:"token,omitempty"`
	Profile types.MemberProfile `json:"profile"`
	State   types.SessionState  `json:"state"`
}

func (api *ApiManagerCtx) Login(w http.ResponseWriter, r *http.Request) error {
	data := &SessionLoginPayload{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	session, token, err := api.members.Login(data.Username, data.Password)
	if err != nil {
		return utils.HttpUnauthorized().WithInternalErr(err)
	}

	sessionData := SessionDataPayload{
		ID:      session.ID(),
		Profile: session.Profile(),
		State:   session.State(),
	}

	if api.sessions.CookieEnabled() {
		api.sessions.CookieSetToken(w, token)
	} else {
		sessionData.Token = token
	}

	return utils.HttpSuccess(w, sessionData)
}

func (api *ApiManagerCtx) Logout(w http.ResponseWriter, r *http.Request) error {
	session, _ := auth.GetSession(r)

	err := api.members.Logout(session.ID())
	if err != nil {
		return utils.HttpUnauthorized().WithInternalErr(err)
	}

	if api.sessions.CookieEnabled() {
		api.sessions.CookieClearToken(w, r)
	}

	return utils.HttpSuccess(w, true)
}

func (api *ApiManagerCtx) Whoami(w http.ResponseWriter, r *http.Request) error {
	session, _ := auth.GetSession(r)

	return utils.HttpSuccess(w, SessionDataPayload{
		ID:      session.ID(),
		Profile: session.Profile(),
		State:   session.State(),
	})
}
