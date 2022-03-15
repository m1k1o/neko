package config

import (
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Remote struct {
	Display      string
	Device       string
	AudioCodec   string
	AudioParams  string
	AudioBitrate uint
	VideoHWEnc   string
	VideoCodec   string
	VideoParams  string
	VideoBitrate uint
	ScreenWidth  int
	ScreenHeight int
	ScreenRate   int
	MaxFPS       int
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

	cmd.PersistentFlags().Int("audio_bitrate", 128, "audio bitrate in kbit/s")
	if err := viper.BindPFlag("audio_bitrate", cmd.PersistentFlags().Lookup("audio_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("video", "", "video codec parameters to use for streaming")
	if err := viper.BindPFlag("video", cmd.PersistentFlags().Lookup("video")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("video_bitrate", 3072, "video bitrate in kbit/s")
	if err := viper.BindPFlag("video_bitrate", cmd.PersistentFlags().Lookup("video_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("screen", "1280x720@30", "default screen resolution and framerate")
	if err := viper.BindPFlag("screen", cmd.PersistentFlags().Lookup("screen")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("max_fps", 25, "maximum fps delivered via WebRTC, 0 is for no maximum")
	if err := viper.BindPFlag("max_fps", cmd.PersistentFlags().Lookup("max_fps")); err != nil {
		return err
	}

	// hw encoding
	cmd.PersistentFlags().String("hwenc", "", "use hardware accelerated encoding")
	if err := viper.BindPFlag("hwenc", cmd.PersistentFlags().Lookup("hwenc")); err != nil {
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
	audioCodec := "Opus"
	if viper.GetBool("opus") {
		audioCodec = "Opus"
	} else if viper.GetBool("g722") {
		audioCodec = "G722"
	} else if viper.GetBool("pcmu") {
		audioCodec = "PCMU"
	} else if viper.GetBool("pcma") {
		audioCodec = "PCMA"
	}

	s.Device = viper.GetString("device")
	s.AudioCodec = audioCodec
	s.AudioParams = viper.GetString("audio")
	s.AudioBitrate = viper.GetUint("audio_bitrate")

	videoCodec := "VP8"
	if viper.GetBool("vp8") {
		videoCodec = "VP8"
	} else if viper.GetBool("vp9") {
		videoCodec = "VP9"
	} else if viper.GetBool("h264") {
		videoCodec = "H264"
	}
	videoHWEnc := ""
	if viper.GetString("hwenc") == "VAAPI" {
		videoHWEnc = "VAAPI"
	}
	s.Display = viper.GetString("display")
	s.VideoHWEnc = videoHWEnc
	s.VideoCodec = videoCodec
	s.VideoParams = viper.GetString("video")
	s.VideoBitrate = viper.GetUint("video_bitrate")

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

	s.MaxFPS = viper.GetInt("max_fps")
}
