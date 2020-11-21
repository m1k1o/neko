package members

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

func (h *MembersHandler) membersList(w http.ResponseWriter, r *http.Request) {

	utils.HttpSuccess(w)
}
