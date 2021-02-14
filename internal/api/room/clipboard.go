package room

import (
	// TODO: Unused now.
	//"bytes"
	//"strings"
	"net/http"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type ClipboardPayload struct {
	Text string `json:"text,omitempty"`
	HTML string `json:"html,omitempty"`
}

func (h *RoomHandler) clipboardGetText(w http.ResponseWriter, r *http.Request) {
	data, err := h.desktop.ClipboardGetText()
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w, ClipboardPayload{
		Text: data.Text,
		HTML: data.HTML,
	})
}

func (h *RoomHandler) clipboardSetText(w http.ResponseWriter, r *http.Request) {
	data := &ClipboardPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	err := h.desktop.ClipboardSetText(types.ClipboardText{
		Text: data.Text,
		HTML: data.HTML,
	})

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

/* TODO: Unused now.
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
	_, err = buffer.ReadFrom(file)
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	err = h.desktop.ClipboardSetBinary("image/png", buffer.Bytes())
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}

func (h *RoomHandler) clipboardGetTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := h.desktop.ClipboardGetTargets()
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w, targets)
}
*/
