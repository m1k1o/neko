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
	ID                 *string `json:"id"`
	Secret             *string `json:"secret,omitempty"`
	Name               *string `json:"name"`
	IsAdmin            *bool   `json:"is_admin"`
	CanLogin           *bool   `json:"can_login"`
	CanConnect         *bool   `json:"can_connect"`
	CanWatch           *bool   `json:"can_watch"`
	CanHost            *bool   `json:"can_host"`
	CanAccessClipboard *bool   `json:"can_access_clipboard"`
}

func (h *MembersHandler) membersCreate(w http.ResponseWriter, r *http.Request) {
	data := &MemberDataPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	if data.Secret == nil || *data.Secret == "" {
		utils.HttpBadRequest(w, "Secret cannot be empty.")
		return
	}

	if data.Name == nil || *data.Name == "" {
		utils.HttpBadRequest(w, "Name cannot be empty.")
		return
	}

	var ID string
	if data.ID == nil || *data.ID == "" {
		var err error
		if ID, err = utils.NewUID(32); err != nil {
			utils.HttpInternalServerError(w, err)
			return
		}
	} else {
		ID = *data.ID
		if _, ok := h.sessions.Get(ID); ok {
			utils.HttpBadRequest(w, "Member ID already exists.")
			return
		}
	}

	session, err := h.sessions.Create(ID, types.MemberProfile{
		Secret: *data.Secret,
		Name: *data.Name,
		IsAdmin: defaultBool(data.IsAdmin, false),
		CanLogin: defaultBool(data.CanLogin, true),
		CanConnect: defaultBool(data.CanConnect, true),
		CanWatch: defaultBool(data.CanWatch, true),
		CanHost: defaultBool(data.CanHost, true),
		CanAccessClipboard: defaultBool(data.CanAccessClipboard, true),
	})

	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w, MemberCreatePayload{
		ID: session.ID(),
	})
}

func (h *MembersHandler) membersRead(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	// TODO: Join structs?
	// TODO: Ugly.
	utils.HttpSuccess(w, MemberDataPayload{
		Name: func(v string) *string { return &v }(member.Name()),
		IsAdmin: func(v bool) *bool { return &v }(member.IsAdmin()),
		CanLogin: func(v bool) *bool { return &v }(member.CanLogin()),
		CanConnect: func(v bool) *bool { return &v }(member.CanConnect()),
		CanWatch: func(v bool) *bool { return &v }(member.CanWatch()),
		CanHost: func(v bool) *bool { return &v }(member.CanHost()),
		CanAccessClipboard: func(v bool) *bool { return &v }(member.CanAccessClipboard()),
	})
}

func (h *MembersHandler) membersUpdate(w http.ResponseWriter, r *http.Request) {
	data := &MemberDataPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	member := GetMember(r)

	secret := ""
	if data.Secret != nil && *data.Secret != "" {
		secret = *data.Secret
	}

	name := member.Name()
	if data.Name != nil && *data.Name != "" {
		name = *data.Name
	}

	// TODO: Join structs?
	err := h.sessions.Update(member.ID(), types.MemberProfile{
		Secret: secret,
		Name: name,
		IsAdmin: defaultBool(data.IsAdmin, member.IsAdmin()),
		CanLogin: defaultBool(data.CanLogin, member.CanLogin()),
		CanConnect: defaultBool(data.CanConnect, member.CanConnect()),
		CanWatch: defaultBool(data.CanWatch, member.CanWatch()),
		CanHost: defaultBool(data.CanHost, member.CanHost()),
		CanAccessClipboard: defaultBool(data.CanAccessClipboard, member.CanAccessClipboard()),
	})

	if err != nil {
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

func defaultBool(val *bool, def bool) bool {
	if val != nil {
		return *val
	}
	return def
}
