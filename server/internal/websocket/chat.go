package websocket

import (
	"n.eko.moe/neko/internal/types"
	"n.eko.moe/neko/internal/types/event"
	"n.eko.moe/neko/internal/types/message"
)

func (h *MessageHandler) chat(id string, session types.Session, payload *message.ChatRecieve) error {
	if session.Muted() {
		return nil
	}

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

func (h *MessageHandler) chatEmote(id string, session types.Session, payload *message.EmoteRecieve) error {
	if session.Muted() {
		return nil
	}

	if err := h.sessions.Brodcast(
		message.EmoteSend{
			Event: event.CHAT_EMOTE,
			Emote: payload.Emote,
			ID:    id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.CONTROL_RELEASE)
		return err
	}
	return nil
}
