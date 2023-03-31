package types

import (
	"errors"

	"github.com/spf13/cobra"
)

type Plugin interface {
	Name() string
	Config() PluginConfig
	Start(PluginManagers) error
	Shutdown() error
}

type DependablePlugin interface {
	Plugin
	DependsOn() []string
}

type ExposablePlugin interface {
	Plugin
	ExposeService() any
}

type PluginConfig interface {
	Init(cmd *cobra.Command) error
	Set()
}

type PluginMetadata struct {
	Name         string
	IsDependable bool
	IsExposable  bool
	DependsOn    []string `json:",omitempty"`
}

type PluginManagers struct {
	SessionManager        SessionManager
	WebSocketManager      WebSocketManager
	ApiManager            ApiManager
	LoadServiceFromPlugin func(string) (any, error)
}

func (p *PluginManagers) Validate() error {
	if p.SessionManager == nil {
		return errors.New("SessionManager is nil")
	}

	if p.WebSocketManager == nil {
		return errors.New("WebSocketManager is nil")
	}

	if p.ApiManager == nil {
		return errors.New("ApiManager is nil")
	}

	if p.LoadServiceFromPlugin == nil {
		return errors.New("LoadServiceFromPlugin is nil")
	}

	return nil
}
