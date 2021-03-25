package session

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"demodesk/neko/internal/types"
)

func (manager *SessionManagerCtx) CookieSetToken(w http.ResponseWriter, token string) {
	sameSite := http.SameSiteDefaultMode
	if manager.config.CookieSecure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    token,
		Expires:  manager.config.CookieExpiration,
		Secure:   manager.config.CookieSecure,
		SameSite: sameSite,
		HttpOnly: true,
	})
}

func (manager *SessionManagerCtx) CookieClearToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil {
		return
	}

	cookie.Value = ""
	cookie.Expires = time.Unix(0, 0)
	http.SetCookie(w, cookie)
}

func (manager *SessionManagerCtx) Authenticate(r *http.Request) (types.Session, error) {
	token, ok := manager.getToken(r)
	if !ok {
		return nil, fmt.Errorf("no authentication provided")
	}

	session, ok := manager.GetByToken(token)
	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	if !session.Profile().CanLogin {
		return nil, fmt.Errorf("login disabled")
	}

	return session, nil
}

func (manager *SessionManagerCtx) getToken(r *http.Request) (string, bool) {
	// get from Header
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) == 2 {
		return strings.TrimSpace(splitToken[1]), true
	}

	// get from Cookie
	cookie, err := r.Cookie(manager.config.CookieName)
	if err == nil {
		return cookie.Value, true
	}

	// get from URL
	token := r.URL.Query().Get("token")
	if token != "" {
		return token, true
	}

	return "", false
}
