package members

import (
	"encoding/json"
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
		utils.HttpBadRequest(w).WithInternalErr(err).Msg("unable to read post body")
		return
	}

	header := &MemberBulkUpdatePayload{}
	if err := json.Unmarshal(bytes, &header); err != nil {
		utils.HttpBadRequest(w).WithInternalErr(err).Msg("unable to unmarshal payload")
		return
	}

	for _, memberId := range header.IDs {
		// TODO: Bulk select?
		profile, err := h.members.Select(memberId)
		if err != nil {
			utils.HttpInternalServerError(w, err).WithInternalMsg("unable to select member profile").Msgf("failed to update member %s", memberId)
			return
		}

		body := &MemberBulkUpdatePayload{
			Profile: profile,
		}

		if err := json.Unmarshal(bytes, &body); err != nil {
			utils.HttpBadRequest(w).WithInternalErr(err).Msgf("unable to unmarshal payload for member %s", memberId)
			return
		}

		if err := h.members.UpdateProfile(memberId, body.Profile); err != nil {
			utils.HttpInternalServerError(w, err).WithInternalMsg("unable to update member profile").Msgf("failed to update member %s", memberId)
			return
		}
	}

	utils.HttpSuccess(w)
}
