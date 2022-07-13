package members

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

type MemberDataPayload struct {
	ID      string              `json:"id"`
	Profile types.MemberProfile `json:"profile"`
}

type MemberCreatePayload struct {
	Username string              `json:"username"`
	Password string              `json:"password"`
	Profile  types.MemberProfile `json:"profile"`
}

type MemberPasswordPayload struct {
	Password string `json:"password"`
}

func (h *MembersHandler) membersList(w http.ResponseWriter, r *http.Request) error {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		// TODO: Default zero.
		limit = 0
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		// TODO: Default zero.
		offset = 0
	}

	entries, err := h.members.SelectAll(limit, offset)
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	members := []MemberDataPayload{}
	for id, profile := range entries {
		members = append(members, MemberDataPayload{
			ID:      id,
			Profile: profile,
		})
	}

	return utils.HttpSuccess(w, members)
}

func (h *MembersHandler) membersCreate(w http.ResponseWriter, r *http.Request) error {
	data := &MemberCreatePayload{
		// default values
		Profile: types.MemberProfile{
			IsAdmin:               false,
			CanLogin:              true,
			CanConnect:            true,
			CanWatch:              true,
			CanHost:               true,
			CanShareMedia:         true,
			CanAccessClipboard:    true,
			SendsInactiveCursor:   true,
			CanSeeInactiveCursors: true,
		},
	}

	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	if data.Username == "" {
		return utils.HttpBadRequest("username cannot be empty")
	}

	if data.Password == "" {
		return utils.HttpBadRequest("password cannot be empty")
	}

	id, err := h.members.Insert(data.Username, data.Password, data.Profile)
	if err != nil {
		if errors.Is(err, types.ErrMemberAlreadyExists) {
			return utils.HttpUnprocessableEntity("member already exists")
		}

		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w, MemberDataPayload{
		ID:      id,
		Profile: data.Profile,
	})
}

func (h *MembersHandler) membersRead(w http.ResponseWriter, r *http.Request) error {
	member := GetMember(r)
	profile := member.Profile

	return utils.HttpSuccess(w, profile)
}

func (h *MembersHandler) membersUpdateProfile(w http.ResponseWriter, r *http.Request) error {
	member := GetMember(r)
	data := &member.Profile

	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	if err := h.members.UpdateProfile(member.ID, *data); err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w)
}

func (h *MembersHandler) membersUpdatePassword(w http.ResponseWriter, r *http.Request) error {
	member := GetMember(r)
	data := &MemberPasswordPayload{}

	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	if err := h.members.UpdatePassword(member.ID, data.Password); err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w)
}

func (h *MembersHandler) membersDelete(w http.ResponseWriter, r *http.Request) error {
	member := GetMember(r)

	if err := h.members.Delete(member.ID); err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w)
}
