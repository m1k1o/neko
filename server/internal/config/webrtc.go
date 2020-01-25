package config

import (
	"github.com/pion/webrtc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type WebRTC struct {
	Device      string
	AudioCodec  string
	AudioParams string
	Display     string
	VideoCodec  string
	VideoParams string
}

func (WebRTC) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("device", "auto_null.monitor", "Audio device to capture")
	if err := viper.BindPFlag("device", cmd.PersistentFlags().Lookup("device")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("aduio", "", "Audio codec parameters to use for streaming")
	if err := viper.BindPFlag("aparams", cmd.PersistentFlags().Lookup("aduio")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("display", ":0.0", "XDisplay to capture")
	if err := viper.BindPFlag("display", cmd.PersistentFlags().Lookup("display")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("video", "", "Video codec parameters to use for streaming")
	if err := viper.BindPFlag("vparams", cmd.PersistentFlags().Lookup("video")); err != nil {
		return err
	}

	// video codecs
	cmd.PersistentFlags().Bool("vp8", false, "Use VP8 codec")
	if err := viper.BindPFlag("vp8", cmd.PersistentFlags().Lookup("vp8")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("vp9", false, "Use VP9 codec")
	if err := viper.BindPFlag("vp9", cmd.PersistentFlags().Lookup("vp9")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("h264", false, "Use H264 codec")
	if err := viper.BindPFlag("h264", cmd.PersistentFlags().Lookup("h264")); err != nil {
		return err
	}

	// audio codecs
	cmd.PersistentFlags().Bool("opus", false, "Use Opus codec")
	if err := viper.BindPFlag("opus", cmd.PersistentFlags().Lookup("opus")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("g722", false, "Use G722 codec")
	if err := viper.BindPFlag("g722", cmd.PersistentFlags().Lookup("g722")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("pcmu", false, "Use PCMU codec")
	if err := viper.BindPFlag("pcmu", cmd.PersistentFlags().Lookup("pcmu")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("pcma", false, "Use PCMA codec")
	if err := viper.BindPFlag("pcmu", cmd.PersistentFlags().Lookup("pcmu")); err != nil {
		return err
	}

	return nil
}

func (s *WebRTC) Set() {
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
	s.AudioParams = viper.GetString("aparams")
	s.Display = viper.GetString("display")
	s.VideoCodec = videoCodec
	s.VideoParams = viper.GetString("vparams")
}
