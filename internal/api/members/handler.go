package members

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"

	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type key int

const (
	keyMemberCtx key = iota
)

type MembersHandler struct {
	members types.MemberManager
}

func New(
	members types.MemberManager,
) *MembersHandler {
	// Init

	return &MembersHandler{
		members: members,
	}
}

func (h *MembersHandler) Route(r chi.Router) {
	r.Get("/", h.membersList)

	r.With(auth.AdminsOnly).Group(func(r chi.Router) {
		r.Post("/", h.membersCreate)
		r.With(h.ExtractMember).Route("/{memberId}", func(r chi.Router) {
			r.Get("/", h.membersRead)
			r.Post("/", h.membersUpdateProfile)
			r.Post("/password", h.membersUpdatePassword)
			r.Delete("/", h.membersDelete)
		})
	})
}

type MemberData struct {
	ID      string
	Profile types.MemberProfile
}

func SetMember(r *http.Request, session MemberData) *http.Request {
	ctx := context.WithValue(r.Context(), keyMemberCtx, session)
	return r.WithContext(ctx)
}

func GetMember(r *http.Request) MemberData {
	return r.Context().Value(keyMemberCtx).(MemberData)
}

func (h *MembersHandler) ExtractMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberId := chi.URLParam(r, "memberId")
		profile, err := h.members.Select(memberId)
		if err != nil {
			utils.HttpNotFound(w, err)
		} else {
			next.ServeHTTP(w, SetMember(r, MemberData{
				ID:      memberId,
				Profile: profile,
			}))
		}
	})
}
