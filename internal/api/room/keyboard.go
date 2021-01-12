package room

import (
	"net/http"

	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/types"
)

type KeyboardLayoutData struct {
	Layout string `json:"layout"`
}

type KeyboardModifiersData struct {
	NumLock  *bool `json:"numlock"`
	CapsLock *bool `json:"capslock"`
}

func (h *RoomHandler) keyboardLayoutSet(w http.ResponseWriter, r *http.Request) {
	data := &KeyboardLayoutData{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	h.desktop.SetKeyboardLayout(data.Layout)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardModifiersSet(w http.ResponseWriter, r *http.Request) {
	data := &KeyboardModifiersData{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	h.desktop.SetKeyboardModifiers(types.KeyboardModifiers{
		NumLock: data.NumLock,
		CapsLock: data.CapsLock,
	})
	utils.HttpSuccess(w)
}

func (h *RoomHandler) keyboardModifiersGet(w http.ResponseWriter, r *http.Request) {
	data := h.desktop.GetKeyboardModifiers()

	utils.HttpSuccess(w, KeyboardModifiersData{
		NumLock: data.NumLock,
		CapsLock: data.CapsLock,
	})
}
