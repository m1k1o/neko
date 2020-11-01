package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) signalProvide(session types.Session) error {
	sdp, lite, ice, err := h.webrtc.CreatePeer(session)
	if err != nil {
		return err
	}

	if err := session.Send(message.SignalProvide{
		Event: event.SIGNAL_PROVIDE,
		ID:    session.ID(),
		SDP:   sdp,
		Lite:  lite,
		ICE:   ice,
	}); err != nil {
		return err
	}

	return nil
}

func (h *MessageHandlerCtx) signalAnswer(session types.Session, payload *message.SignalAnswer) error {
	session.SetName(payload.DisplayName)

	if err := session.SignalAnswer(payload.SDP); err != nil {
		return err
	}

	return nil
}
