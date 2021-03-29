package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/codec"
	"demodesk/neko/internal/utils"
)

type Capture struct {
	Display string

	VideoCodec     codec.RTPCodec
	VideoIDs       []string
	VideoPipelines map[string]types.VideoConfig

	AudioDevice   string
	AudioCodec    codec.RTPCodec
	AudioPipeline string

	BroadcastAudioBitrate int
	BroadcastVideoBitrate int
	BroadcastPreset       string
	BroadcastPipeline     string

	ScreencastEnabled  bool
	ScreencastRate     string
	ScreencastQuality  string
	ScreencastPipeline string
}

func (Capture) Init(cmd *cobra.Command) error {
	// audio
	cmd.PersistentFlags().String("capture.audio.device", "auto_null.monitor", "audio device to capture")
	if err := viper.BindPFlag("capture.audio.device", cmd.PersistentFlags().Lookup("capture.audio.device")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.audio.codec", "opus", "audio codec to be used")
	if err := viper.BindPFlag("capture.audio.codec", cmd.PersistentFlags().Lookup("capture.audio.codec")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.audio.pipeline", "", "gstreamer pipeline used for audio streaming")
	if err := viper.BindPFlag("capture.audio.pipeline", cmd.PersistentFlags().Lookup("capture.audio.pipeline")); err != nil {
		return err
	}

	// videos
	cmd.PersistentFlags().String("capture.video.codec", "vp8", "video codec to be used")
	if err := viper.BindPFlag("capture.video.codec", cmd.PersistentFlags().Lookup("capture.video.codec")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("capture.video.ids", []string{}, "ordered list of video ids")
	if err := viper.BindPFlag("capture.video.ids", cmd.PersistentFlags().Lookup("capture.video.ids")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.video.pipelines", "", "pipelines config in JSON used for video streaming")
	if err := viper.BindPFlag("capture.video.pipelines", cmd.PersistentFlags().Lookup("capture.video.pipelines")); err != nil {
		return err
	}

	// broadcast
	cmd.PersistentFlags().Int("capture.screencast.audio_bitrate", 128, "broadcast audio bitrate in KB/s")
	if err := viper.BindPFlag("capture.screencast.audio_bitrate", cmd.PersistentFlags().Lookup("capture.screencast.audio_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("capture.screencast.video_bitrate", 4096, "broadcast video bitrate in KB/s")
	if err := viper.BindPFlag("capture.screencast.video_bitrate", cmd.PersistentFlags().Lookup("capture.screencast.video_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.screencast.preset", "veryfast", "broadcast speed preset for h264 encoding")
	if err := viper.BindPFlag("capture.screencast.preset", cmd.PersistentFlags().Lookup("capture.screencast.preset")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.broadcast.pipeline", "", "gstreamer pipeline used for broadcasting")
	if err := viper.BindPFlag("capture.broadcast.pipeline", cmd.PersistentFlags().Lookup("capture.broadcast.pipeline")); err != nil {
		return err
	}

	// screencast
	cmd.PersistentFlags().Bool("capture.screencast.enabled", false, "enable screencast")
	if err := viper.BindPFlag("capture.screencast.enabled", cmd.PersistentFlags().Lookup("capture.screencast.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.screencast.rate", "10/1", "screencast frame rate")
	if err := viper.BindPFlag("capture.screencast.rate", cmd.PersistentFlags().Lookup("capture.screencast.rate")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.screencast.quality", "60", "screencast JPEG quality")
	if err := viper.BindPFlag("capture.screencast.quality", cmd.PersistentFlags().Lookup("capture.screencast.quality")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.screencast.pipeline", "", "gstreamer pipeline used for screencasting")
	if err := viper.BindPFlag("capture.screencast.pipeline", cmd.PersistentFlags().Lookup("capture.screencast.pipeline")); err != nil {
		return err
	}

	return nil
}

func (s *Capture) Set() {
	// Display is provided by env variable
	s.Display = os.Getenv("DISPLAY")

	// video
	videoCodec := viper.GetString("capture.video.codec")
	switch videoCodec {
	case "vp8":
		s.VideoCodec = codec.VP8()
	case "vp9":
		s.VideoCodec = codec.VP9()
	case "h264":
		s.VideoCodec = codec.H264()
	default:
		log.Warn().Str("codec", videoCodec).Msgf("unknown video codec, using Vp8")
		s.VideoCodec = codec.VP8()
	}

	s.VideoIDs = viper.GetStringSlice("capture.video.ids")
	if err := viper.UnmarshalKey("capture.video.pipelines", &s.VideoPipelines, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.VideoPipelines),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse video pipelines")
	}

	// default video
	if len(s.VideoPipelines) == 0 {
		log.Warn().Msgf("no video pipelines specified, using defaults")
	
		s.VideoCodec = codec.VP8()
		s.VideoPipelines = map[string]types.VideoConfig{
			"main": types.VideoConfig{
				GstPipeline: "ximagesrc display-name={display} show-pointer=false use-damage=false "+
					"! video/x-raw "+
					"! videoconvert "+
					"! queue "+
					"! vp8enc end-usage=cbr cpu-used=4 threads=4 deadline=1 keyframe-max-dist=25 "+
					"! appsink name=appsink",
			},
		}
		s.VideoIDs = []string{"main"}
	}

	// audio
	s.AudioDevice = viper.GetString("capture.audio.device")
	s.AudioPipeline = viper.GetString("capture.audio.pipeline")

	audioCodec := viper.GetString("capture.audio.codec")
	switch audioCodec {
	case "opus":
		s.AudioCodec = codec.Opus()
	case "g722":
		s.AudioCodec = codec.G722()
	case "pcmu":
		s.AudioCodec = codec.PCMU()
	case "pcma":
		s.AudioCodec = codec.PCMA()
	default:
		log.Warn().Str("codec", audioCodec).Msgf("unknown audio codec, using Opus")
		s.AudioCodec = codec.Opus()
	}

	// broadcast
	s.BroadcastAudioBitrate = viper.GetInt("capture.broadcast.audio_bitrate")
	s.BroadcastVideoBitrate = viper.GetInt("capture.broadcast.video_bitrate")
	s.BroadcastPreset = viper.GetString("capture.broadcast.preset")
	s.BroadcastPipeline = viper.GetString("capture.broadcast.pipeline")

	// screencast
	s.ScreencastEnabled = viper.GetBool("capture.screencast.enabled")
	s.ScreencastRate = viper.GetString("capture.screencast.rate")
	s.ScreencastQuality = viper.GetString("capture.screencast.quality")
	s.ScreencastPipeline = viper.GetString("capture.screencast.pipeline")
}
