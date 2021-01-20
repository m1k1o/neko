package websocket

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (ws *WebSocketManagerCtx) fileChooserDialogEvents() {
	var file_chooser_dialog_member types.Session

	// when dialog opens, everyone should be notified.
	ws.desktop.OnFileChooserDialogOpened(func() {
		ws.logger.Info().Msg("FileChooserDialog opened")

		host := ws.sessions.GetHost()
		if host == nil {
			ws.logger.Warn().Msg("no host for FileChooserDialog found, closing")
			go ws.desktop.CloseFileChooserDialog()
			return
		}

		file_chooser_dialog_member = host

		go ws.sessions.Broadcast(message.MemberID{
			Event:  event.FILE_CHOOSER_DIALOG_OPENED,
			ID:     host.ID(),
		}, nil)
	})

	// when dialog closes, everyone should be notified.
	ws.desktop.OnFileChooserDialogClosed(func() {
		ws.logger.Info().Msg("FileChooserDialog closed")

		file_chooser_dialog_member = nil

		go ws.sessions.Broadcast(message.MemberID{
			Event:  event.FILE_CHOOSER_DIALOG_CLOSED,
		}, nil)
	})


	// when new user joins, and someone holds dialog, he shouldd be notified about it.
	ws.sessions.OnConnected(func(session types.Session) {
		if file_chooser_dialog_member == nil {
			return
		}

		if err := session.Send(message.MemberID{
			Event:  event.FILE_CHOOSER_DIALOG_OPENED,
			ID:     file_chooser_dialog_member.ID(),
		}); err != nil {
			ws.logger.Warn().
				Str("id", session.ID()).
				Err(err).
				Msgf("could not send event `%s` to session", event.FILE_CHOOSER_DIALOG_OPENED)
		}
	})

	// when user, that holds dialog, disconnects, it should be closed.
	ws.sessions.OnDisconnected(func(session types.Session) {
		if file_chooser_dialog_member == nil {
			return
		}

		if session.ID() != file_chooser_dialog_member.ID() {
			return
		}

		ws.desktop.CloseFileChooserDialog()
	})
}
