package members

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

func (h *MembersHandler) membersCreate(w http.ResponseWriter, r *http.Request) {

	utils.HttpSuccess(w)
}

func (h *MembersHandler) membersRead(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	utils.HttpSuccess(w, "Your name is " + member.Name() + ".")
}

func (h *MembersHandler) membersUpdate(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	utils.HttpSuccess(w, "Your name is " + member.Name() + ".")
}

func (h *MembersHandler) membersDelete(w http.ResponseWriter, r *http.Request) {
	member := GetMember(r)

	utils.HttpSuccess(w, "Your name is " + member.Name() + ".")
}
