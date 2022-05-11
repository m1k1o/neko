package auth

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/demodesk/neko/server/pkg/types"
	"gitlab.com/demodesk/neko/server/pkg/utils"
)

type key int

const keySessionCtx key = iota

func SetSession(r *http.Request, session types.Session) context.Context {
	return context.WithValue(r.Context(), keySessionCtx, session)
}

func GetSession(r *http.Request) (types.Session, bool) {
	session, ok := r.Context().Value(keySessionCtx).(types.Session)
	return session, ok
}

func AdminsOnly(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	session, ok := GetSession(r)
	if !ok || !session.Profile().IsAdmin {
		return nil, utils.HttpForbidden("session is not admin")
	}

	return nil, nil
}

func HostsOnly(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	session, ok := GetSession(r)
	if !ok || !session.IsHost() {
		return nil, utils.HttpForbidden("session is not host")
	}

	return nil, nil
}

func CanWatchOnly(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	session, ok := GetSession(r)
	if !ok || !session.Profile().CanWatch {
		return nil, utils.HttpForbidden("session cannot watch")
	}

	return nil, nil
}

func CanHostOnly(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	session, ok := GetSession(r)
	if !ok || !session.Profile().CanHost {
		return nil, utils.HttpForbidden("session cannot host")
	}

	if session.PrivateModeEnabled() {
		return nil, utils.HttpUnprocessableEntity("private mode is enabled")
	}

	return nil, nil
}

func CanAccessClipboardOnly(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	session, ok := GetSession(r)
	if !ok || !session.Profile().CanAccessClipboard {
		return nil, utils.HttpForbidden("session cannot access clipboard")
	}

	return nil, nil
}

func PluginsGenericOnly[V comparable](key string, value V) func(w http.ResponseWriter, r *http.Request) (context.Context, error)  {
	return func(w http.ResponseWriter, r *http.Request) (context.Context, error) {
		session, ok := GetSession(r)
		if !ok {
			return nil, utils.HttpForbidden("session not found")
		}

		plugins := session.Profile().Plugins
		if val, ok := plugins[key].(V); !ok || val != value {
			return nil, utils.HttpForbidden(fmt.Sprintf("%s is set to %v; expected %v", key, value, val))
		}
	
		return nil, nil
	}
}
