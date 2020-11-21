package member

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

func (h *MemberHandler) memberList(w http.ResponseWriter, r *http.Request) {

	utils.HttpSuccess(w)
}
