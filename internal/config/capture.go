package config

import (
	"os"

	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/codec"
	"github.com/demodesk/neko/pkg/utils"
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

	WebcamEnabled bool
	WebcamDevice  string
	WebcamWidth   int
	WebcamHeight  int

	MicrophoneEnabled bool
	MicrophoneDevice  string
}

func (Capture) Init(cmd *cobra.Command) error {
	// audio
	cmd.PersistentFlags().String("capture.audio.device", "audio_output.monitor", "pulseaudio device to capture")
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
	cmd.PersistentFlags().Int("capture.broadcast.audio_bitrate", 128, "broadcast audio bitrate in KB/s")
	if err := viper.BindPFlag("capture.broadcast.audio_bitrate", cmd.PersistentFlags().Lookup("capture.broadcast.audio_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("capture.broadcast.video_bitrate", 4096, "broadcast video bitrate in KB/s")
	if err := viper.BindPFlag("capture.broadcast.video_bitrate", cmd.PersistentFlags().Lookup("capture.broadcast.video_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.broadcast.preset", "veryfast", "broadcast speed preset for h264 encoding")
	if err := viper.BindPFlag("capture.broadcast.preset", cmd.PersistentFlags().Lookup("capture.broadcast.preset")); err != nil {
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

	// webcam
	cmd.PersistentFlags().Bool("capture.webcam.enabled", false, "enable webcam stream")
	if err := viper.BindPFlag("capture.webcam.enabled", cmd.PersistentFlags().Lookup("capture.webcam.enabled")); err != nil {
		return err
	}

	// sudo apt install v4l2loopback-dkms v4l2loopback-utils
	// sudo apt-get install linux-headers-`uname -r` linux-modules-extra-`uname -r`
	// sudo modprobe v4l2loopback exclusive_caps=1
	cmd.PersistentFlags().String("capture.webcam.device", "/dev/video0", "v4l2sink device used for webcam")
	if err := viper.BindPFlag("capture.webcam.device", cmd.PersistentFlags().Lookup("capture.webcam.device")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("capture.webcam.width", 1280, "webcam stream width")
	if err := viper.BindPFlag("capture.webcam.width", cmd.PersistentFlags().Lookup("capture.webcam.width")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("capture.webcam.height", 720, "webcam stream height")
	if err := viper.BindPFlag("capture.webcam.height", cmd.PersistentFlags().Lookup("capture.webcam.height")); err != nil {
		return err
	}

	// microphone
	cmd.PersistentFlags().Bool("capture.microphone.enabled", true, "enable microphone stream")
	if err := viper.BindPFlag("capture.microphone.enabled", cmd.PersistentFlags().Lookup("capture.microphone.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.microphone.device", "audio_input", "pulseaudio device used for microphone")
	if err := viper.BindPFlag("capture.microphone.device", cmd.PersistentFlags().Lookup("capture.microphone.device")); err != nil {
		return err
	}

	return nil
}

func (s *Capture) Set() {
	var ok bool

	// Display is provided by env variable
	s.Display = os.Getenv("DISPLAY")

	// video
	videoCodec := viper.GetString("capture.video.codec")
	s.VideoCodec, ok = codec.ParseStr(videoCodec)
	if !ok || s.VideoCodec.Type != webrtc.RTPCodecTypeVideo {
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
			"main": {
				GstPipeline: "ximagesrc display-name={display} show-pointer=false use-damage=false " +
					"! video/x-raw " +
					"! videoconvert " +
					"! queue " +
					"! vp8enc end-usage=cbr cpu-used=4 threads=4 deadline=1 keyframe-max-dist=25 " +
					"! appsink name=appsink",
			},
		}
		s.VideoIDs = []string{"main"}
	}

	// audio
	s.AudioDevice = viper.GetString("capture.audio.device")
	s.AudioPipeline = viper.GetString("capture.audio.pipeline")

	audioCodec := viper.GetString("capture.audio.codec")
	s.AudioCodec, ok = codec.ParseStr(audioCodec)
	if !ok || s.AudioCodec.Type != webrtc.RTPCodecTypeAudio {
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

	// webcam
	s.WebcamEnabled = viper.GetBool("capture.webcam.enabled")
	s.WebcamDevice = viper.GetString("capture.webcam.device")
	s.WebcamWidth = viper.GetInt("capture.webcam.width")
	s.WebcamHeight = viper.GetInt("capture.webcam.height")

	// microphone
	s.MicrophoneEnabled = viper.GetBool("capture.microphone.enabled")
	s.MicrophoneDevice = viper.GetString("capture.microphone.device")
}
