package room

import (
	"net/http"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

func (h *RoomHandler) keyboardMapSet(w http.ResponseWriter, r *http.Request) error {
	keyboardMap := types.KeyboardMap{}
	if err := utils.HttpJsonRequest(w, r, &keyboardMap); err != nil {
		return err
	}

	err := h.desktop.SetKeyboardMap(keyboardMap)
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardMapGet(w http.ResponseWriter, r *http.Request) error {
	keyboardMap, err := h.desktop.GetKeyboardMap()
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

	h.desktop.SetKeyboardModifiers(keyboardModifiers)
	return utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardModifiersGet(w http.ResponseWriter, r *http.Request) error {
	keyboardModifiers := h.desktop.GetKeyboardModifiers()

	return utils.HttpSuccess(w, keyboardModifiers)
}
