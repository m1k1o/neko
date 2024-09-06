package websocket

import (
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
)

func (manager *WebSocketManagerCtx) fileChooserDialogEvents() {
	var activeSession types.Session

	// when dialog opens, everyone should be notified.
	manager.desktop.OnFileChooserDialogOpened(func() {
		manager.logger.Info().Msg("file chooser dialog opened")

		host, hasHost := manager.sessions.GetHost()
		if !hasHost {
			manager.logger.Warn().Msg("no host for file chooser dialog found, closing")
			go manager.desktop.CloseFileChooserDialog()
			return
		}

		activeSession = host

		go manager.sessions.Broadcast(
			event.FILE_CHOOSER_DIALOG_OPENED,
			message.SessionID{
				ID: host.ID(),
			})
	})

	// when dialog closes, everyone should be notified.
	manager.desktop.OnFileChooserDialogClosed(func() {
		manager.logger.Info().Msg("file chooser dialog closed")

		activeSession = nil

		go manager.sessions.Broadcast(
			event.FILE_CHOOSER_DIALOG_CLOSED,
			message.SessionID{})
	})

	// when new user joins, and someone holds dialog, he shouldd be notified about it.
	manager.sessions.OnConnected(func(session types.Session) {
		if activeSession == nil {
			return
		}

		manager.logger.Debug().Str("session_id", session.ID()).Msg("sending file chooser dialog status to a new session")

		session.Send(
			event.FILE_CHOOSER_DIALOG_OPENED,
			message.SessionID{
				ID: activeSession.ID(),
			})
	})

	// when user, that holds dialog, disconnects, it should be closed.
	manager.sessions.OnDisconnected(func(session types.Session) {
		if activeSession == nil || activeSession != session {
			return
		}

		manager.logger.Info().Msg("file chooser dialog owner left, closing")
		manager.desktop.CloseFileChooserDialog()
	})
}
