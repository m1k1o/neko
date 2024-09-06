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
	"github.com/demodesk/neko/internal/plugins/chat"
	"github.com/demodesk/neko/internal/plugins/filetransfer"
	"github.com/demodesk/neko/pkg/types"
)

type ManagerCtx struct {
	logger  zerolog.Logger
	config  *config.Plugins
	plugins dependiencies
}

func New(config *config.Plugins) *ManagerCtx {
	manager := &ManagerCtx{
		logger: log.With().Str("module", "plugins").Logger(),
		config: config,
		plugins: dependiencies{
			deps: make(map[string]*dependency),
		},
	}

	manager.plugins.logger = manager.logger

	if config.Enabled {
		err := manager.loadDir(config.Dir)

		// only log error if plugin is not required
		if err != nil && config.Required {
			manager.logger.Fatal().Err(err).Msg("error loading plugins")
		}

		manager.logger.Info().Msgf("loading finished, total %d plugins", manager.plugins.len())
	}

	// add built-in plugins
	manager.plugins.addPlugin(filetransfer.NewPlugin())
	manager.plugins.addPlugin(chat.NewPlugin())

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

		// return error if plugin is required
		if err != nil && manager.config.Required {
			return err
		}

		// otherwise only log error if plugin is not required
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
		return fmt.Errorf("failed to add plugin: %w", err)
	}

	return nil
}

func (manager *ManagerCtx) InitConfigs(cmd *cobra.Command) {
	_ = manager.plugins.forEach(func(d *dependency) error {
		if err := d.plugin.Config().Init(cmd); err != nil {
			log.Err(err).Str("plugin", d.plugin.Name()).Msg("unable to initialize configuration")
		}
		return nil
	})
}

func (manager *ManagerCtx) SetConfigs() {
	_ = manager.plugins.forEach(func(d *dependency) error {
		d.plugin.Config().Set()
		return nil
	})
}

func (manager *ManagerCtx) Start(
	sessionManager types.SessionManager,
	webSocketManager types.WebSocketManager,
	apiManager types.ApiManager,
) {
	err := manager.plugins.start(types.PluginManagers{
		SessionManager:        sessionManager,
		WebSocketManager:      webSocketManager,
		ApiManager:            apiManager,
		LoadServiceFromPlugin: manager.LookupService,
	})

	if err != nil {
		if manager.config.Required {
			manager.logger.Fatal().Err(err).Msg("failed to start plugins, exiting...")
		} else {
			manager.logger.Err(err).Msg("failed to start plugins, skipping...")
		}
	}
}

func (manager *ManagerCtx) Shutdown() error {
	_ = manager.plugins.forEach(func(d *dependency) error {
		err := d.plugin.Shutdown()
		manager.logger.Err(err).Str("plugin", d.plugin.Name()).Msg("plugin shutdown")
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

func (manager *ManagerCtx) Metadata() []types.PluginMetadata {
	var plugins []types.PluginMetadata

	_ = manager.plugins.forEach(func(d *dependency) error {
		dependsOn := make([]string, 0)
		deps, isDependalbe := d.plugin.(types.DependablePlugin)
		if isDependalbe {
			dependsOn = deps.DependsOn()
		}

		_, isExposable := d.plugin.(types.ExposablePlugin)

		plugins = append(plugins, types.PluginMetadata{
			Name:         d.plugin.Name(),
			IsDependable: isDependalbe,
			IsExposable:  isExposable,
			DependsOn:    dependsOn,
		})

		return nil
	})

	return plugins
}
