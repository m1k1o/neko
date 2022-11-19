package handler

import (
	"errors"
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
	"m1k1o/neko/internal/utils"
)

func (h *MessageHandler) setFileTransferStatus(session types.Session, payload *message.FileTransferStatus) error {
	if !session.Admin() {
		return errors.New(session.Member().Name + " tried to toggle file transfer but they're not admin")
	}
	h.state.SetFileTransferState(payload.Admin, payload.Unpriv)
	err := h.sessions.Broadcast(message.FileTransferStatus{
		Event:  event.FILETRANSFER_STATUS,
		Admin:  payload.Admin,
		Unpriv: payload.Admin && payload.Unpriv,
	}, nil)
	if err != nil {
		return err
	}

	files, err := utils.ListFiles(h.state.FileTransferPath())
	if err != nil {
		return err
	}
	msg := message.FileList{
		Event: event.FILETRANSFER_LIST,
		Cwd:   h.state.FileTransferPath(),
		Files: *files,
	}
	if payload.Unpriv {
		return h.sessions.Broadcast(msg, nil)
	} else {
		return h.sessions.AdminBroadcast(msg, nil)
	}
}

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
