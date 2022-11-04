package handler

import (
	"errors"
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
	"m1k1o/neko/internal/utils"
)

func (h *MessageHandler) refresh(session types.Session) error {
	if !(h.state.FileTransferEnabled() && session.Admin() || h.state.UnprivFileTransferEnabled()) {
		return errors.New(session.Member().Name + " tried to refresh file list when they can't")
	}

	files, err := utils.ListFiles(h.state.FileTransferPath())
	if err != nil {
		return err
	}
	return session.Send(
		message.FileList{
			Event: event.FILETRANSFER_LIST,
			Cwd:   h.state.FileTransferPath(),
			Files: *files,
		})
}
