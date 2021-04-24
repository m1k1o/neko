package members

import (
	"net/http"
	"strconv"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
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

func (h *MembersHandler) membersList(w http.ResponseWriter, r *http.Request) {
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
		utils.HttpInternalServerError(w, err)
		return
	}

	members := []MemberDataPayload{}
	for id, profile := range entries {
		members = append(members, MemberDataPayload{
			ID:      id,
			Profile: profile,
		})
	}

	utils.HttpSuccess(w, members)
}

func (h *MembersHandler) membersCreate(w http.ResponseWriter, r *http.Request) {
	data := &MemberCreatePayload{
		// default values
		Profile: types.MemberProfile{
			IsAdmin:            false,
			CanLogin:           true,
			CanConnect:         true,
			CanWatch:           true,
			CanHost:            true,
			CanAccessClipboard: true,
		},
	}

	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	if data.Username == "" {
		utils.HttpBadRequest(w, "username cannot be empty")
		return
	}

	if data.Password == "" {
		utils.HttpBadRequest(w, "password cannot be empty")
		return
	}

	id, err := h.members.Insert(data.Username, data.Password, data.Profile)
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w, MemberDataPayload{
		ID:      id,
		Profile: data.Profile,
	})
}

func (h *MembersHandler) membersRead(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)
	profile := member.Profile

	utils.HttpSuccess(w, profile)
}

func (h *MembersHandler) membersUpdateProfile(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)
	profile := member.Profile

	if !utils.HttpJsonRequest(w, r, &profile) {
		return
	}

	if err := h.members.UpdateProfile(member.ID, profile); err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}

func (h *MembersHandler) membersUpdatePassword(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)
	data := MemberPasswordPayload{}

	if !utils.HttpJsonRequest(w, r, &data) {
		return
	}

	if err := h.members.UpdatePassword(member.ID, data.Password); err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}

func (h *MembersHandler) membersDelete(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	if err := h.members.Delete(member.ID); err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}
