package room

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

type ClipboardPayload struct {
	Text string `json:"text,omitempty"`
	HTML string `json:"html,omitempty"`
}

func (h *RoomHandler) clipboardGetTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := h.desktop.ClipboardGetTargets()
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w, targets)
}

func (h *RoomHandler) clipboardGetPlainText(w http.ResponseWriter, r *http.Request) {
	text, err := h.desktop.ClipboardGetPlainText()
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w, ClipboardPayload{
		Text: text,
	})
}

func (h *RoomHandler) clipboardSetPlainText(w http.ResponseWriter, r *http.Request) {
	data := &ClipboardPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	err := h.desktop.ClipboardSetPlainText(data.Text)
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}

func (h *RoomHandler) clipboardGetRichText(w http.ResponseWriter, r *http.Request) {
	html, err := h.desktop.ClipboardGetRichText()
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w, ClipboardPayload{
		HTML: html,
	})
}

func (h *RoomHandler) clipboardSetRichText(w http.ResponseWriter, r *http.Request) {
	data := &ClipboardPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	err := h.desktop.ClipboardSetRichText(data.HTML)
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}
