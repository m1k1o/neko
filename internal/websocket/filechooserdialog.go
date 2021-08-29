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
		manager.logger.Info().Msg("file chooser dialog opened")

		host := manager.sessions.GetHost()
		if host == nil {
			manager.logger.Warn().Msg("no host for file chooser dialog found, closing")
			go manager.desktop.CloseFileChooserDialog()
			return
		}

		activeSession = host

		go manager.sessions.Broadcast(message.SessionID{
			Event: event.FILE_CHOOSER_DIALOG_OPENED,
			ID:    host.ID(),
		}, nil)
	})

	// when dialog closes, everyone should be notified.
	manager.desktop.OnFileChooserDialogClosed(func() {
		manager.logger.Info().Msg("file chooser dialog closed")

		activeSession = nil

		go manager.sessions.Broadcast(message.SessionID{
			Event: event.FILE_CHOOSER_DIALOG_CLOSED,
		}, nil)
	})

	// when new user joins, and someone holds dialog, he shouldd be notified about it.
	manager.sessions.OnConnected(func(session types.Session) {
		if activeSession == nil {
			return
		}

		logger := manager.logger.With().Str("session_id", session.ID()).Logger()
		logger.Debug().Msg("sending file chooser dialog status to a new session")

		if err := session.Send(message.SessionID{
			Event: event.FILE_CHOOSER_DIALOG_OPENED,
			ID:    activeSession.ID(),
		}); err != nil {
			logger.Warn().Err(err).
				Str("event", event.FILE_CHOOSER_DIALOG_OPENED).
				Msg("could not send event")
		}
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
