package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Plugins struct {
	Enabled  bool
	Dir      string
	Required bool
}

func (Plugins) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("plugins.enabled", false, "load plugins in runtime")
	if err := viper.BindPFlag("plugins.enabled", cmd.PersistentFlags().Lookup("plugins.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("plugins.dir", "./bin/plugins", "path to neko plugins to load")
	if err := viper.BindPFlag("plugins.dir", cmd.PersistentFlags().Lookup("plugins.dir")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("plugins.required", false, "if true, neko will exit if there is an error when loading a plugin")
	if err := viper.BindPFlag("plugins.required", cmd.PersistentFlags().Lookup("plugins.required")); err != nil {
		return err
	}

	return nil
}

func (s *Plugins) Set() {
	s.Enabled = viper.GetBool("plugins.enabled")
	s.Dir = viper.GetString("plugins.dir")
	s.Required = viper.GetBool("plugins.required")
}
