package session

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

func (manager *SessionManagerCtx) CookieSetToken(w http.ResponseWriter, token string) {
	sameSite := http.SameSiteDefaultMode
	if manager.config.Cookie.Secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     manager.config.Cookie.Name,
		Value:    token,
		Expires:  time.Now().Add(manager.config.Cookie.Expiration),
		Secure:   manager.config.Cookie.Secure,
		SameSite: sameSite,
		HttpOnly: manager.config.Cookie.HTTPOnly,
		Domain:   manager.config.Cookie.Domain,
		Path:     manager.config.Cookie.Path,
	})
}

func (manager *SessionManagerCtx) CookieClearToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.config.Cookie.Name)
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
		return nil, errors.New("no authentication provided")
	}

	session, ok := manager.GetByToken(token)
	if !ok {
		return nil, types.ErrSessionNotFound
	}

	if !session.Profile().CanLogin {
		return nil, types.ErrSessionLoginDisabled
	}

	return session, nil
}

func (manager *SessionManagerCtx) getToken(r *http.Request) (string, bool) {
	if manager.CookieEnabled() {
		// get from Cookie
		cookie, err := r.Cookie(manager.config.Cookie.Name)
		if err == nil {
			return cookie.Value, true
		}
	}

	// get from Header
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) == 2 {
		return strings.TrimSpace(splitToken[1]), true
	}

	// get from URL
	token := r.URL.Query().Get("token")
	if token != "" {
		return token, true
	}

	return "", false
}
