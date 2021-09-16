package auth

import (
	"context"
	"net/http"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type key int

const keySessionCtx key = iota

func SetSession(r *http.Request, session types.Session) *http.Request {
	ctx := context.WithValue(r.Context(), keySessionCtx, session)
	return r.WithContext(ctx)
}

func GetSession(r *http.Request) types.Session {
	return r.Context().Value(keySessionCtx).(types.Session)
}

func AdminsOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)
		if !session.Profile().IsAdmin {
			utils.HttpForbidden(w).Msg("session is not admin")
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func HostsOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)
		if !session.IsHost() {
			utils.HttpForbidden(w).Msg("session is not host")
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func CanWatchOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)
		if !session.Profile().CanWatch {
			utils.HttpForbidden(w).Msg("session cannot watch")
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func CanHostOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)
		if !session.Profile().CanHost {
			utils.HttpForbidden(w).Msg("session cannot host")
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func CanAccessClipboardOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)
		if !session.Profile().CanAccessClipboard {
			utils.HttpForbidden(w).Msg("session cannot access clipboard")
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
