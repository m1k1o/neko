package members

import (
	"net/http"

	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/types"
)

type MemberDataPayload struct {
	ID string `json:"id"`
	*types.MemberProfile
}

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

func (h *MembersHandler) membersCreate(w http.ResponseWriter, r *http.Request) {
	data := &MemberDataPayload{
		// default values
		MemberProfile: &types.MemberProfile{
			IsAdmin: false,
			CanLogin: true,
			CanConnect: true,
			CanWatch: true,
			CanHost: true,
			CanAccessClipboard: true,
		},
	}

	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	if data.Secret == "" {
		utils.HttpBadRequest(w, "Secret cannot be empty.")
		return
	}

	if data.Name == "" {
		utils.HttpBadRequest(w, "Name cannot be empty.")
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

	session, err := h.sessions.Create(data.ID, *data.MemberProfile)
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w, MemberDataPayload{
		ID: session.ID(),
	})
}

func (h *MembersHandler) membersRead(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)
	profile := member.GetProfile()

	utils.HttpSuccess(w, profile)
}

func (h *MembersHandler) membersUpdate(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)
	profile := member.GetProfile()

	if !utils.HttpJsonRequest(w, r, &profile) {
		return
	}

	if err := h.sessions.Update(member.ID(), profile); err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}

func (h *MembersHandler) membersDelete(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	if err := h.sessions.Delete(member.ID()); err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}
