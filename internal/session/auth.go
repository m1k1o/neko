package session

import (
	"fmt"
	"net/http"

	"demodesk/neko/internal/types"
)

func (manager *SessionManagerCtx) Authenticate(r *http.Request) (types.Session, error) {
	id, secret, ok := r.BasicAuth()
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

	return session, nil
}
