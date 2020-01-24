package webrtc

import (
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"n.eko.moe/neko/internal/types"
)

type Peer struct {
	id         string
	engine     webrtc.MediaEngine
	api        *webrtc.API
	video      *webrtc.Track
	audio      *webrtc.Track
	connection *webrtc.PeerConnection
}

func (peer *Peer) WriteAudioSample(sample types.Sample) error {
	if err := peer.audio.WriteSample(media.Sample(sample)); err != nil {
		return err
	}
	return nil
}

func (peer *Peer) WriteVideoSample(sample types.Sample) error {
	if err := peer.video.WriteSample(media.Sample(sample)); err != nil {
		return err
	}
	return nil
}

func (peer *Peer) WriteData(v interface{}) error {
	return nil
}

func (peer *Peer) Destroy() error {
	if peer.connection != nil && peer.connection.ConnectionState() == webrtc.PeerConnectionStateConnected {
		if err := peer.connection.Close(); err != nil {
			return err
		}
	}

	return nil
}
