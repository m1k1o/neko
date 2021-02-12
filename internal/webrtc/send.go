package webrtc

import (
	"bytes"
    "encoding/binary"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

const (
	OP_CURSOR_POSITION = 0x01
	OP_CURSOR_IMAGE = 0x02
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
	Uri    string
}

func (webrtc_peer *WebRTCPeerCtx) SendCursorPosition(x, y int) error {
	if webrtc_peer.dataChannel == nil {
		return nil
	}

	data := PayloadCursorPosition{
		PayloadHeader: PayloadHeader{
			Event: OP_CURSOR_POSITION,
			Length: 7,
		},
		X: uint16(x),
		Y: uint16(y),
	}

	var buffer bytes.Buffer
	binary.Write(&buffer, binary.LittleEndian, data)
	return webrtc_peer.dataChannel.Send(buffer.Bytes())
}

func (webrtc_peer *WebRTCPeerCtx) SendCursorImage(cur *types.CursorImage) error {
	if webrtc_peer.dataChannel == nil {
		return nil
	}

	uri, err := utils.GetCursorImageURI(cur)
	if err != nil {
		return err
	}

	data := PayloadCursorImage{
		PayloadHeader: PayloadHeader{
			Event: OP_CURSOR_IMAGE,
			Length: uint16(11 + len(uri)),
		},
		Width: cur.Width,
		Height: cur.Height,
		Xhot: cur.Xhot,
		Yhot: cur.Yhot,
		Uri: uri,
	}

	var buffer bytes.Buffer
	binary.Write(&buffer, binary.LittleEndian, data)
	return webrtc_peer.dataChannel.Send(buffer.Bytes())
}
