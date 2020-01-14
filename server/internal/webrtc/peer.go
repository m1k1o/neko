package webrtc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/pion/webrtc/v2"

	"n.eko.moe/neko/internal/keys"
)

func (manager *WebRTCManager) createPeer(session *session, raw []byte) error {
	payload := messageSDP{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}

	peer, err := manager.api.NewPeerConnection(manager.config)
	if err != nil {
		return err
	}

	_, err = peer.AddTrack(manager.video)
	if err != nil {
		return err
	}

	_, err = peer.AddTrack(manager.audio)
	if err != nil {
		return err
	}

	peer.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  payload.SDP,
		Type: webrtc.SDPTypeOffer,
	})

	answer, err := peer.CreateAnswer(nil)
	if err != nil {
		return err
	}

	if err = peer.SetLocalDescription(answer); err != nil {
		return err
	}

	session.send(messageSDP{
		message{Event: "sdp/reply"},
		answer.SDP,
	})

	session.peer = peer

	peer.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			if err = manager.onData(session, msg); err != nil {
				manager.logger.Warn().Err(err).Msg("onData failed")
			}
		})
	})

	peer.OnConnectionStateChange(func(connectionState webrtc.PeerConnectionState) {
		switch connectionState {
		case webrtc.PeerConnectionStateDisconnected:
		case webrtc.PeerConnectionStateFailed:
			manager.destroy(session)
			break
		case webrtc.PeerConnectionStateConnected:
			manager.logger.Info().Str("ID", session.id).Msg("Peer connected")
			break
		}
	})

	return nil
}

func (manager *WebRTCManager) onData(session *session, msg webrtc.DataChannelMessage) error {
	if manager.controller != session.id {
		return nil
	}

	header := &dataHeader{}
	buffer := bytes.NewBuffer(msg.Data)
	byt := make([]byte, 3)
	_, err := buffer.Read(byt)
	if err != nil {
		return err
	}

	err = binary.Read(bytes.NewBuffer(byt), binary.LittleEndian, header)
	if err != nil {
		return err
	}

	buffer = bytes.NewBuffer(msg.Data)

	switch header.Event {
	case 0x01: // MOUSE_MOVE
		payload := &dataMouseMove{}
		if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
			return err
		}

		robotgo.Move(int(payload.X), int(payload.Y))
		break
	case 0x02: // MOUSE_UP
		payload := &dataMouseKey{}
		if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
			return err
		}

		code := int(payload.Key)
		if key, ok := keys.Mouse[code]; ok {
			if _, ok := manager.debounce[code]; !ok {
				return nil
			}
			delete(manager.debounce, code)
			robotgo.MouseToggle("up", key)
			manager.logger.Debug().Msgf("MOUSE_UP key: %v (%v)", code, key)
		} else {
			manager.logger.Warn().Msgf("Unknown MOUSE_UP key: %v", code)
		}
		break
	case 0x03: // MOUSE_DOWN
		payload := &dataMouseKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}

		code := int(payload.Key)
		if key, ok := keys.Mouse[code]; ok {
			if _, ok := manager.debounce[code]; ok {
				return nil
			}

			manager.debounce[code] = time.Now()
			robotgo.MouseToggle("down", key)
			manager.logger.Debug().Msgf("MOUSE_DOWN key: %v (%v)", code, key)
		} else {
			manager.logger.Warn().Msgf("Unknown MOUSE_DOWN key: %v", code)
		}
		break
	case 0x04: // MOUSE_CLK
		payload := &dataMouseKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}

		code := int(payload.Key)
		if key, ok := keys.Mouse[code]; ok {
			switch code {
			case keys.MOUSE_WHEEL_DOWN:
				robotgo.Scroll(0, -1)
				break
			case keys.MOUSE_WHEEL_UP:
				robotgo.Scroll(0, 1)
				break
			case keys.MOUSE_WHEEL_LEFT:
				robotgo.Scroll(-1, 0)
				break
			case keys.MOUSE_WHEEL_RIGH:
				robotgo.Scroll(1, 0)
				break
			default:
				robotgo.Click(key, false)
			}

			manager.logger.Debug().Msgf("MOUSE_CLK key: %v (%v)", code, key)
		} else {
			manager.logger.Warn().Msgf("Unknown MOUSE_CLK key: %v", code)
		}
		break
	case 0x05: // KEY_DOWN
		payload := &dataKeyboardKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}

		code := int(payload.Key)
		if key, ok := keys.Keyboard[code]; ok {
			if _, ok := manager.debounce[code]; ok {
				return nil
			}
			manager.debounce[code] = time.Now()
			robotgo.KeyToggle(key, "down")
			manager.logger.Debug().Msgf("KEY_DOWN key: %v (%v)", code, key)
		} else {
			manager.logger.Warn().Msgf("Unknown KEY_DOWN key: %v", code)
		}
		break
	case 0x06: // KEY_UP
		payload := &dataKeyboardKey{}
		if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
			return err
		}

		code := int(payload.Key)
		if key, ok := keys.Keyboard[code]; ok {
			if _, ok := manager.debounce[code]; !ok {
				return nil
			}
			delete(manager.debounce, code)
			robotgo.KeyToggle(key, "up")
			manager.logger.Debug().Msgf("KEY_UP key: %v (%v)", code, key)
		} else {
			manager.logger.Warn().Msgf("Unknown KEY_UP key: %v", code)
		}
		break
	case 0x07: // KEY_CLK
		payload := &dataKeyboardKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}

		code := int(payload.Key)
		if key, ok := keys.Keyboard[code]; ok {
			robotgo.KeyTap(key)
			manager.logger.Debug().Msgf("KEY_CLK key: %v (%v)", code, key)
		} else {
			manager.logger.Warn().Msgf("Unknown KEY_CLK key: %v", code)
		}
		break
	}

	return nil
}

func (manager *WebRTCManager) clearKeys() {
	for code := range manager.debounce {
		if key, ok := keys.Keyboard[code]; ok {
			robotgo.MouseToggle(key, "up")
		}

		if key, ok := keys.Mouse[code]; ok {
			robotgo.KeyToggle(key, "up")
		}

		delete(manager.debounce, code)
	}
}

func (manager *WebRTCManager) checkKeys() {
	t := time.Now()
	for code, start := range manager.debounce {
		if t.Sub(start) < (time.Second * 10) {
			continue
		}

		if key, ok := keys.Keyboard[code]; ok {
			robotgo.MouseToggle(key, "up")
		}

		if key, ok := keys.Mouse[code]; ok {
			robotgo.KeyToggle(key, "up")
		}

		delete(manager.debounce, code)
	}
}
