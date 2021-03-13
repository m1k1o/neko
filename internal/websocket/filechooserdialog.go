package websocket

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (manager *WebSocketManagerCtx) fileChooserDialogEvents() {
	var activeSession types.Session

	// when dialog opens, everyone should be notified.
	manager.desktop.OnFileChooserDialogOpened(func() {
		manager.logger.Info().Msg("FileChooserDialog opened")

		host := manager.sessions.GetHost()
		if host == nil {
			manager.logger.Warn().Msg("no host for FileChooserDialog found, closing")
			go manager.desktop.CloseFileChooserDialog()
			return
		}

		activeSession = host

		go manager.sessions.Broadcast(message.MemberID{
			Event: event.FILE_CHOOSER_DIALOG_OPENED,
			ID:    host.ID(),
		}, nil)
	})

	// when dialog closes, everyone should be notified.
	manager.desktop.OnFileChooserDialogClosed(func() {
		manager.logger.Info().Msg("FileChooserDialog closed")

		activeSession = nil

		go manager.sessions.Broadcast(message.MemberID{
			Event: event.FILE_CHOOSER_DIALOG_CLOSED,
		}, nil)
	})

	// when new user joins, and someone holds dialog, he shouldd be notified about it.
	manager.sessions.OnConnected(func(session types.Session) {
		if activeSession == nil {
			return
		}

		if err := session.Send(message.MemberID{
			Event: event.FILE_CHOOSER_DIALOG_OPENED,
			ID:    activeSession.ID(),
		}); err != nil {
			manager.logger.Warn().
				Str("session_id", session.ID()).
				Err(err).
				Msgf("could not send event `%s` to session", event.FILE_CHOOSER_DIALOG_OPENED)
		}
	})

	// when user, that holds dialog, disconnects, it should be closed.
	manager.sessions.OnDisconnected(func(session types.Session) {
		if activeSession == nil || activeSession != session {
			return
		}

		manager.desktop.CloseFileChooserDialog()
	})
}
