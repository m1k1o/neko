package webrtc

import (
	"bytes"
	"encoding/binary"

	"github.com/demodesk/neko/internal/webrtc/payload"
	"github.com/demodesk/neko/pkg/types"
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

	// handle cursor move event
	if header.Event == payload.OP_MOVE {
		payload := &payload.Move{}
		if err := binary.Read(buffer, binary.BigEndian, payload); err != nil {
			return err
		}

		x, y := int(payload.X), int(payload.Y)
		if session.IsHost() {
			// handle active cursor movement
			manager.desktop.Move(x, y)
			manager.curPosition.Set(x, y)
		} else {
			// handle inactive cursor movement
			session.SetCursor(types.Cursor{
				X: x,
				Y: y,
			})
		}

		return nil
	}

	// continue only if session is host
	if !session.IsHost() {
		return nil
	}

	switch header.Event {
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
