package member

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/http/auth"
)

type key int

const (
    keyMemberCtx key = iota
)

type MemberHandler struct {
	sessions types.SessionManager
}

func New(
	sessions types.SessionManager,
) *MemberHandler {
	// Init

	return &MemberHandler{
		sessions: sessions,
	}
}

func (h *MemberHandler) Route(r chi.Router) {
	
	r.With(auth.AdminsOnly).Group(func(r chi.Router) {
		r.Get("/", h.memberList)

		r.Post("/", h.memberCreate)
		r.Get("/{memberId}/", h.memberRead)
		r.Post("/{memberId}/", h.memberUpdate)
		r.Delete("/{memberId}/", h.memberDelete)
	})

}

func SetMember(r *http.Request, session types.Session) *http.Request {
	ctx := context.WithValue(r.Context(), keyMemberCtx, session)
	return r.WithContext(ctx)
}

func GetMember(r *http.Request) types.Session {
	return r.Context().Value(keyMemberCtx).(types.Session)
}

func (h *MemberHandler) ExtractMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberId := chi.URLParam(r, "memberId")
		session, ok := h.sessions.Get(memberId)
		if !ok {
			utils.HttpNotFound(w, "Member was not found.")
		} else {
			next.ServeHTTP(w, SetMember(r, session))
		}
	})
}
