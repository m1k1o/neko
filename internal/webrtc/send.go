package webrtc

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
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
	if peer.dataChannel == nil {
		return fmt.Errorf("no data channel")
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

func (peer *WebRTCPeerCtx) SendCursorImage(cur *types.CursorImage) error {
	if peer.dataChannel == nil {
		return fmt.Errorf("no data channel")
	}

	img, err := utils.GetCursorImage(cur)
	if err != nil {
		return err
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
