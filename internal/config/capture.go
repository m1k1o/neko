package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"demodesk/neko/internal/types/codec"
)

type Capture struct {
	Display string

	AudioDevice   string
	AudioCodec    codec.RTPCodec
	AudioPipeline string

	BroadcastPipeline string

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

	// broadcast
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

	s.BroadcastPipeline = viper.GetString("capture.broadcast.pipeline")

	s.ScreencastEnabled = viper.GetBool("capture.screencast.enabled")
	s.ScreencastRate = viper.GetString("capture.screencast.rate")
	s.ScreencastQuality = viper.GetString("capture.screencast.quality")
	s.ScreencastPipeline = viper.GetString("capture.screencast.pipeline")
}
