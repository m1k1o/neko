package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"gitlab.com/demodesk/neko/server/internal/config"
	"gitlab.com/demodesk/neko/server/pkg/types"
)

type PluginsManagerCtx struct {
	logger  zerolog.Logger
	plugins map[string]types.Plugin
}

func New(config *config.Plugins) *PluginsManagerCtx {
	manager := &PluginsManagerCtx{
		logger:  log.With().Str("module", "plugins").Logger(),
		plugins: map[string]types.Plugin{},
	}

	if config.Enabled {
		err := manager.loadDir(config.Dir)
		manager.logger.Err(err).Msgf("loading finished, total %d plugins", len(manager.plugins))
	}

	return manager
}

func (manager *PluginsManagerCtx) loadDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		err = manager.load(path)
		manager.logger.Err(err).Str("plugin", path).Msg("loading a plugin")

		return nil
	})
}

func (manager *PluginsManagerCtx) load(path string) error {
	pl, err := plugin.Open(path)
	if err != nil {
		return err
	}

	sym, err := pl.Lookup("Plugin")
	if err != nil {
		return err
	}

	p, ok := sym.(types.Plugin)
	if !ok {
		return fmt.Errorf("not a valid plugin")
	}

	manager.plugins[path] = p
	return nil
}

func (manager *PluginsManagerCtx) InitConfigs(cmd *cobra.Command) {
	for path, plug := range manager.plugins {
		if err := plug.Config().Init(cmd); err != nil {
			log.Err(err).Str("plugin", path).Msg("unable to initialize configuration")
		}
	}
}

func (manager *PluginsManagerCtx) SetConfigs() {
	for _, plug := range manager.plugins {
		plug.Config().Set()
	}
}

func (manager *PluginsManagerCtx) Start(
	sessionManager types.SessionManager,
	webSocketManager types.WebSocketManager,
	apiManager types.ApiManager,
) {
	for _, plug := range manager.plugins {
		plug.Start(types.PluginManagers{
			SessionManager:   sessionManager,
			WebSocketManager: webSocketManager,
			ApiManager:       apiManager,
		})
	}
}

func (manager *PluginsManagerCtx) Shutdown() error {
	for path, plug := range manager.plugins {
		err := plug.Shutdown()
		manager.logger.Err(err).Str("plugin", path).Msg("plugin shutdown")
	}

	return nil
}
