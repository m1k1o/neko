package webrtc

import (
	"fmt"
	"math/rand"

	"github.com/pion/webrtc/v2"
)

func (m *WebRTCManager) createVideoTrack(engine webrtc.MediaEngine) (*webrtc.Track, error) {
	var codec *webrtc.RTPCodec
	for _, videoCodec := range engine.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
		if videoCodec.Name == m.videoPipeline.CodecName {
			codec = videoCodec
			break
		}
	}

	if codec == nil || codec.PayloadType == 0 {
		return nil, fmt.Errorf("remote peer does not support %s", m.videoPipeline.CodecName)
	}

	return webrtc.NewTrack(codec.PayloadType, rand.Uint32(), "stream", "stream", codec)
}

func (m *WebRTCManager) createAudioTrack(engine webrtc.MediaEngine) (*webrtc.Track, error) {
	var codec *webrtc.RTPCodec
	for _, videoCodec := range engine.GetCodecsByKind(webrtc.RTPCodecTypeAudio) {
		if videoCodec.Name == m.audioPipeline.CodecName {
			codec = videoCodec
			break
		}
	}

	if codec == nil || codec.PayloadType == 0 {
		return nil, fmt.Errorf("remote peer does not support %s", m.audioPipeline.CodecName)
	}

	return webrtc.NewTrack(codec.PayloadType, rand.Uint32(), "stream", "stream", codec)
}
