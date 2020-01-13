package webrtc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

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

var debounce = map[int]bool{}

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
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}
		robotgo.Move(int(payload.X), int(payload.Y))
		break
	case 0x02: // MOUSE_UP
		payload := &dataMouseKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}

		if key, ok := keys.Mouse[int(payload.Key)]; ok {
			if !debounce[int(payload.Key)] {
				return nil
			}
			debounce[int(payload.Key)] = false
			robotgo.MouseToggle("up", key)
		} else {
			manager.logger.Warn().Msgf("Unknown MOUSE_DOWN key: %v", payload.Key)
		}
		break
	case 0x03: // MOUSE_DOWN
		payload := &dataMouseKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}

		if key, ok := keys.Mouse[int(payload.Key)]; ok {
			if debounce[int(payload.Key)] {
				return nil
			}
			debounce[int(payload.Key)] = true

			robotgo.MouseToggle("down", key)
		} else {
			manager.logger.Warn().Msgf("Unknown MOUSE_DOWN key: %v", payload.Key)
		}
		break
	case 0x04: // MOUSE_CLK
		payload := &dataMouseKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}

		if key, ok := keys.Mouse[int(payload.Key)]; ok {
			switch int(payload.Key) {
			case keys.MOUSE_WHEEL_DOWN:
				robotgo.Scroll(0, -10)
				break
			case keys.MOUSE_WHEEL_UP:
				robotgo.Scroll(0, 10)
				break
			case keys.MOUSE_WHEEL_LEFT:
				robotgo.Scroll(-10, 0)
				break
			case keys.MOUSE_WHEEL_RIGH:
				robotgo.Scroll(10, 0)
				break
			default:
				robotgo.Click(key, false)
			}
		} else {
			manager.logger.Warn().Msgf("Unknown MOUSE_CLK key: %v", payload.Key)
		}
		break
	case 0x05: // KEY_DOWN
		payload := &dataKeyboardKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}
		if key, ok := keys.Keyboard[int(payload.Key)]; ok {
			if debounce[int(payload.Key)] {
				return nil
			}
			debounce[int(payload.Key)] = true
			robotgo.KeyToggle(key, "down")
		} else {
			manager.logger.Warn().Msgf("Unknown KEY_DOWN key: %v", payload.Key)
		}
		break
	case 0x06: // KEY_UP
		payload := &dataKeyboardKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}
		if key, ok := keys.Keyboard[int(payload.Key)]; ok {
			if !debounce[int(payload.Key)] {
				return nil
			}
			debounce[int(payload.Key)] = false
			robotgo.KeyToggle(key, "up")
		} else {
			manager.logger.Warn().Msgf("Unknown KEY_UP key: %v", payload.Key)
		}
		break
	case 0x07: // KEY_CLK
		payload := &dataKeyboardKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}
		if key, ok := keys.Keyboard[int(payload.Key)]; ok {
			robotgo.KeyTap(key)
		} else {
			manager.logger.Warn().Msgf("Unknown KEY_CLK key: %v", payload.Key)
		}
		break
	}

	return nil
}
