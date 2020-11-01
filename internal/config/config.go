package config

import "github.com/spf13/cobra"

type Config interface {
	Init(cmd *cobra.Command) error
	Set()
}
