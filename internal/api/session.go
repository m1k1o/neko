package api

import (
	"net/http"
	"os"
	"time"

	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

var CookieExpirationDate = time.Now().Add(365 * 24 * time.Hour)
var UnsecureCookies = os.Getenv("DISABLE_SECURE_COOKIES") == "true"

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

	// TODO: Proper login.
	session, token, err := api.sessions.Create(data.Username, types.MemberProfile{
		Name:               data.Username,
		IsAdmin:            true,
		CanLogin:           true,
		CanConnect:         true,
		CanWatch:           true,
		CanHost:            true,
		CanAccessClipboard: true,
	})

	if err != nil {
		utils.HttpUnauthorized(w, err)
		return
	}

	sameSite := http.SameSiteNoneMode
	if UnsecureCookies {
		sameSite = http.SameSiteDefaultMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "NEKO_SESSION",
		Value:    token,
		Expires:  CookieExpirationDate,
		Secure:   !UnsecureCookies,
		SameSite: sameSite,
		HttpOnly: true,
	})

	utils.HttpSuccess(w, SessionDataPayload{
		ID:      session.ID(),
		Profile: session.GetProfile(),
		State:   session.GetState(),
	})
}

func (api *ApiManagerCtx) Logout(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)

	// TODO: Proper logout.
	err := api.sessions.Delete(session.ID())
	if err != nil {
		utils.HttpUnauthorized(w, err)
		return
	}

	sameSite := http.SameSiteNoneMode
	if UnsecureCookies {
		sameSite = http.SameSiteDefaultMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "NEKO_SESSION",
		Value:    "",
		Expires:  time.Unix(0, 0),
		Secure:   !UnsecureCookies,
		SameSite: sameSite,
		HttpOnly: true,
	})

	utils.HttpSuccess(w, true)
}

func (api *ApiManagerCtx) Whoami(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)

	utils.HttpSuccess(w, SessionDataPayload{
		ID:      session.ID(),
		Profile: session.GetProfile(),
		State:   session.GetState(),
	})
}
