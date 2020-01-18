package webrtc

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"github.com/pion/webrtc/v2"
	"n.eko.moe/neko/internal/hid"
)

const OP_MOVE = 0x01
const OP_SCROLL = 0x02
const OP_KEY_DOWN = 0x03
const OP_KEY_UP = 0x04
const OP_KEY_CLK = 0x05

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
	Key uint16
}

func (m *WebRTCManager) handle(id string, msg webrtc.DataChannelMessage) error {
	if !m.sessions.IsHost(id) {
		return nil
	}

	buffer := bytes.NewBuffer(msg.Data)
	header := &PayloadHeader{}
	hbytes := make([]byte, 3)

	if _, err := buffer.Read(hbytes); err != nil {
		return err
	}

	if err := binary.Read(bytes.NewBuffer(hbytes), binary.LittleEndian, header); err != nil {
		return err
	}

	buffer = bytes.NewBuffer(msg.Data)

	switch header.Event {
	case OP_MOVE:
		payload := &PayloadMove{}
		if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
			return err
		}

		hid.Move(int(payload.X), int(payload.Y))
		break
	case OP_SCROLL:
		payload := &PayloadScroll{}
		if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
			return err
		}

		m.logger.
			Debug().
			Str("x", strconv.Itoa(int(payload.X))).
			Str("y", strconv.Itoa(int(payload.Y))).
			Msg("scroll")

		hid.Scroll(int(payload.X), int(payload.Y))
		break
	case OP_KEY_DOWN:
		payload := &PayloadKey{}
		if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
			return err
		}

		if payload.Key < 8 {
			button, err := hid.ButtonDown(int(payload.Key))
			if err != nil {
				m.logger.Warn().Err(err).Msg("key down failed")
				return nil
			}

			m.logger.Debug().Msgf("button down %s(%d)", button.Name, payload.Key)
		} else {
			key, err := hid.KeyDown(int(payload.Key))
			if err != nil {
				m.logger.Warn().Err(err).Msg("key down failed")
				return nil
			}

			m.logger.Debug().Msgf("key down %s(%d)", key.Name, payload.Key)
		}

		break
	case OP_KEY_UP:
		payload := &PayloadKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}

		if payload.Key < 8 {
			button, err := hid.ButtonUp(int(payload.Key))
			if err != nil {
				m.logger.Warn().Err(err).Msg("button up failed")
				return nil
			}

			m.logger.Debug().Msgf("button up %s(%d)", button.Name, payload.Key)
		} else {
			key, err := hid.KeyUp(int(payload.Key))
			if err != nil {
				m.logger.Warn().Err(err).Msg("keyup failed")
				return nil
			}

			m.logger.Debug().Msgf("key up %s(%d)", key.Name, payload.Key)
		}
		break
	case OP_KEY_CLK:
		// unused
		break
	}

	return nil
}
