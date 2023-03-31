package config

import (
	"m1k1o/neko/internal/types/codec"
	"strings"

	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type HwEnc int

const (
	HwEncNone HwEnc = iota
	HwEncVAAPI
	HwEncNVENC
)

type Capture struct {
	// video
	Display       string
	VideoCodec    codec.RTPCodec
	VideoHWEnc    HwEnc // TODO: Pipeline builder.
	VideoBitrate  uint  // TODO: Pipeline builder.
	VideoMaxFPS   int16 // TODO: Pipeline builder.
	VideoPipeline string

	// audio
	AudioDevice   string
	AudioCodec    codec.RTPCodec
	AudioBitrate  uint // TODO: Pipeline builder.
	AudioPipeline string

	// broadcast
	BroadcastPipeline string
	BroadcastUrl      string
}

func (Capture) Init(cmd *cobra.Command) error {
	//
	// video
	//

	cmd.PersistentFlags().String("display", ":99.0", "XDisplay to capture")
	if err := viper.BindPFlag("display", cmd.PersistentFlags().Lookup("display")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("video_codec", "vp8", "video codec to be used")
	if err := viper.BindPFlag("video_codec", cmd.PersistentFlags().Lookup("video_codec")); err != nil {
		return err
	}

	// DEPRECATED: video codec
	cmd.PersistentFlags().Bool("vp8", false, "DEPRECATED: use video_codec")
	if err := viper.BindPFlag("vp8", cmd.PersistentFlags().Lookup("vp8")); err != nil {
		return err
	}

	// DEPRECATED: video codec
	cmd.PersistentFlags().Bool("vp9", false, "DEPRECATED: use video_codec")
	if err := viper.BindPFlag("vp9", cmd.PersistentFlags().Lookup("vp9")); err != nil {
		return err
	}

	// DEPRECATED: video codec
	cmd.PersistentFlags().Bool("av1", false, "DEPRECATED: use video_codec")
	if err := viper.BindPFlag("av1", cmd.PersistentFlags().Lookup("av1")); err != nil {
		return err
	}

	// DEPRECATED: video codec
	cmd.PersistentFlags().Bool("h264", false, "DEPRECATED: use video_codec")
	if err := viper.BindPFlag("h264", cmd.PersistentFlags().Lookup("h264")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("hwenc", "", "use hardware accelerated encoding")
	if err := viper.BindPFlag("hwenc", cmd.PersistentFlags().Lookup("hwenc")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("video_bitrate", 3072, "video bitrate in kbit/s")
	if err := viper.BindPFlag("video_bitrate", cmd.PersistentFlags().Lookup("video_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("max_fps", 25, "maximum fps delivered via WebRTC, 0 is for no maximum")
	if err := viper.BindPFlag("max_fps", cmd.PersistentFlags().Lookup("max_fps")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("video", "", "video codec parameters to use for streaming")
	if err := viper.BindPFlag("video", cmd.PersistentFlags().Lookup("video")); err != nil {
		return err
	}

	//
	// audio
	//

	cmd.PersistentFlags().String("device", "audio_output.monitor", "audio device to capture")
	if err := viper.BindPFlag("device", cmd.PersistentFlags().Lookup("device")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("audio_codec", "opus", "audio codec to be used")
	if err := viper.BindPFlag("audio_codec", cmd.PersistentFlags().Lookup("audio_codec")); err != nil {
		return err
	}

	// DEPRECATED: audio codec
	cmd.PersistentFlags().Bool("opus", false, "DEPRECATED: use audio_codec")
	if err := viper.BindPFlag("opus", cmd.PersistentFlags().Lookup("opus")); err != nil {
		return err
	}

	// DEPRECATED: audio codec
	cmd.PersistentFlags().Bool("g722", false, "DEPRECATED: use audio_codec")
	if err := viper.BindPFlag("g722", cmd.PersistentFlags().Lookup("g722")); err != nil {
		return err
	}

	// DEPRECATED: audio codec
	cmd.PersistentFlags().Bool("pcmu", false, "DEPRECATED: use audio_codec")
	if err := viper.BindPFlag("pcmu", cmd.PersistentFlags().Lookup("pcmu")); err != nil {
		return err
	}

	// DEPRECATED: audio codec
	cmd.PersistentFlags().Bool("pcma", false, "DEPRECATED: use audio_codec")
	if err := viper.BindPFlag("pcma", cmd.PersistentFlags().Lookup("pcma")); err != nil {
		return err
	}
	// audio codecs

	cmd.PersistentFlags().Int("audio_bitrate", 128, "audio bitrate in kbit/s")
	if err := viper.BindPFlag("audio_bitrate", cmd.PersistentFlags().Lookup("audio_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("audio", "", "audio codec parameters to use for streaming")
	if err := viper.BindPFlag("audio", cmd.PersistentFlags().Lookup("audio")); err != nil {
		return err
	}

	//
	// broadcast
	//

	cmd.PersistentFlags().String("broadcast_pipeline", "", "custom gst pipeline used for broadcasting, strings {url} {device} {display} will be replaced")
	if err := viper.BindPFlag("broadcast_pipeline", cmd.PersistentFlags().Lookup("broadcast_pipeline")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("broadcast_url", "", "URL for broadcasting, setting this value will automatically enable broadcasting")
	if err := viper.BindPFlag("broadcast_url", cmd.PersistentFlags().Lookup("broadcast_url")); err != nil {
		return err
	}

	return nil
}

func (s *Capture) Set() {
	var ok bool

	//
	// video
	//

	s.Display = viper.GetString("display")

	videoCodec := viper.GetString("video_codec")
	s.VideoCodec, ok = codec.ParseStr(videoCodec)
	if !ok || s.VideoCodec.Type != webrtc.RTPCodecTypeVideo {
		log.Warn().Str("codec", videoCodec).Msgf("unknown video codec, using Vp8")
		s.VideoCodec = codec.VP8()
	}

	if viper.GetBool("vp8") {
		s.VideoCodec = codec.VP8()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_VP8=true', use 'NEKO_VIDEO_CODEC=vp8' instead")
	} else if viper.GetBool("vp9") {
		s.VideoCodec = codec.VP9()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_VP9=true', use 'NEKO_VIDEO_CODEC=vp9' instead")
	} else if viper.GetBool("h264") {
		s.VideoCodec = codec.H264()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_H264=true', use 'NEKO_VIDEO_CODEC=h264' instead")
	} else if viper.GetBool("av1") {
		s.VideoCodec = codec.AV1()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_AV1=true', use 'NEKO_VIDEO_CODEC=av1' instead")
	}

	videoHWEnc := strings.ToLower(viper.GetString("hwenc"))
	switch videoHWEnc {
	case "":
		fallthrough
	case "none":
		s.VideoHWEnc = HwEncNone
	case "vaapi":
		s.VideoHWEnc = HwEncVAAPI
	case "nvenc":
		s.VideoHWEnc = HwEncNVENC
	default:
		log.Warn().Str("hwenc", videoHWEnc).Msgf("unknown video hw encoder, using CPU")
	}

	s.VideoBitrate = viper.GetUint("video_bitrate")
	s.VideoMaxFPS = int16(viper.GetInt("max_fps"))
	s.VideoPipeline = viper.GetString("video")

	//
	// audio
	//

	s.AudioDevice = viper.GetString("device")

	audioCodec := viper.GetString("audio_codec")
	s.AudioCodec, ok = codec.ParseStr(audioCodec)
	if !ok || s.AudioCodec.Type != webrtc.RTPCodecTypeAudio {
		log.Warn().Str("codec", audioCodec).Msgf("unknown audio codec, using Opus")
		s.AudioCodec = codec.Opus()
	}

	if viper.GetBool("opus") {
		s.AudioCodec = codec.Opus()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_OPUS=true', use 'NEKO_VIDEO_CODEC=opus' instead")
	} else if viper.GetBool("g722") {
		s.AudioCodec = codec.G722()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_G722=true', use 'NEKO_VIDEO_CODEC=g722' instead")
	} else if viper.GetBool("pcmu") {
		s.AudioCodec = codec.PCMU()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_PCMU=true', use 'NEKO_VIDEO_CODEC=pcmu' instead")
	} else if viper.GetBool("pcma") {
		s.AudioCodec = codec.PCMA()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_PCMA=true', use 'NEKO_VIDEO_CODEC=pcma' instead")
	}

	s.AudioBitrate = viper.GetUint("audio_bitrate")
	s.AudioPipeline = viper.GetString("audio")

	//
	// broadcast
	//

	s.BroadcastPipeline = viper.GetString("broadcast_pipeline")
	s.BroadcastUrl = viper.GetString("broadcast_url")
}
