package config

import (
	"m1k1o/neko/internal/types/codec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Capture struct {
	// video
	Display       string
	VideoCodec    codec.RTPCodec
	VideoHWEnc    string // TODO: Pipeline builder.
	VideoBitrate  uint   // TODO: Pipeline builder.
	VideoMaxFPS   int16  // TODO: Pipeline builder.
	VideoPipeline string

	// audio
	AudioDevice   string
	AudioCodec    codec.RTPCodec
	AudioBitrate  uint // TODO: Pipeline builder.
	AudioPipeline string

	// broadcast
	BroadcastPipeline string
	BroadcastUrl      string
	BroadcastStarted  bool
}

func (Capture) Init(cmd *cobra.Command) error {
	//
	// video
	//

	cmd.PersistentFlags().String("display", ":99.0", "XDisplay to capture")
	if err := viper.BindPFlag("display", cmd.PersistentFlags().Lookup("display")); err != nil {
		return err
	}

	// video codecs
	// TODO: video.codec
	cmd.PersistentFlags().Bool("vp8", false, "use VP8 video codec")
	if err := viper.BindPFlag("vp8", cmd.PersistentFlags().Lookup("vp8")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("vp9", false, "use VP9 video codec")
	if err := viper.BindPFlag("vp9", cmd.PersistentFlags().Lookup("vp9")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("h264", false, "use H264 video codec")
	if err := viper.BindPFlag("h264", cmd.PersistentFlags().Lookup("h264")); err != nil {
		return err
	}
	// video codecs

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

	cmd.PersistentFlags().String("device", "auto_null.monitor", "audio device to capture")
	if err := viper.BindPFlag("device", cmd.PersistentFlags().Lookup("device")); err != nil {
		return err
	}

	// audio codecs
	// TODO: audio.codec
	cmd.PersistentFlags().Bool("opus", false, "use Opus audio codec")
	if err := viper.BindPFlag("opus", cmd.PersistentFlags().Lookup("opus")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("g722", false, "use G722 audio codec")
	if err := viper.BindPFlag("g722", cmd.PersistentFlags().Lookup("g722")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("pcmu", false, "use PCMU audio codec")
	if err := viper.BindPFlag("pcmu", cmd.PersistentFlags().Lookup("pcmu")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("pcma", false, "use PCMA audio codec")
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
	//
	// video
	//

	s.Display = viper.GetString("display")

	videoCodec := codec.VP8()
	if viper.GetBool("vp8") {
		videoCodec = codec.VP8()
	} else if viper.GetBool("vp9") {
		videoCodec = codec.VP9()
	} else if viper.GetBool("h264") {
		videoCodec = codec.H264()
	}
	s.VideoCodec = videoCodec

	videoHWEnc := ""
	if viper.GetString("hwenc") == "VAAPI" {
		videoHWEnc = "VAAPI"
	}
	s.VideoHWEnc = videoHWEnc

	s.VideoBitrate = viper.GetUint("video_bitrate")
	s.VideoMaxFPS = int16(viper.GetInt("max_fps"))
	s.VideoPipeline = viper.GetString("video")

	//
	// audio
	//

	s.AudioDevice = viper.GetString("device")

	audioCodec := codec.Opus()
	if viper.GetBool("opus") {
		audioCodec = codec.Opus()
	} else if viper.GetBool("g722") {
		audioCodec = codec.G722()
	} else if viper.GetBool("pcmu") {
		audioCodec = codec.PCMU()
	} else if viper.GetBool("pcma") {
		audioCodec = codec.PCMA()
	}
	s.AudioCodec = audioCodec

	s.AudioBitrate = viper.GetUint("audio_bitrate")
	s.AudioPipeline = viper.GetString("audio")

	//
	// broadcast
	//

	s.BroadcastPipeline = viper.GetString("broadcast_pipeline")
	s.BroadcastUrl = viper.GetString("broadcast_url")
	s.BroadcastStarted = s.BroadcastUrl != ""
}
