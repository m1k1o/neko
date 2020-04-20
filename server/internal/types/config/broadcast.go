package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Broadcast struct {
	Enabled     bool
	Display     string
	Device      string
	AudioParams string
	VideoParams string
	RTMP        string
}

func (Broadcast) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("broadcast", false, "use PCMA audio codec")
	if err := viper.BindPFlag("broadcast", cmd.PersistentFlags().Lookup("broadcast")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("rtmp", "", "RMTP url for broadcasting")
	if err := viper.BindPFlag("rtmp", cmd.PersistentFlags().Lookup("rtmp")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("cast_audio", "", "audio codec parameters to use for broadcasting")
	if err := viper.BindPFlag("cast_audio", cmd.PersistentFlags().Lookup("cast_audio")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("cast_video", "", "video codec parameters to use for broadcasting")
	if err := viper.BindPFlag("cast_video", cmd.PersistentFlags().Lookup("cast_video")); err != nil {
		return err
	}

	return nil
}

func (s *Broadcast) Set() {
	s.Enabled = viper.GetBool("broadcast")
	s.Display = viper.GetString("display")
	s.Device = viper.GetString("device")
	s.AudioParams = viper.GetString("cast_audio")
	s.VideoParams = viper.GetString("cast_video")
	s.RTMP = viper.GetString("rtmp")
}
