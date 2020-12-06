package session

import (
	"fmt"
	"net/http"

	"demodesk/neko/internal/types"
)

func (manager *SessionManagerCtx) Authenticate(r *http.Request) (types.Session, error) {
	id, secret, ok := getAuthData(r)
	if !ok {
		return nil, fmt.Errorf("no authentication provided")
	}

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
	id, secret, ok := r.BasicAuth()
	if ok {
		return id, secret, true
	}

	id = r.URL.Query().Get("id")
	secret = r.URL.Query().Get("secret")

	if id != "" && secret != "" {
		return id, secret, true
	}

	return "", "", false
}