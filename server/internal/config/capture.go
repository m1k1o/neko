package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/internal/quality"
	"github.com/m1k1o/neko/server/pkg/gst"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type Capture struct {
	Display string

	VideoCodec     codec.RTPCodec
	VideoProfile   quality.Name
	VideoEncoder   quality.Encoder
	VideoAdaptive  bool
	VideoIDs       []string
	VideoPipelines map[string]types.VideoConfig
	// VideoPipelineFallbacks contains same-codec candidates for profile
	// pipelines that may fail when a hardware device is initialized at runtime.
	VideoPipelineFallbacks map[string][]types.VideoConfig
	VideoShowPointer       bool

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

	cmd.PersistentFlags().String("capture.video.profile", "", "optional Chromium M1 quality profile (low, balanced, high)")
	if err := viper.BindPFlag("capture.video.profile", cmd.PersistentFlags().Lookup("capture.video.profile")); err != nil {
		return err
	}
	cmd.PersistentFlags().String("capture.video.encoder", "auto", "Chromium M1 profile encoder (auto, software, vaapi, nvenc)")
	if err := viper.BindPFlag("capture.video.encoder", cmd.PersistentFlags().Lookup("capture.video.encoder")); err != nil {
		return err
	}
	cmd.PersistentFlags().Bool("capture.video.adaptive", false, "build a Chromium M1 quality ladder for bandwidth adaptation")
	if err := viper.BindPFlag("capture.video.adaptive", cmd.PersistentFlags().Lookup("capture.video.adaptive")); err != nil {
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

	cmd.PersistentFlags().Bool("capture.video.show_pointer", true, "show mouse pointer in captured video, overrides show_pointer of all video pipelines")
	if err := viper.BindPFlag("capture.video.show_pointer", cmd.PersistentFlags().Lookup("capture.video.show_pointer")); err != nil {
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

		} else if strings.TrimSpace(viper.GetString("capture.video.profile")) == "" {
			log.Warn().Msgf("no video pipelines specified, using default")

			s.VideoCodec = codec.VP8()
			s.VideoPipelines = map[string]types.VideoConfig{
				"main": {
					Fps:         "25",
					GstEncoder:  "vp8enc",
					ShowPointer: true,
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

		}
	} else if videoPipeline != "" {
		log.Warn().Msg("you are setting both single video pipeline and multiple video pipelines, ignoring single video pipeline")
	}

	s.VideoShowPointer = viper.GetBool("capture.video.show_pointer")
	s.VideoAdaptive = viper.GetBool("capture.video.adaptive")
	if viper.IsSet("capture.video.show_pointer") {
		for k, p := range s.VideoPipelines {
			p.ShowPointer = s.VideoShowPointer
			s.VideoPipelines[k] = p
		}
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

// ApplyVideoProfile applies an explicitly selected Chromium M1 quality
// profile after explicit configuration has been processed. Existing
// defaults and custom pipelines remain untouched when no profile is selected.
func (s *Capture) ApplyVideoProfile() error {
	return s.applyVideoProfile(gst.CheckElement)
}

func (s *Capture) applyVideoProfile(probe quality.ElementProbe) error {
	value := strings.ToLower(strings.TrimSpace(viper.GetString("capture.video.profile")))
	if value == "" {
		if viper.IsSet("capture.video.encoder") {
			return fmt.Errorf("capture.video.encoder requires capture.video.profile")
		}
		return nil
	}

	conflictingKeys := []string{
		"capture.video.ids",
		"capture.video.pipeline",
		"capture.video.pipelines",
		"video",
		"video_bitrate",
		"max_fps",
		"hwenc",
		"video_codec",
		"vp8",
		"vp9",
		"h264",
		"av1",
	}
	for _, key := range conflictingKeys {
		if viper.IsSet(key) {
			return fmt.Errorf("capture.video.profile cannot be combined with %s", key)
		}
	}

	profile, err := quality.Parse(value)
	if err != nil {
		return err
	}
	requestedEncoder, err := quality.ParseEncoder(viper.GetString("capture.video.encoder"))
	if err != nil {
		return err
	}
	selection, err := quality.ResolveEncoder(s.VideoCodec, requestedEncoder, probe)
	if err != nil {
		return err
	}
	s.VideoProfile = profile.Name
	s.VideoEncoder = selection.Encoder
	s.VideoCodec = selection.Codec
	s.VideoAdaptive = viper.GetBool("capture.video.adaptive")
	s.VideoIDs = []string{"main"}
	s.VideoPipelineFallbacks = make(map[string][]types.VideoConfig)
	profiles := []quality.Profile{profile}
	ids := []string{"main"}
	if s.VideoAdaptive {
		profiles = profiles[:0]
		ids = ids[:0]
		for _, name := range []quality.Name{quality.Low, quality.Balanced, quality.High} {
			tier, parseErr := quality.Parse(string(name))
			if parseErr != nil {
				return parseErr
			}
			profiles = append(profiles, tier)
			ids = append(ids, string(name))
			if name == profile.Name {
				break
			}
		}
	}
	s.VideoIDs = ids
	s.VideoPipelines = make(map[string]types.VideoConfig, len(profiles))
	s.VideoPipelineFallbacks = make(map[string][]types.VideoConfig, len(profiles))
	for index, tier := range profiles {
		videoConfig, configErr := tier.VideoConfig(selection.Codec, selection.Encoder, selection.Element, s.VideoShowPointer)
		if configErr != nil {
			return configErr
		}
		id := ids[index]
		s.VideoPipelines[id] = videoConfig
		if selection.Codec.Name == codec.H264().Name && selection.Element != "x264enc" && probe("x264enc") == nil && probe("h264parse") == nil {
			softwareConfig, fallbackErr := tier.VideoConfig(codec.H264(), quality.EncoderSoftware, "x264enc", s.VideoShowPointer)
			if fallbackErr == nil {
				s.VideoPipelineFallbacks[id] = []types.VideoConfig{softwareConfig}
			}
		}
	}
	if len(selection.Unavailable) > 0 {
		log.Warn().
			Str("requested", string(selection.Requested)).
			Str("selected", string(selection.Encoder)).
			Str("codec", selection.Codec.Name).
			Strs("unavailable_elements", selection.Unavailable).
			Msg("video encoder capability fallback applied")
	}
	log.Info().
		Str("profile", string(profile.Name)).
		Str("codec", selection.Codec.Name).
		Str("encoder", string(selection.Encoder)).
		Str("element", selection.Element).
		Bool("adaptive", s.VideoAdaptive).
		Int("width", profile.Width).
		Int("height", profile.Height).
		Int("fps", profile.FPS).
		Int("bitrate_kbps", profile.BitrateKbps).
		Msg("using explicit video quality profile")
	return nil
}
