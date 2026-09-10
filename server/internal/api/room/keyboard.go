package room

import (
	"net/http"

	"github.com/m1k1o/neko/server/pkg/auth"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

func (h *RoomHandler) keyboardMapSet(w http.ResponseWriter, r *http.Request) error {
	keyboardMap := types.KeyboardMap{}
	if err := utils.HttpJsonRequest(w, r, &keyboardMap); err != nil {
		return err
	}

	session, _ := auth.GetSession(r)
	err := h.desktopApp.SetKeyboardMap(session, keyboardMap)
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardMapGet(w http.ResponseWriter, r *http.Request) error {
	keyboardMap, err := h.desktopApp.GetKeyboardMap()
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w, keyboardMap)
}

func (h *RoomHandler) keyboardModifiersSet(w http.ResponseWriter, r *http.Request) error {
	keyboardModifiers := types.KeyboardModifiers{}
	if err := utils.HttpJsonRequest(w, r, &keyboardModifiers); err != nil {
		return err
	}

	session, _ := auth.GetSession(r)
	if err := h.desktopApp.SetKeyboardModifiers(session, keyboardModifiers); err != nil {
		return utils.HttpUnprocessableEntity(err.Error())
	}
	return utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardModifiersGet(w http.ResponseWriter, r *http.Request) error {
	keyboardModifiers := h.desktopApp.GetKeyboardModifiers()

	return utils.HttpSuccess(w, keyboardModifiers)
}
