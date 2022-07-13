package room

import (
	"net/http"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

type KeyboardMapData struct {
	Layout  string `json:"layout"`
	Variant string `json:"variant"`
}

type KeyboardModifiersData struct {
	NumLock  *bool `json:"numlock"`
	CapsLock *bool `json:"capslock"`
}

func (h *RoomHandler) keyboardMapSet(w http.ResponseWriter, r *http.Request) error {
	data := &KeyboardMapData{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	err := h.desktop.SetKeyboardMap(types.KeyboardMap{
		Layout:  data.Layout,
		Variant: data.Variant,
	})

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
		Layout:  data.Layout,
		Variant: data.Variant,
	})
}

func (h *RoomHandler) keyboardModifiersSet(w http.ResponseWriter, r *http.Request) error {
	data := &KeyboardModifiersData{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	h.desktop.SetKeyboardModifiers(types.KeyboardModifiers{
		NumLock:  data.NumLock,
		CapsLock: data.CapsLock,
	})

	return utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardModifiersGet(w http.ResponseWriter, r *http.Request) error {
	data := h.desktop.GetKeyboardModifiers()

	return utils.HttpSuccess(w, KeyboardModifiersData{
		NumLock:  data.NumLock,
		CapsLock: data.CapsLock,
	})
}
