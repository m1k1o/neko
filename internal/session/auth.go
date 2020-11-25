package session

import (
	"fmt"
	"net/http"
	"strings"

	"demodesk/neko/internal/types"
)

const (
	token_name = "password"
)

func (manager *SessionManagerCtx) Authenticate(r *http.Request) (types.Session, error) {
	token := getToken(r)
	if token == "" {
		return nil, fmt.Errorf("no password provided")
	}

	isAdmin := (token == manager.config.AdminPassword)
	isUser := (token == manager.config.Password)

	if !isAdmin && !isUser {
		return nil, fmt.Errorf("invalid password")
	}

	// TODO: Enable persistent user autentication.
	return manager.Create(types.MemberProfile{
		IsAdmin: isAdmin,
	})
}

func getToken(r *http.Request) string {
	// Get token from query
	if token := r.URL.Query().Get(token_name); token != "" {
		return token
	}

	// Get token from authorization header
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}

	// Get token from cookie
	cookie, err := r.Cookie(token_name)
	if err == nil {
		return cookie.Value
	}

	return ""
}
