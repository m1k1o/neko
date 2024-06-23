package filetransfer

import (
	"github.com/demodesk/neko/pkg/types"
)

type Plugin struct {
	config  *Config
	manager *Manager
}

func NewPlugin() *Plugin {
	return &Plugin{
		config: &Config{},
	}
}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Config() types.PluginConfig {
	return p.config
}

func (p *Plugin) Start(m types.PluginManagers) error {
	p.manager = NewManager(m.SessionManager, p.config)
	m.ApiManager.AddRouter("/filetransfer", p.manager.Route)
	m.WebSocketManager.AddHandler(p.manager.WebSocketHandler)
	return p.manager.Start()
}

func (p *Plugin) Shutdown() error {
	return p.manager.Shutdown()
}
