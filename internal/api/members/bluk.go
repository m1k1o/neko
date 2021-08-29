package members

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type MemberBulkUpdatePayload struct {
	IDs     []string            `json:"ids"`
	Profile types.MemberProfile `json:"profile"`
}

func (h *MembersHandler) membersBulkUpdate(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	header := &MemberBulkUpdatePayload{}
	if err := json.Unmarshal(bytes, &header); err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	for _, memberId := range header.IDs {
		// TODO: Bulk select?
		profile, err := h.members.Select(memberId)
		if err != nil {
			utils.HttpInternalServerError(w, fmt.Sprintf("member %s: %v", memberId, err))
			return
		}

		body := &MemberBulkUpdatePayload{
			Profile: profile,
		}

		if err := json.Unmarshal(bytes, &body); err != nil {
			utils.HttpInternalServerError(w, fmt.Sprintf("member %s: %v", memberId, err))
			return
		}

		if err := h.members.UpdateProfile(memberId, body.Profile); err != nil {
			utils.HttpInternalServerError(w, fmt.Sprintf("member %s: %v", memberId, err))
			return
		}
	}

	utils.HttpSuccess(w)
}
