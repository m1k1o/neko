package members

import (
	"net/http"

	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/types"
)

type MemberCreatePayload struct {
	ID string `json:"id"`
}

type MemberDataPayload struct {
	ID                 string `json:"id"`
	Secret             string `json:"secret,omitempty"`
	Name               string `json:"name"`
	IsAdmin            bool   `json:"is_admin"`
	CanLogin           bool   `json:"can_login"`
	CanConnect         bool   `json:"can_connect"`
	CanWatch           bool   `json:"can_watch"`
	CanHost            bool   `json:"can_host"`
	CanAccessClipboard bool   `json:"can_access_clipboard"`
}

func (h *MembersHandler) membersCreate(w http.ResponseWriter, r *http.Request) {
	data := &MemberDataPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	if data.ID == "" {
		var err error
		if data.ID, err = utils.NewUID(32); err != nil {
			utils.HttpInternalServerError(w, err)
			return
		}
	} else {
		if _, ok := h.sessions.Get(data.ID); ok {
			utils.HttpBadRequest(w, "Member ID already exists.")
			return
		}
	}

	session := h.sessions.Create(data.ID, types.MemberProfile{
		Secret: data.Secret,
		Name: data.Name,
		IsAdmin: data.IsAdmin,
	})

	utils.HttpSuccess(w, MemberCreatePayload{
		ID: session.ID(),
	})
}

func (h *MembersHandler) membersRead(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	utils.HttpSuccess(w, MemberDataPayload{
		ID: member.ID(),
		Name: member.Name(),
		IsAdmin: member.IsAdmin(),
	})
}

func (h *MembersHandler) membersUpdate(w http.ResponseWriter, r *http.Request) {
	data := &MemberDataPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	member := GetMember(r)

	utils.HttpSuccess(w, MemberDataPayload{
		ID: member.ID(),
		Name: member.Name(),
		IsAdmin: member.IsAdmin(),
	})
}

func (h *MembersHandler) membersDelete(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	if err := h.sessions.Delete(member.ID()); err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}
