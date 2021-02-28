package webrtc

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"github.com/pion/webrtc/v3"
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

func (manager *WebRTCManagerCtx) handle(msg webrtc.DataChannelMessage) error {
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

		manager.logger.
			Debug().
			Str("x", strconv.Itoa(int(payload.X))).
			Str("y", strconv.Itoa(int(payload.Y))).
			Msg("scroll")

		manager.desktop.Scroll(int(payload.X), int(payload.Y))
	case OP_KEY_DOWN:
		payload := &PayloadKey{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.KeyDown(payload.Key); err != nil {
			manager.logger.Warn().Err(err).Msg("key down failed")
			break
		}

		manager.logger.Debug().Msgf("key down %d", payload.Key)
	case OP_KEY_UP:
		payload := &PayloadKey{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.KeyUp(payload.Key); err != nil {
			manager.logger.Warn().Err(err).Msg("key up failed")
			break
		}

		manager.logger.Debug().Msgf("key up %d", payload.Key)
	case OP_BTN_DOWN:
		payload := &PayloadKey{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.ButtonDown(payload.Key); err != nil {
			manager.logger.Warn().Err(err).Msg("button down failed")
			break
		}

		manager.logger.Debug().Msgf("button down %d", payload.Key)
	case OP_BTN_UP:
		payload := &PayloadKey{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.ButtonUp(payload.Key); err != nil {
			manager.logger.Warn().Err(err).Msg("button up failed")
			break
		}

		manager.logger.Debug().Msgf("button up %d", payload.Key)
	}

	return nil
}
