package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"demodesk/neko/internal/types/codec"
)

type Capture struct {
	Device             string
	AudioCodec         codec.RTPCodec
	AudioPipeline      string

	Display            string
	//VideoCodec         codec.RTPCodec
	//VideoPipeline      string

	BroadcastPipeline  string

	Screencast         bool
	ScreencastRate     string
	ScreencastQuality  string
	ScreencastPipeline string
}

func (Capture) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("display", ":99.0", "XDisplay to capture")
	if err := viper.BindPFlag("display", cmd.PersistentFlags().Lookup("display")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("device", "auto_null.monitor", "audio device to capture")
	if err := viper.BindPFlag("device", cmd.PersistentFlags().Lookup("device")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("audio", "", "audio codec parameters to use for streaming")
	if err := viper.BindPFlag("audio", cmd.PersistentFlags().Lookup("audio")); err != nil {
		return err
	}

	//cmd.PersistentFlags().String("video", "", "video codec parameters to use for streaming")
	//if err := viper.BindPFlag("video", cmd.PersistentFlags().Lookup("video")); err != nil {
	//	return err
	//}

	// video codecs
	//cmd.PersistentFlags().Bool("vp8", false, "use VP8 video codec")
	//if err := viper.BindPFlag("vp8", cmd.PersistentFlags().Lookup("vp8")); err != nil {
	//	return err
	//}
	//
	//cmd.PersistentFlags().Bool("vp9", false, "use VP9 video codec")
	//if err := viper.BindPFlag("vp9", cmd.PersistentFlags().Lookup("vp9")); err != nil {
	//	return err
	//}
	//
	//cmd.PersistentFlags().Bool("h264", false, "use H264 video codec")
	//if err := viper.BindPFlag("h264", cmd.PersistentFlags().Lookup("h264")); err != nil {
	//	return err
	//}

	// audio codecs
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
	cmd.PersistentFlags().String("broadcast_pipeline", "", "audio video codec parameters to use for broadcasting")
	if err := viper.BindPFlag("broadcast_pipeline", cmd.PersistentFlags().Lookup("broadcast_pipeline")); err != nil {
		return err
	}

	// screencast
	cmd.PersistentFlags().Bool("screencast", false, "enable screencast")
	if err := viper.BindPFlag("screencast", cmd.PersistentFlags().Lookup("screencast")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("screencast_rate", "10/1", "set screencast frame rate")
	if err := viper.BindPFlag("screencast_rate", cmd.PersistentFlags().Lookup("screencast_rate")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("screencast_quality", "60", "set screencast JPEG quality")
	if err := viper.BindPFlag("screencast_quality", cmd.PersistentFlags().Lookup("screencast_quality")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("screencast_pipeline", "", "custom screencast pipeline")
	if err := viper.BindPFlag("screencast_pipeline", cmd.PersistentFlags().Lookup("screencast_pipeline")); err != nil {
		return err
	}

	return nil
}

func (s *Capture) Set() {
	//var videoCodec codec.RTPCodec
	//if viper.GetBool("vp8") {
	//	videoCodec = codec.VP8()
	//} else if viper.GetBool("vp9") {
	//	videoCodec = codec.VP9()
	//} else if viper.GetBool("h264") {
	//	videoCodec = codec.H264()
	//} else {
	//	// default
	//	videoCodec = codec.VP8()
	//}

	var audioCodec codec.RTPCodec
	if viper.GetBool("opus") {
		audioCodec = codec.Opus()
	} else if viper.GetBool("g722") {
		audioCodec = codec.G722()
	} else if viper.GetBool("pcmu") {
		audioCodec = codec.PCMU()
	} else if viper.GetBool("pcma") {
		audioCodec = codec.PCMA()
	} else {
		// default
		audioCodec = codec.Opus()
	}

	s.Device = viper.GetString("device")
	s.AudioCodec = audioCodec
	s.AudioPipeline = viper.GetString("audio")

	s.Display = viper.GetString("display")
	//s.VideoCodec = videoCodec
	//s.VideoPipeline = viper.GetString("video")

	s.BroadcastPipeline = viper.GetString("broadcast_pipeline")

	s.Screencast = viper.GetBool("screencast")
	s.ScreencastRate = viper.GetString("screencast_rate")
	s.ScreencastQuality = viper.GetString("screencast_quality")
	s.ScreencastPipeline = viper.GetString("screencast_pipeline")
}
