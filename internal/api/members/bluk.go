package members

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

type MemberBulkUpdatePayload struct {
	IDs     []string            `json:"ids"`
	Profile types.MemberProfile `json:"profile"`
}

func (h *MembersHandler) membersBulkUpdate(w http.ResponseWriter, r *http.Request) error {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return utils.HttpBadRequest("unable to read post body").WithInternalErr(err)
	}

	header := &MemberBulkUpdatePayload{}
	if err := json.Unmarshal(bytes, &header); err != nil {
		return utils.HttpBadRequest("unable to unmarshal payload").WithInternalErr(err)
	}

	for _, memberId := range header.IDs {
		// TODO: Bulk select?
		profile, err := h.members.Select(memberId)
		if err != nil {
			return utils.HttpInternalServerError().
				WithInternalErr(err).
				WithInternalMsg("unable to select member profile").
				Msgf("failed to update member %s", memberId)
		}

		body := &MemberBulkUpdatePayload{
			Profile: profile,
		}

		if err := json.Unmarshal(bytes, &body); err != nil {
			return utils.HttpBadRequest().
				WithInternalErr(err).
				Msgf("unable to unmarshal payload for member %s", memberId)
		}

		if err := h.members.UpdateProfile(memberId, body.Profile); err != nil {
			return utils.HttpInternalServerError().
				WithInternalErr(err).
				WithInternalMsg("unable to update member profile").
				Msgf("failed to update member %s", memberId)
		}
	}

	return utils.HttpSuccess(w)
}
