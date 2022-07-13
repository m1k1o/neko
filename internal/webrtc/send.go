package webrtc

import (
	"bytes"
	"encoding/binary"

	"github.com/demodesk/neko/internal/webrtc/payload"
	"github.com/demodesk/neko/pkg/types"
)

func (peer *WebRTCPeerCtx) SendCursorPosition(x, y int) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.dataChannel == nil {
		return types.ErrWebRTCDataChannelNotFound
	}

	data := payload.CursorPosition{
		Header: payload.Header{
			Event:  payload.OP_CURSOR_POSITION,
			Length: 7,
		},
		X: uint16(x),
		Y: uint16(y),
	}

	buffer := &bytes.Buffer{}
	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return err
	}

	return peer.dataChannel.Send(buffer.Bytes())
}

func (peer *WebRTCPeerCtx) SendCursorImage(cur *types.CursorImage, img []byte) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.dataChannel == nil {
		return types.ErrWebRTCDataChannelNotFound
	}

	data := payload.CursorImage{
		Header: payload.Header{
			Event:  payload.OP_CURSOR_IMAGE,
			Length: uint16(11 + len(img)),
		},
		Width:  cur.Width,
		Height: cur.Height,
		Xhot:   cur.Xhot,
		Yhot:   cur.Yhot,
	}

	buffer := &bytes.Buffer{}

	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, img); err != nil {
		return err
	}

	return peer.dataChannel.Send(buffer.Bytes())
}
