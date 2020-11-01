package session

import (
	"fmt"
	"net/http"

	"demodesk/neko/internal/utils"
)

// TODO: Refactor
func (manager *SessionManagerCtx) Authenticate(r *http.Request) (string, string, bool, error) {
	ip := r.RemoteAddr

	//if ws.conf.Proxy {
	//	ip = utils.ReadUserIP(r)
	//}

	id, err := utils.NewUID(32)
	if err != nil {
		return "", ip, false, err
	}

	passwords, ok := r.URL.Query()["password"]
	if !ok || len(passwords[0]) < 1 {
		return "", ip, false, fmt.Errorf("no password provided")
	}

	if passwords[0] == manager.config.AdminPassword {
		return id, ip, true, nil
	}

	if passwords[0] == manager.config.Password {
		return id, ip, false, nil
	}

	return "", ip, false, fmt.Errorf("invalid password: %s", passwords[0])
}
