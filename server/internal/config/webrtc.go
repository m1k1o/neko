package config

import (
	"strings"

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

	cmd.PersistentFlags().String("ac", "opus", "Audio codec to use for streaming")
	if err := viper.BindPFlag("acodec", cmd.PersistentFlags().Lookup("ac")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("ap", "", "Audio codec parameters to use for streaming")
	if err := viper.BindPFlag("aparams", cmd.PersistentFlags().Lookup("ap")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("display", ":0.0", "XDisplay to capture")
	if err := viper.BindPFlag("display", cmd.PersistentFlags().Lookup("display")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("vc", "vp8", "Video codec to use for streaming")
	if err := viper.BindPFlag("vcodec", cmd.PersistentFlags().Lookup("vc")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("vp", "", "Video codec parameters to use for streaming")
	if err := viper.BindPFlag("vparams", cmd.PersistentFlags().Lookup("vp")); err != nil {
		return err
	}

	return nil
}

func (s *WebRTC) Set() {
	s.Device = strings.ToLower(viper.GetString("device"))
	s.AudioCodec = strings.ToLower(viper.GetString("acodec"))
	s.AudioParams = strings.ToLower(viper.GetString("aparams"))
	s.Display = strings.ToLower(viper.GetString("display"))
	s.VideoCodec = strings.ToLower(viper.GetString("vcodec"))
	s.VideoParams = strings.ToLower(viper.GetString("vparams"))
}
