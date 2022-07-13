package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/demodesk/neko/internal/config"
	"github.com/demodesk/neko/pkg/types"
)

type ManagerCtx struct {
	logger  zerolog.Logger
	plugins dependiencies
}

func New(config *config.Plugins) *ManagerCtx {
	manager := &ManagerCtx{
		logger: log.With().Str("module", "plugins").Logger(),
		plugins: dependiencies{
			deps: make(map[string]*dependency),
		},
	}

	manager.plugins.logger = manager.logger

	if config.Enabled {
		err := manager.loadDir(config.Dir)
		manager.logger.Err(err).Msgf("loading finished, total %d plugins", manager.plugins.len())
	}

	return manager
}

func (manager *ManagerCtx) loadDir(dir string) error {
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

func (manager *ManagerCtx) load(path string) error {
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

	if err = manager.plugins.addPlugin(p); err != nil {
		return fmt.Errorf("failed to add plugin '%s': %w", p.Name(), err)
	}

	manager.logger.Info().Msgf("loaded plugin '%s', total %d plugins", p.Name(), manager.plugins.len())
	return nil
}

func (manager *ManagerCtx) InitConfigs(cmd *cobra.Command) {
	_ = manager.plugins.forEach(func(plug *dependency) error {
		if err := plug.plugin.Config().Init(cmd); err != nil {
			log.Err(err).Str("plugin", plug.plugin.Name()).Msg("unable to initialize configuration")
		}
		return nil
	})
}

func (manager *ManagerCtx) SetConfigs() {
	_ = manager.plugins.forEach(func(plug *dependency) error {
		plug.plugin.Config().Set()
		return nil
	})
}

func (manager *ManagerCtx) Start(
	sessionManager types.SessionManager,
	webSocketManager types.WebSocketManager,
	apiManager types.ApiManager,
) {
	_ = manager.plugins.start(types.PluginManagers{
		SessionManager:        sessionManager,
		WebSocketManager:      webSocketManager,
		ApiManager:            apiManager,
		LoadServiceFromPlugin: manager.LookupService,
	})
}

func (manager *ManagerCtx) Shutdown() error {
	_ = manager.plugins.forEach(func(plug *dependency) error {
		err := plug.plugin.Shutdown()
		manager.logger.Err(err).Str("plugin", plug.plugin.Name()).Msg("plugin shutdown")
		return nil
	})
	return nil
}

func (manager *ManagerCtx) LookupService(pluginName string) (any, error) {
	plug, ok := manager.plugins.findPlugin(pluginName)
	if !ok {
		return nil, fmt.Errorf("plugin '%s' not found", pluginName)
	}

	expPlug, ok := plug.plugin.(types.ExposablePlugin)
	if !ok {
		return nil, fmt.Errorf("plugin '%s' is not exposable", pluginName)
	}

	return expPlug.ExposeService(), nil
}
