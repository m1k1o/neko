package room

import (
	"net/http"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

type KeyboardMapData struct {
	types.KeyboardMap
}

type KeyboardModifiersData struct {
	types.KeyboardModifiers
}

func (h *RoomHandler) keyboardMapSet(w http.ResponseWriter, r *http.Request) error {
	data := &KeyboardMapData{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	err := h.desktop.SetKeyboardMap(data.KeyboardMap)
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardMapGet(w http.ResponseWriter, r *http.Request) error {
	data, err := h.desktop.GetKeyboardMap()
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w, KeyboardMapData{
		KeyboardMap: *data,
	})
}

func (h *RoomHandler) keyboardModifiersSet(w http.ResponseWriter, r *http.Request) error {
	data := &KeyboardModifiersData{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	h.desktop.SetKeyboardModifiers(data.KeyboardModifiers)
	return utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardModifiersGet(w http.ResponseWriter, r *http.Request) error {
	return utils.HttpSuccess(w, KeyboardModifiersData{
		KeyboardModifiers: h.desktop.GetKeyboardModifiers(),
	})
}
