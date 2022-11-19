package handler

import (
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
	"m1k1o/neko/internal/utils"
)

func (h *MessageHandler) FileTransferRefresh(session types.Session) error {
	if !h.state.FileTransferEnabled() {
		return nil
	}

	fileTransferPath := h.state.FileTransferPath("") // root

	// allow users only if file transfer is not locked
	if session != nil && !(session.Admin() || !h.state.IsLocked("file_transfer")) {
		h.logger.Debug().Msg("file transfer is locked for users")
		return nil
	}

	// TODO: keep list of files in memory and update it on file changes
	files, err := utils.ListFiles(fileTransferPath)
	if err != nil {
		return err
	}

	message := message.FileTransferList{
		Event: event.FILETRANSFER_LIST,
		Cwd:   fileTransferPath,
		Files: files,
	}

	// send to just one user
	if session != nil {
		return session.Send(message)
	}

	// broadcast to all admins
	if h.state.IsLocked("file_transfer") {
		return h.sessions.AdminBroadcast(message, nil)
	}

	// broadcast to all users
	return h.sessions.Broadcast(message, nil)
}
