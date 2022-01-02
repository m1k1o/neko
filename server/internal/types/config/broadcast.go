package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Broadcast struct {
	Pipeline        string
	DefaultEndpoint string
	Enabled         bool
}

func (Broadcast) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("broadcast_pipeline", "", "custom gst pipeline used for broadcasting, strings {url} {device} {display} will be replaced")
	if err := viper.BindPFlag("broadcast_pipeline", cmd.PersistentFlags().Lookup("broadcast_pipeline")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("broadcast_default_endpoint", "", "default endpoint for broadcasting. Setting an endpoint will automatically enable broadcasting")
	if err := viper.BindPFlag("broadcast_default_endpoint", cmd.PersistentFlags().Lookup("broadcast_default_endpoint")); err != nil {
		return err
	}

	return nil
}

func (s *Broadcast) Set() {
	s.Pipeline = viper.GetString("broadcast_pipeline")
	s.DefaultEndpoint = viper.GetString("broadcast_default_endpoint")
	s.Enabled = s.DefaultEndpoint != ""
}
