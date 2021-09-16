package room

import (
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	"demodesk/neko/internal/utils"
)

// TODO: Extract file uploading to custom utility.

// maximum upload size of 32 MB
const maxUploadSize = 32 << 20

func (h *RoomHandler) uploadDrop(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		utils.HttpBadRequest(w).WithInternalErr(err).Msg("failed to parse multipart form")
		return
	}

	//nolint
	defer r.MultipartForm.RemoveAll()

	X, err := strconv.Atoi(r.FormValue("x"))
	if err != nil {
		utils.HttpBadRequest(w).WithInternalErr(err).Msg("no X coordinate received")
		return
	}

	Y, err := strconv.Atoi(r.FormValue("y"))
	if err != nil {
		utils.HttpBadRequest(w).WithInternalErr(err).Msg("no Y coordinate received")
		return
	}

	req_files := r.MultipartForm.File["files"]
	if len(req_files) == 0 {
		utils.HttpBadRequest(w).Msg("no files received")
		return
	}

	dir, err := os.MkdirTemp("", "neko-drop-*")
	if err != nil {
		utils.HttpInternalServerError(w, err).WithInternalMsg("unable to create temporary directory").Send()
		return
	}

	files := []string{}
	for _, req_file := range req_files {
		path := path.Join(dir, req_file.Filename)

		srcFile, err := req_file.Open()
		if err != nil {
			utils.HttpInternalServerError(w, err).WithInternalMsg("unable to open uploaded file").Send()
			return
		}

		defer srcFile.Close()

		dstFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			utils.HttpInternalServerError(w, err).WithInternalMsg("unable to open destination file").Send()
			return
		}

		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			utils.HttpInternalServerError(w, err).WithInternalMsg("unable to copy uploaded file to destination file").Send()
			return
		}

		files = append(files, path)
	}

	if !h.desktop.DropFiles(X, Y, files) {
		utils.HttpInternalServerError(w, nil).WithInternalMsg("unable to drop files").Send()
		return
	}

	utils.HttpSuccess(w)
}

func (h *RoomHandler) uploadDialogPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		utils.HttpBadRequest(w).WithInternalErr(err).Msg("failed to parse multipart form")
		return
	}

	//nolint
	defer r.MultipartForm.RemoveAll()

	if !h.desktop.IsFileChooserDialogOpened() {
		utils.HttpUnprocessableEntity(w).Msg("file chooser dialog is not open")
		return
	}

	req_files := r.MultipartForm.File["files"]
	if len(req_files) == 0 {
		utils.HttpInternalServerError(w, err).WithInternalMsg("unable to copy uploaded file to destination file").Send()
		return
	}

	dir, err := os.MkdirTemp("", "neko-dialog-*")
	if err != nil {
		utils.HttpInternalServerError(w, err).WithInternalMsg("unable to create temporary directory").Send()
		return
	}

	for _, req_file := range req_files {
		path := path.Join(dir, req_file.Filename)

		srcFile, err := req_file.Open()
		if err != nil {
			utils.HttpInternalServerError(w, err).WithInternalMsg("unable to open uploaded file").Send()
			return
		}

		defer srcFile.Close()

		dstFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			utils.HttpInternalServerError(w, err).WithInternalMsg("unable to open destination file").Send()
			return
		}

		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			utils.HttpInternalServerError(w, err).WithInternalMsg("unable to copy uploaded file to destination file").Send()
			return
		}
	}

	if err := h.desktop.HandleFileChooserDialog(dir); err != nil {
		utils.HttpInternalServerError(w, err).WithInternalMsg("unable to handle file chooser dialog").Send()
		return
	}

	utils.HttpSuccess(w)
}

func (h *RoomHandler) uploadDialogClose(w http.ResponseWriter, r *http.Request) {
	if !h.desktop.IsFileChooserDialogOpened() {
		utils.HttpUnprocessableEntity(w).Msg("file chooser dialog is not open")
		return
	}

	h.desktop.CloseFileChooserDialog()
	utils.HttpSuccess(w)
}
