package member

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

func (h *MemberHandler) memberCreate(w http.ResponseWriter, r *http.Request) {

	utils.HttpSuccess(w)
}

func (h *MemberHandler) memberRead(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	utils.HttpSuccess(w, "Your name is " + member.Name() + ".")
}

func (h *MemberHandler) memberUpdate(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	utils.HttpSuccess(w, "Your name is " + member.Name() + ".")
}

func (h *MemberHandler) memberDelete(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	utils.HttpSuccess(w, "Your name is " + member.Name() + ".")
}
