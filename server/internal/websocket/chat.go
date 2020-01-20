package websocket

import (
	"n.eko.moe/neko/internal/event"
	"n.eko.moe/neko/internal/message"
	"n.eko.moe/neko/internal/session"
)

func (h *MessageHandler) chat(id string, session *session.Session, payload *message.ChatRecieve) error {
	if err := h.sessions.Brodcast(
		message.ChatSend{
			Event:   event.CHAT_MESSAGE,
			Content: payload.Content,
			ID:      id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.CONTROL_RELEASE)
		return err
	}
	return nil
}

func (h *MessageHandler) chatEmoji(id string, session *session.Session, payload *message.EmojiRecieve) error {
	if err := h.sessions.Brodcast(
		message.EmojiSend{
			Event: event.CHAT_MESSAGE,
			Emoji: payload.Emoji,
			ID:    id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.CONTROL_RELEASE)
		return err
	}
	return nil
}
