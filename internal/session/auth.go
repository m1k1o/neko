package session

import (
	"fmt"
	"net/http"

	"demodesk/neko/internal/types"
)

func (manager *SessionManagerCtx) AuthenticateRequest(r *http.Request) (types.Session, error) {
	id, secret, ok := getAuthData(r)
	if !ok {
		return nil, fmt.Errorf("no authentication provided")
	}

	return manager.Authenticate(id, secret)
}

func (manager *SessionManagerCtx) Authenticate(id string, secret string) (types.Session, error) {
	session, ok := manager.Get(id)
	if !ok {
		return nil, fmt.Errorf("member not found")
	}

	if !session.VerifySecret(secret) {
		return nil, fmt.Errorf("invalid password provided")
	}

	if !session.CanLogin() {
		return nil, fmt.Errorf("login disabled")
	}

	return session, nil
}

func getAuthData(r *http.Request) (string, string, bool) {
	// get from Cookies
	cid, err1 := r.Cookie("neko-id")
	csecret, err2 := r.Cookie("neko-secret")
	if err1 == nil && err2 == nil {
		return cid.Value, csecret.Value, true
	}

	// get from BasicAuth
	id, secret, ok := r.BasicAuth()
	if ok {
		return id, secret, true
	}

	// get from QueryParams
	id = r.URL.Query().Get("id")
	secret = r.URL.Query().Get("secret")
	if id != "" && secret != "" {
		return id, secret, true
	}

	return "", "", false
}
