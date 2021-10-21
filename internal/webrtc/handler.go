package webrtc

import (
	"bytes"
	"encoding/binary"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/webrtc/payload"
)

func (manager *WebRTCManagerCtx) handle(data []byte, session types.Session) error {
	// add session id to logger context
	logger := manager.logger.With().Str("session_id", session.ID()).Logger()

	//
	// parse header
	//

	buffer := bytes.NewBuffer(data)

	hbytes := make([]byte, 3)
	if _, err := buffer.Read(hbytes); err != nil {
		return err
	}

	header := &payload.Header{}
	if err := binary.Read(bytes.NewBuffer(hbytes), binary.BigEndian, header); err != nil {
		return err
	}

	//
	// parse body
	//

	buffer = bytes.NewBuffer(data)

	switch header.Event {
	case payload.OP_MOVE:
		payload := &payload.Move{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		manager.desktop.Move(int(payload.X), int(payload.Y))
		manager.curPosition.Set(int(payload.X), int(payload.Y))
	case payload.OP_SCROLL:
		payload := &payload.Scroll{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		manager.desktop.Scroll(int(payload.X), int(payload.Y))
		logger.Trace().
			Int16("x", payload.X).
			Int16("y", payload.Y).
			Msg("scroll")
	case payload.OP_KEY_DOWN:
		payload := &payload.Key{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.KeyDown(payload.Key); err != nil {
			logger.Warn().Err(err).Uint32("key", payload.Key).Msg("key down failed")
		} else {
			logger.Trace().Uint32("key", payload.Key).Msg("key down")
		}
	case payload.OP_KEY_UP:
		payload := &payload.Key{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.KeyUp(payload.Key); err != nil {
			logger.Warn().Err(err).Uint32("key", payload.Key).Msg("key up failed")
		} else {
			logger.Trace().Uint32("key", payload.Key).Msg("key up")
		}
	case payload.OP_BTN_DOWN:
		payload := &payload.Key{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		if err := manager.desktop.ButtonDown(payload.Key); err != nil {
			logger.Warn().Err(err).Uint32("key", payload.Key).Msg("button down failed")
		} else {
			logger.Trace().Uint32("key", payload.Key).Msg("button down")
		}
	case payload.OP_BTN_UP:
		payload := &payload.Key{}
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
