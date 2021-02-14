package api

import (
	"net/http"
	"time"

	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type SessionLoginPayload struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

type SessionWhoamiPayload struct {
	ID      string              `json:"id"`
	Profile types.MemberProfile `json:"profile"`
	State   types.MemberState   `json:"state"`
}

func (api *ApiManagerCtx) Login(w http.ResponseWriter, r *http.Request) {
	data := &SessionLoginPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	session, err := api.sessions.Authenticate(data.ID, data.Secret)
	if err != nil {
		utils.HttpUnauthorized(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "neko-id",
		Value:    session.ID(),
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: false,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "neko-secret",
		Value:    data.Secret,
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
	})

	utils.HttpSuccess(w, SessionWhoamiPayload{
		ID:      session.ID(),
		Profile: session.GetProfile(),
		State:   session.GetState(),
	})
}

func (api *ApiManagerCtx) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "neko-id",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: false,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "neko-secret",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	utils.HttpSuccess(w, true)
}

func (api *ApiManagerCtx) Whoami(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)

	utils.HttpSuccess(w, SessionWhoamiPayload{
		ID:      session.ID(),
		Profile: session.GetProfile(),
		State:   session.GetState(),
	})
}
