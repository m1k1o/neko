package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Broadcast struct {
	Pipeline string
}

func (Broadcast) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("broadcast_pipeline", "", "custom gst pipeline used for broadcasting, strings {url} {device} {display} will be replaced")
	if err := viper.BindPFlag("broadcast_pipeline", cmd.PersistentFlags().Lookup("broadcast_pipeline")); err != nil {
		return err
	}

	return nil
}

func (s *Broadcast) Set() {
	s.Pipeline = viper.GetString("broadcast_pipeline")
}
