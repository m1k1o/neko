package room

import (
	"net/http"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type KeyboardMapData struct {
	Layout  string `json:"layout"`
	Variant string `json:"variant"`
}

type KeyboardModifiersData struct {
	NumLock  *bool `json:"numlock"`
	CapsLock *bool `json:"capslock"`
}

func (h *RoomHandler) keyboardMapSet(w http.ResponseWriter, r *http.Request) {
	data := &KeyboardMapData{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	err := h.desktop.SetKeyboardMap(types.KeyboardMap{
		Layout:  data.Layout,
		Variant: data.Variant,
	})

	if err != nil {
		utils.HttpInternalServerError(w, "Unable to change keyboard map.")
		return
	}

	utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardMapGet(w http.ResponseWriter, r *http.Request) {
	data, err := h.desktop.GetKeyboardMap()

	if err != nil {
		utils.HttpInternalServerError(w, "Unable to get keyboard map.")
		return
	}

	utils.HttpSuccess(w, KeyboardMapData{
		Layout:  data.Layout,
		Variant: data.Variant,
	})
}

func (h *RoomHandler) keyboardModifiersSet(w http.ResponseWriter, r *http.Request) {
	data := &KeyboardModifiersData{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	h.desktop.SetKeyboardModifiers(types.KeyboardModifiers{
		NumLock:  data.NumLock,
		CapsLock: data.CapsLock,
	})
	utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardModifiersGet(w http.ResponseWriter, r *http.Request) {
	data := h.desktop.GetKeyboardModifiers()

	utils.HttpSuccess(w, KeyboardModifiersData{
		NumLock:  data.NumLock,
		CapsLock: data.CapsLock,
	})
}
