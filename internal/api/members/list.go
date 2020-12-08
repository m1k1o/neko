package members

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

func (h *MembersHandler) membersList(w http.ResponseWriter, r *http.Request) {
	members := []MemberDataPayload{}
	for _, session := range h.sessions.Members() {
		profile := session.GetProfile()
		members = append(members, MemberDataPayload{
			ID:            session.ID(),
			MemberProfile: &profile,
		})
	}

	utils.HttpSuccess(w, members)
}
