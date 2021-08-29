package webrtc

import (
	"bytes"
	"encoding/binary"

	"demodesk/neko/internal/types"
)

const (
	OP_CURSOR_POSITION = 0x01
	OP_CURSOR_IMAGE    = 0x02
)

type PayloadCursorPosition struct {
	PayloadHeader
	X uint16
	Y uint16
}

type PayloadCursorImage struct {
	PayloadHeader
	Width  uint16
	Height uint16
	Xhot   uint16
	Yhot   uint16
}

func (peer *WebRTCPeerCtx) SendCursorPosition(x, y int) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.dataChannel == nil {
		return types.ErrWebRTCDataChannelNotFound
	}

	data := PayloadCursorPosition{
		PayloadHeader: PayloadHeader{
			Event:  OP_CURSOR_POSITION,
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

	data := PayloadCursorImage{
		PayloadHeader: PayloadHeader{
			Event:  OP_CURSOR_IMAGE,
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
