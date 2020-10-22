package config

import (
	"regexp"
	"strconv"

	"github.com/pion/webrtc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Remote struct {
	Display      string
	Device       string
	AudioCodec   string
	AudioParams  string
	VideoCodec   string
	VideoParams  string
	ScreenWidth  int
	ScreenHeight int
	ScreenRate   int
}

func (Remote) Init(cmd *cobra.Command) error {
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

	cmd.PersistentFlags().String("video", "", "video codec parameters to use for streaming")
	if err := viper.BindPFlag("video", cmd.PersistentFlags().Lookup("video")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("screen", "1280x720@30", "default screen resolution and framerate")
	if err := viper.BindPFlag("screen", cmd.PersistentFlags().Lookup("screen")); err != nil {
		return err
	}

	// video codecs
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

	return nil
}

func (s *Remote) Set() {
	videoCodec := webrtc.VP8
	if viper.GetBool("vp8") {
		videoCodec = webrtc.VP8
	} else if viper.GetBool("vp9") {
		videoCodec = webrtc.VP9
	} else if viper.GetBool("h264") {
		videoCodec = webrtc.H264
	}

	audioCodec := webrtc.Opus
	if viper.GetBool("opus") {
		audioCodec = webrtc.Opus
	} else if viper.GetBool("g722") {
		audioCodec = webrtc.G722
	} else if viper.GetBool("pcmu") {
		audioCodec = webrtc.PCMU
	} else if viper.GetBool("pcma") {
		audioCodec = webrtc.PCMA
	}

	s.Device = viper.GetString("device")
	s.AudioCodec = audioCodec
	s.AudioParams = viper.GetString("audio")
	s.Display = viper.GetString("display")
	s.VideoCodec = videoCodec
	s.VideoParams = viper.GetString("video")

	s.ScreenWidth = 1280
	s.ScreenHeight = 720
	s.ScreenRate = 30

	r := regexp.MustCompile(`([0-9]{1,4})x([0-9]{1,4})@([0-9]{1,3})`)
	res := r.FindStringSubmatch(viper.GetString("screen"))

	if len(res) > 0 {
		width, err1 := strconv.ParseInt(res[1], 10, 64)
		height, err2 := strconv.ParseInt(res[2], 10, 64)
		rate, err3 := strconv.ParseInt(res[3], 10, 64)

		if err1 == nil && err2 == nil && err3 == nil {
			s.ScreenWidth = int(width)
			s.ScreenHeight = int(height)
			s.ScreenRate = int(rate)
		}
	}
}
