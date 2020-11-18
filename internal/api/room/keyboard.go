package room

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

type KeyboardLayoutData struct {
	Layout string `json:"layout"`
}

type KeyboardModifiersData struct {
	NumLock    *bool `json:"numlock"`
	CapsLock   *bool `json:"capslock"`
	ScrollLock *bool `json:"scrollock"`
}

func (h *RoomHandler) KeyboardLayoutSet(w http.ResponseWriter, r *http.Request) {
	data := &KeyboardLayoutData{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	h.desktop.SetKeyboardLayout(data.Layout)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) KeyboardModifiersSet(w http.ResponseWriter, r *http.Request) {
	data := &KeyboardModifiersData{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	var NumLock = 0
	if data.NumLock == nil {
		NumLock = -1
	} else if *data.NumLock {
		NumLock = 1
	}

	var CapsLock = 0
	if data.CapsLock == nil {
		CapsLock = -1
	} else if *data.CapsLock {
		CapsLock = 1
	}

	var ScrollLock = 0
	if data.ScrollLock == nil {
		ScrollLock = -1
	} else if *data.ScrollLock {
		ScrollLock = 1
	}

	h.desktop.SetKeyboardModifiers(NumLock, CapsLock, ScrollLock)

	utils.HttpSuccess(w)
}
