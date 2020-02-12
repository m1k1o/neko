package webrtc

import (
	"fmt"
	"math/rand"

	"github.com/pion/webrtc/v2"
)

func (m *WebRTCManager) createVideoTrack() (*webrtc.Track, error) {
	var codec *webrtc.RTPCodec
	for _, videoCodec := range m.engine.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
		if videoCodec.Name == m.videoPipeline.CodecName {
			codec = videoCodec
			break
		}
	}

	if codec == nil || codec.PayloadType == 0 {
		return nil, fmt.Errorf("remote peer does not support video codec %s", m.videoPipeline.CodecName)
	}

	return webrtc.NewTrack(codec.PayloadType, rand.Uint32(), "stream", "stream", codec)
}

func (m *WebRTCManager) createAudioTrack() (*webrtc.Track, error) {
	var codec *webrtc.RTPCodec
	for _, videoCodec := range m.engine.GetCodecsByKind(webrtc.RTPCodecTypeAudio) {
		if videoCodec.Name == m.audioPipeline.CodecName {
			codec = videoCodec
			break
		}
	}

	if codec == nil || codec.PayloadType == 0 {
		return nil, fmt.Errorf("remote peer does not support audio codec %s", m.audioPipeline.CodecName)
	}

	return webrtc.NewTrack(codec.PayloadType, rand.Uint32(), "stream", "stream", codec)
}
