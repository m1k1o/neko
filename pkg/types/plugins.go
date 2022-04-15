package types

import (
	"github.com/spf13/cobra"
)

type Plugin interface {
	Config() PluginConfig
	Start(PluginManagers)
	Shutdown() error
}

type PluginConfig interface {
	Init(cmd *cobra.Command) error
	Set()
}

type PluginManagers struct {
	SessionManager   SessionManager
	WebSocketManager WebSocketManager
	ApiManager       ApiManager
}
