package members

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

type MembersHandler struct {
	sessions types.SessionManager
}

func New(
	sessions types.SessionManager,
) *MembersHandler {
	// Init

	return &MembersHandler{
		sessions: sessions,
	}
}

func (h *MembersHandler) Route(r chi.Router) {
	r.Get("/", h.membersList)

	r.With(auth.AdminsOnly).Group(func(r chi.Router) {
		r.Post("/", h.membersCreate)
		r.With(h.ExtractMember).Route("/{memberId}", func(r chi.Router) {
			r.Get("/", h.membersRead)
			r.Post("/", h.membersUpdate)
			r.Delete("/", h.membersDelete)
		})
	})
}

func SetMember(r *http.Request, session types.Session) *http.Request {
	ctx := context.WithValue(r.Context(), keyMemberCtx, session)
	return r.WithContext(ctx)
}

func GetMember(r *http.Request) types.Session {
	return r.Context().Value(keyMemberCtx).(types.Session)
}

func (h *MembersHandler) ExtractMember(next http.Handler) http.Handler {
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
