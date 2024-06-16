package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/demodesk/neko/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	ErrPluginSettingsNotFound = errors.New("plugin settings not found")
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

type PluginSettings map[string]any

func (p PluginSettings) Unmarshal(name string, def any) error {
	if p == nil {
		return fmt.Errorf("%w: %s", ErrPluginSettingsNotFound, name)
	}
	// loop through the plugin settings and take only the one that starts with the name
	// because the settings are stored in a map["plugin_name.setting_name"] = value
	newMap := make(map[string]any)
	for k, v := range p {
		if strings.HasPrefix(k, name+".") {
			newMap[strings.TrimPrefix(k, name+".")] = v
		}
	}
	fmt.Printf("newMap: %+v\n", newMap)
	if len(newMap) == 0 {
		return fmt.Errorf("%w: %s", ErrPluginSettingsNotFound, name)
	}
	return utils.Decode(newMap, def)
}
