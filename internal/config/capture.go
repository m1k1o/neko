package config

import (
	"os"

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

	Screencast         bool
	ScreencastRate     string
	ScreencastQuality  string
	ScreencastPipeline string
}

func (Capture) Init(cmd *cobra.Command) error {
	// audio
	cmd.PersistentFlags().String("audio_device", "auto_null.monitor", "audio device to capture")
	if err := viper.BindPFlag("audio_device", cmd.PersistentFlags().Lookup("audio_device")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("audio_pipeline", "", "gstreamer pipeline used for audio streaming")
	if err := viper.BindPFlag("audio_pipeline", cmd.PersistentFlags().Lookup("audio_pipeline")); err != nil {
		return err
	}

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

	// broadcast
	cmd.PersistentFlags().String("broadcast_pipeline", "", "gstreamer pipeline used for broadcasting")
	if err := viper.BindPFlag("broadcast_pipeline", cmd.PersistentFlags().Lookup("broadcast_pipeline")); err != nil {
		return err
	}

	// screencast
	cmd.PersistentFlags().Bool("screencast", false, "enable screencast")
	if err := viper.BindPFlag("screencast", cmd.PersistentFlags().Lookup("screencast")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("screencast_rate", "10/1", "screencast frame rate")
	if err := viper.BindPFlag("screencast_rate", cmd.PersistentFlags().Lookup("screencast_rate")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("screencast_quality", "60", "screencast JPEG quality")
	if err := viper.BindPFlag("screencast_quality", cmd.PersistentFlags().Lookup("screencast_quality")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("screencast_pipeline", "", "gstreamer pipeline used for screencasting")
	if err := viper.BindPFlag("screencast_pipeline", cmd.PersistentFlags().Lookup("screencast_pipeline")); err != nil {
		return err
	}

	return nil
}

func (s *Capture) Set() {
	// Display is provided by env variable
	s.Display = os.Getenv("DISPLAY")

	s.AudioDevice = viper.GetString("audio_device")
	s.AudioPipeline = viper.GetString("audio_pipeline")

	if viper.GetBool("opus") {
		s.AudioCodec = codec.Opus()
	} else if viper.GetBool("g722") {
		s.AudioCodec = codec.G722()
	} else if viper.GetBool("pcmu") {
		s.AudioCodec = codec.PCMU()
	} else if viper.GetBool("pcma") {
		s.AudioCodec = codec.PCMA()
	} else {
		// default
		s.AudioCodec = codec.Opus()
	}

	s.BroadcastPipeline = viper.GetString("broadcast_pipeline")

	s.Screencast = viper.GetBool("screencast")
	s.ScreencastRate = viper.GetString("screencast_rate")
	s.ScreencastQuality = viper.GetString("screencast_quality")
	s.ScreencastPipeline = viper.GetString("screencast_pipeline")
}
