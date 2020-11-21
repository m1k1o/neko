package members

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

type MemberCreatePayload struct {
	ID    string  `json:"id"`
}

type MemberDataPayload struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Admin bool    `json:"admin"`
}

func (h *MembersHandler) membersCreate(w http.ResponseWriter, r *http.Request) {
	data := &MemberDataPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	utils.HttpSuccess(w, MemberCreatePayload{
		ID: "some_id",
	})
}

func (h *MembersHandler) membersRead(w http.ResponseWriter, r *http.Request) {
	data := &MemberDataPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	member := GetMember(r)

	utils.HttpSuccess(w, MemberDataPayload{
		ID: member.ID(),
		Name: member.Name(),
		Admin: member.Admin(),
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
		Admin: member.Admin(),
	})
}

func (h *MembersHandler) membersDelete(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	utils.HttpSuccess(w, MemberDataPayload{
		ID: member.ID(),
		Name: member.Name(),
		Admin: member.Admin(),
	})
}
