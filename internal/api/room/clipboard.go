package room

import (
	"bytes"
	"strings"
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

func (h *RoomHandler) clipboardGetImage(w http.ResponseWriter, r *http.Request) {
	bytes, err := h.desktop.ClipboardGetBinary("image/png")
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "image/png")
	//nolint
	w.Write(bytes)
}

func (h *RoomHandler) clipboardSetImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(MAX_UPLOAD_SIZE)
	if err != nil {
		utils.HttpBadRequest(w, "Failed to parse multipart form.")
		return
	}

	//nolint
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.HttpBadRequest(w, "No file received.")
		return
	}

	defer file.Close()

	mime := header.Header.Get("Content-Type")
	if !strings.HasPrefix(mime, "image/") {
		utils.HttpBadRequest(w, "File must be image.")
		return
	}

	buffer := new(bytes.Buffer)
	buffer.ReadFrom(file)

	err = h.desktop.ClipboardSetBinary("image/png", buffer.Bytes())
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}
