package webrtc

import (
	"bytes"
	"encoding/binary"

	"github.com/pion/webrtc/v3"

	"demodesk/neko/internal/types"
)

const (
	OP_MOVE     = 0x01
	OP_SCROLL   = 0x02
	OP_KEY_DOWN = 0x03
	OP_KEY_UP   = 0x04
	OP_BTN_DOWN = 0x05
	OP_BTN_UP   = 0x06
)

type PayloadHeader struct {
	Event  uint8
	Length uint16
}

type PayloadMove struct {
	PayloadHeader
	X uint16
	Y uint16
}

type PayloadScroll struct {
	PayloadHeader
	X int16
	Y int16
}

type PayloadKey struct {
	PayloadHeader
	Key uint32
}

func (manager *WebRTCManagerCtx) handle(msg webrtc.DataChannelMessage, session types.Session) error {
	// add session id to logger context
	logger := manager.logger.With().Str("session_id", session.ID()).Logger()

	buffer := bytes.NewBuffer(msg.Data)
	header := &PayloadHeader{}
	hbytes := make([]byte, 3)

	if _, err := buffer.Read(hbytes); err != nil {
		return err
	}

	if err := binary.Read(bytes.NewBuffer(hbytes), binary.BigEndian, header); err != nil {
		return err
	}

	buffer = bytes.NewBuffer(msg.Data)

	switch header.Event {
	case OP_MOVE:
		payload := &PayloadMove{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		manager.desktop.Move(int(payload.X), int(payload.Y))
		manager.curPosition.Set(int(payload.X), int(payload.Y))
	case OP_SCROLL:
		payload := &PayloadScroll{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		manager.desktop.Scroll(int(payload.X), int(payload.Y))
		logger.Trace().
			Int16("x", payload.X).
			Int16("y", payload.Y).
			Msg("scroll")
	case OP_KEY_DOWN:
		payload := &PayloadKey{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.KeyDown(payload.Key); err != nil {
			logger.Warn().Err(err).Uint32("key", payload.Key).Msg("key down failed")
		} else {
			logger.Trace().Uint32("key", payload.Key).Msg("key down")
		}
	case OP_KEY_UP:
		payload := &PayloadKey{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.KeyUp(payload.Key); err != nil {
			logger.Warn().Err(err).Uint32("key", payload.Key).Msg("key up failed")
		} else {
			logger.Trace().Uint32("key", payload.Key).Msg("key up")
		}
	case OP_BTN_DOWN:
		payload := &PayloadKey{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.ButtonDown(payload.Key); err != nil {
			logger.Warn().Err(err).Uint32("key", payload.Key).Msg("button down failed")
		} else {
			logger.Trace().Uint32("key", payload.Key).Msg("button down")
		}
	case OP_BTN_UP:
		payload := &PayloadKey{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.ButtonUp(payload.Key); err != nil {
			logger.Warn().Err(err).Uint32("key", payload.Key).Msg("button up failed")
		} else {
			logger.Trace().Uint32("key", payload.Key).Msg("button up")
		}
	}

	return nil
}
