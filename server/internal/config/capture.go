package config

import (
	"os"
	"strings"

	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
	"github.com/m1k1o/neko/server/pkg/utils"
)

// Legacy capture configuration
type HwEnc int

// Legacy capture configuration
const (
	HwEncUnset HwEnc = iota
	HwEncNone
	HwEncVAAPI
	HwEncNVENC
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
	BroadcastUrl          string
	BroadcastAutostart    bool

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
	cmd.PersistentFlags().String("capture.video.display", "", "X display to capture")
	if err := viper.BindPFlag("capture.video.display", cmd.PersistentFlags().Lookup("capture.video.display")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.video.codec", "vp8", "video codec to be used")
	if err := viper.BindPFlag("capture.video.codec", cmd.PersistentFlags().Lookup("capture.video.codec")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("capture.video.ids", []string{}, "ordered list of video ids")
	if err := viper.BindPFlag("capture.video.ids", cmd.PersistentFlags().Lookup("capture.video.ids")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.video.pipelines", "{}", "pipelines config used for video streaming")
	if err := viper.BindPFlag("capture.video.pipelines", cmd.PersistentFlags().Lookup("capture.video.pipelines")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("capture.video.pipeline", "", "shortcut for configuring only a single gstreamer pipeline, ignored if pipelines is set")
	if err := viper.BindPFlag("capture.video.pipeline", cmd.PersistentFlags().Lookup("capture.video.pipeline")); err != nil {
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

	cmd.PersistentFlags().String("capture.broadcast.url", "", "initial URL for broadcasting, setting this value will automatically start broadcasting")
	if err := viper.BindPFlag("capture.broadcast.url", cmd.PersistentFlags().Lookup("capture.broadcast.url")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("capture.broadcast.autostart", true, "automatically start broadcasting when neko starts and broadcast_url is set")
	if err := viper.BindPFlag("capture.broadcast.autostart", cmd.PersistentFlags().Lookup("capture.broadcast.autostart")); err != nil {
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

func (Capture) InitV2(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("display", "", "V2: XDisplay to capture")
	if err := viper.BindPFlag("display", cmd.PersistentFlags().Lookup("display")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("video_codec", "", "V2: video codec to be used")
	if err := viper.BindPFlag("video_codec", cmd.PersistentFlags().Lookup("video_codec")); err != nil {
		return err
	}

	// DEPRECATED: video codec
	cmd.PersistentFlags().Bool("vp8", false, "V2 DEPRECATED: use video_codec")
	if err := viper.BindPFlag("vp8", cmd.PersistentFlags().Lookup("vp8")); err != nil {
		return err
	}

	// DEPRECATED: video codec
	cmd.PersistentFlags().Bool("vp9", false, "V2 DEPRECATED: use video_codec")
	if err := viper.BindPFlag("vp9", cmd.PersistentFlags().Lookup("vp9")); err != nil {
		return err
	}

	// DEPRECATED: video codec
	cmd.PersistentFlags().Bool("av1", false, "V2 DEPRECATED: use video_codec")
	if err := viper.BindPFlag("av1", cmd.PersistentFlags().Lookup("av1")); err != nil {
		return err
	}

	// DEPRECATED: video codec
	cmd.PersistentFlags().Bool("h264", false, "V2 DEPRECATED: use video_codec")
	if err := viper.BindPFlag("h264", cmd.PersistentFlags().Lookup("h264")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("hwenc", "", "V2: use hardware accelerated encoding")
	if err := viper.BindPFlag("hwenc", cmd.PersistentFlags().Lookup("hwenc")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("video_bitrate", 0, "V2: video bitrate in kbit/s")
	if err := viper.BindPFlag("video_bitrate", cmd.PersistentFlags().Lookup("video_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("max_fps", 0, "V2: maximum fps delivered via WebRTC, 0 is for no maximum")
	if err := viper.BindPFlag("max_fps", cmd.PersistentFlags().Lookup("max_fps")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("video", "", "V2: video codec parameters to use for streaming")
	if err := viper.BindPFlag("video", cmd.PersistentFlags().Lookup("video")); err != nil {
		return err
	}

	//
	// audio
	//

	cmd.PersistentFlags().String("device", "", "V2: audio device to capture")
	if err := viper.BindPFlag("device", cmd.PersistentFlags().Lookup("device")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("audio_codec", "", "V2: audio codec to be used")
	if err := viper.BindPFlag("audio_codec", cmd.PersistentFlags().Lookup("audio_codec")); err != nil {
		return err
	}

	// DEPRECATED: audio codec
	cmd.PersistentFlags().Bool("opus", false, "V2 DEPRECATED: use audio_codec")
	if err := viper.BindPFlag("opus", cmd.PersistentFlags().Lookup("opus")); err != nil {
		return err
	}

	// DEPRECATED: audio codec
	cmd.PersistentFlags().Bool("g722", false, "V2 DEPRECATED: use audio_codec")
	if err := viper.BindPFlag("g722", cmd.PersistentFlags().Lookup("g722")); err != nil {
		return err
	}

	// DEPRECATED: audio codec
	cmd.PersistentFlags().Bool("pcmu", false, "V2 DEPRECATED: use audio_codec")
	if err := viper.BindPFlag("pcmu", cmd.PersistentFlags().Lookup("pcmu")); err != nil {
		return err
	}

	// DEPRECATED: audio codec
	cmd.PersistentFlags().Bool("pcma", false, "V2 DEPRECATED: use audio_codec")
	if err := viper.BindPFlag("pcma", cmd.PersistentFlags().Lookup("pcma")); err != nil {
		return err
	}
	// audio codecs

	cmd.PersistentFlags().Int("audio_bitrate", 0, "V2: audio bitrate in kbit/s")
	if err := viper.BindPFlag("audio_bitrate", cmd.PersistentFlags().Lookup("audio_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("audio", "", "V2: audio codec parameters to use for streaming")
	if err := viper.BindPFlag("audio", cmd.PersistentFlags().Lookup("audio")); err != nil {
		return err
	}

	//
	// broadcast
	//

	cmd.PersistentFlags().String("broadcast_pipeline", "", "V2: custom gst pipeline used for broadcasting, strings {hostname} {url} {device} {display} will be replaced")
	if err := viper.BindPFlag("broadcast_pipeline", cmd.PersistentFlags().Lookup("broadcast_pipeline")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("broadcast_url", "", "V2: a default default URL for broadcast streams, can be disabled/changed later by admins in the GUI")
	if err := viper.BindPFlag("broadcast_url", cmd.PersistentFlags().Lookup("broadcast_url")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("broadcast_autostart", false, "V2: automatically start broadcasting when neko starts and broadcast_url is set")
	if err := viper.BindPFlag("broadcast_autostart", cmd.PersistentFlags().Lookup("broadcast_autostart")); err != nil {
		return err
	}

	return nil
}

func (s *Capture) Set() {
	var ok bool

	s.Display = viper.GetString("capture.video.display")

	// Display is provided by env variable unless explicitly set
	if s.Display == "" {
		s.Display = os.Getenv("DISPLAY")
	}

	// video
	videoCodec := viper.GetString("capture.video.codec")
	s.VideoCodec, ok = codec.ParseStr(videoCodec)
	if !ok || !s.VideoCodec.IsVideo() {
		log.Warn().Str("codec", videoCodec).Msgf("unknown video codec, using Vp8")
		s.VideoCodec = codec.VP8()
	}

	s.VideoIDs = viper.GetStringSlice("capture.video.ids")
	if err := viper.UnmarshalKey("capture.video.pipelines", &s.VideoPipelines, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.VideoPipelines),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse video pipelines")
	}

	videoPipeline := viper.GetString("capture.video.pipeline")

	// if no video pipelines are set
	if len(s.VideoPipelines) == 0 {
		// maybe single video pipeline is set
		if videoPipeline != "" {
			log.Info().Str("pipeline", videoPipeline).Msg("using single video pipeline")

			s.VideoPipelines = map[string]types.VideoConfig{
				"main": {
					GstPipeline: videoPipeline,
				},
			}
			s.VideoIDs = []string{"main"}

			if viper.GetBool("legacy") {
				legacyPipeline := s.VideoPipelines["main"]
				// Hacky way to enable pointer for legacy pipeline.
				legacyPipeline.GstPipeline = strings.Replace(legacyPipeline.GstPipeline, "show-pointer=false", "show-pointer=true", 1)
				s.VideoPipelines["legacy"] = legacyPipeline
				// we do not add legacy to VideoIDs so that its ignored by bandwidth estimator
			}
		} else {
			log.Warn().Msgf("no video pipelines specified, using default")

			s.VideoCodec = codec.VP8()
			s.VideoPipelines = map[string]types.VideoConfig{
				"main": {
					Fps:        "25",
					GstEncoder: "vp8enc",
					GstParams: map[string]string{
						"target-bitrate":      "round(3072 * 650)",
						"cpu-used":            "4",
						"end-usage":           "cbr",
						"threads":             "4",
						"deadline":            "1",
						"undershoot":          "95",
						"buffer-size":         "(3072 * 4)",
						"buffer-initial-size": "(3072 * 2)",
						"buffer-optimal-size": "(3072 * 3)",
						"keyframe-max-dist":   "25",
						"min-quantizer":       "4",
						"max-quantizer":       "20",
					},
				},
			}
			s.VideoIDs = []string{"main"}

			if viper.GetBool("legacy") {
				legacyPipeline := s.VideoPipelines["main"]
				legacyPipeline.ShowPointer = true
				s.VideoPipelines["legacy"] = legacyPipeline
				// we do not add legacy to VideoIDs so that its ignored by bandwidth estimator
			}
		}
	} else if videoPipeline != "" {
		log.Warn().Msg("you are setting both single video pipeline and multiple video pipelines, ignoring single video pipeline")
	}

	// audio
	s.AudioDevice = viper.GetString("capture.audio.device")
	s.AudioPipeline = viper.GetString("capture.audio.pipeline")

	audioCodec := viper.GetString("capture.audio.codec")
	s.AudioCodec, ok = codec.ParseStr(audioCodec)
	if !ok || !s.AudioCodec.IsAudio() {
		log.Warn().Str("codec", audioCodec).Msgf("unknown audio codec, using Opus")
		s.AudioCodec = codec.Opus()
	}

	// broadcast
	s.BroadcastAudioBitrate = viper.GetInt("capture.broadcast.audio_bitrate")
	s.BroadcastVideoBitrate = viper.GetInt("capture.broadcast.video_bitrate")
	s.BroadcastPreset = viper.GetString("capture.broadcast.preset")
	s.BroadcastPipeline = viper.GetString("capture.broadcast.pipeline")
	s.BroadcastUrl = viper.GetString("capture.broadcast.url")
	s.BroadcastAutostart = viper.GetBool("capture.broadcast.autostart")

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

func (s *Capture) SetV2() {
	enableLegacy := false

	var ok bool

	//
	// video
	//

	if display := viper.GetString("display"); display != "" {
		s.Display = display
		log.Warn().Msg("you are using v2 configuration 'NEKO_DISPLAY' which is deprecated, please use 'NEKO_CAPTURE_VIDEO_DISPLAY' and/or 'NEKO_DESKTOP_DISPLAY' instead, also consider using 'DISPLAY' env variable if both should be the same")
		enableLegacy = true
	}

	modifiedVideoCodec := false
	if videoCodec := viper.GetString("video_codec"); videoCodec != "" {
		s.VideoCodec, ok = codec.ParseStr(videoCodec)
		if !ok || s.VideoCodec.Type != webrtc.RTPCodecTypeVideo {
			log.Warn().Str("codec", videoCodec).Msgf("unknown video codec, using Vp8")
			s.VideoCodec = codec.VP8()
		}
		log.Warn().Msg("you are using v2 configuration 'NEKO_VIDEO_CODEC' which is deprecated, please use 'NEKO_CAPTURE_VIDEO_CODEC' instead")
		enableLegacy = true
		modifiedVideoCodec = true
	}

	if viper.GetBool("vp8") {
		s.VideoCodec = codec.VP8()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_VP8=true', use 'NEKO_CAPTURE_VIDEO_CODEC=vp8' instead")
		enableLegacy = true
		modifiedVideoCodec = true
	} else if viper.GetBool("vp9") {
		s.VideoCodec = codec.VP9()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_VP9=true', use 'NEKO_CAPTURE_VIDEO_CODEC=vp9' instead")
		enableLegacy = true
		modifiedVideoCodec = true
	} else if viper.GetBool("h264") {
		s.VideoCodec = codec.H264()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_H264=true', use 'NEKO_CAPTURE_VIDEO_CODEC=h264' instead")
		enableLegacy = true
		modifiedVideoCodec = true
	} else if viper.GetBool("av1") {
		s.VideoCodec = codec.AV1()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_AV1=true', use 'NEKO_CAPTURE_VIDEO_CODEC=av1' instead")
		enableLegacy = true
		modifiedVideoCodec = true
	}

	videoHWEnc := HwEncUnset
	if hwenc := strings.ToLower(viper.GetString("hwenc")); hwenc != "" {
		switch hwenc {
		case "none":
			videoHWEnc = HwEncNone
		case "vaapi":
			videoHWEnc = HwEncVAAPI
		case "nvenc":
			videoHWEnc = HwEncNVENC
		default:
			log.Warn().Str("hwenc", hwenc).Msgf("unknown video hw encoder, using CPU")
		}
	}

	videoBitrate := viper.GetUint("video_bitrate")
	videoMaxFPS := int16(viper.GetInt("max_fps"))
	videoPipeline := viper.GetString("video")

	// video pipeline
	if modifiedVideoCodec || videoHWEnc != HwEncUnset || videoBitrate != 0 || videoMaxFPS != 0 || videoPipeline != "" {
		pipeline, err := NewVideoPipeline(s.VideoCodec, s.Display, videoPipeline, videoMaxFPS, videoBitrate, videoHWEnc)
		if err != nil {
			log.Warn().Err(err).Msg("unable to create video pipeline, using default")
		} else {
			s.VideoPipelines = map[string]types.VideoConfig{
				"main": {
					// Hacky way to disable pointer.
					GstPipeline: strings.Replace(pipeline, "show-pointer=true", "show-pointer=false", 1),
				},
				"legacy": {
					GstPipeline: pipeline,
				},
			}
			// we do not add legacy to VideoIDs so that its ignored by bandwidth estimator
			s.VideoIDs = []string{"main"}
		}

		if videoPipeline != "" {
			log.Warn().Msg("you are using v2 configuration 'NEKO_VIDEO' which is deprecated, please use 'NEKO_CAPTURE_VIDEO_PIPELINE' instead")
		}

		// TODO: add deprecated warning and proper alternative for HW enc, bitrate and max fps
		enableLegacy = true
	}

	//
	// audio
	//

	if audioDevice := viper.GetString("device"); audioDevice != "" {
		s.AudioDevice = audioDevice
		log.Warn().Msg("you are using v2 configuration 'NEKO_DEVICE' which is deprecated, please use 'NEKO_CAPTURE_AUDIO_DEVICE' instead")
		enableLegacy = true
	}

	modifiedAudioCodec := false
	if audioCodec := viper.GetString("audio_codec"); audioCodec != "" {
		s.AudioCodec, ok = codec.ParseStr(audioCodec)
		if !ok || s.AudioCodec.Type != webrtc.RTPCodecTypeAudio {
			log.Warn().Str("codec", audioCodec).Msgf("unknown audio codec, using Opus")
			s.AudioCodec = codec.Opus()
		}
		log.Warn().Msg("you are using v2 configuration 'NEKO_AUDIO_CODEC' which is deprecated, please use 'NEKO_CAPTURE_AUDIO_CODEC' instead")
		enableLegacy = true
		modifiedAudioCodec = true
	}

	if viper.GetBool("opus") {
		s.AudioCodec = codec.Opus()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_OPUS=true', use 'NEKO_CAPTURE_AUDIO_CODEC=opus' instead")
		enableLegacy = true
		modifiedAudioCodec = true
	} else if viper.GetBool("g722") {
		s.AudioCodec = codec.G722()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_G722=true', use 'NEKO_CAPTURE_AUDIO_CODEC=g722' instead")
		enableLegacy = true
		modifiedAudioCodec = true
	} else if viper.GetBool("pcmu") {
		s.AudioCodec = codec.PCMU()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_PCMU=true', use 'NEKO_CAPTURE_AUDIO_CODEC=pcmu' instead")
		enableLegacy = true
		modifiedAudioCodec = true
	} else if viper.GetBool("pcma") {
		s.AudioCodec = codec.PCMA()
		log.Warn().Msg("you are using deprecated config setting 'NEKO_PCMA=true', use 'NEKO_CAPTURE_AUDIO_CODEC=pcma' instead")
		enableLegacy = true
		modifiedAudioCodec = true
	}

	audioBitrate := viper.GetUint("audio_bitrate")
	audioPipeline := viper.GetString("audio")

	// audio pipeline
	if modifiedAudioCodec || audioBitrate != 0 || audioPipeline != "" {
		pipeline, err := NewAudioPipeline(s.AudioCodec, s.AudioDevice, audioPipeline, audioBitrate)
		if err != nil {
			log.Warn().Err(err).Msg("unable to create audio pipeline, using default")
		} else {
			s.AudioPipeline = pipeline
		}

		if audioPipeline != "" {
			log.Warn().Msg("you are using v2 configuration 'NEKO_AUDIO' which is deprecated, please use 'NEKO_CAPTURE_AUDIO_PIPELINE' instead")
		}

		// TODO: add deprecated warning and proper alternative for audio bitrate
		enableLegacy = true
	}

	//
	// broadcast
	//

	if viper.IsSet("broadcast_pipeline") {
		s.BroadcastPipeline = viper.GetString("broadcast_pipeline")
		log.Warn().Msg("you are using v2 configuration 'NEKO_BROADCAST_PIPELINE' which is deprecated, please use 'NEKO_CAPTURE_BROADCAST_PIPELINE' instead")
		enableLegacy = true
	}
	if viper.IsSet("broadcast_url") {
		s.BroadcastUrl = viper.GetString("broadcast_url")
		log.Warn().Msg("you are using v2 configuration 'NEKO_BROADCAST_URL' which is deprecated, please use 'NEKO_CAPTURE_BROADCAST_URL' instead")
		enableLegacy = true
	}
	if viper.IsSet("broadcast_autostart") {
		s.BroadcastAutostart = viper.GetBool("broadcast_autostart")
		log.Warn().Msg("you are using v2 configuration 'NEKO_BROADCAST_AUTOSTART' which is deprecated, please use 'NEKO_CAPTURE_BROADCAST_AUTOSTART' instead")
		enableLegacy = true
	}

	// set legacy flag if any V2 configuration was used
	if !viper.IsSet("legacy") && enableLegacy {
		log.Warn().Msg("legacy configuration is enabled because at least one V2 configuration was used, please migrate to V3 configuration, visit https://neko.m1k1o.net/docs/v3/migration-from-v2 for more details")
		viper.Set("legacy", true)
	}
}
