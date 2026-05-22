package telemetry

import (
	"github.com/m1k1o/neko/server/pkg/types"
)

// Plugin forwards neko session lifecycle events to kernel-images-api as
// live_view_connect / live_view_disconnect telemetry events.
type Plugin struct {
	config  *Config
	manager *Manager
}

// NewPlugin constructs the telemetry plugin in disabled state; Start is a
// no-op until Config.Enabled is true.
func NewPlugin() *Plugin {
	return &Plugin{
		config: &Config{},
	}
}

// Name returns the plugin identifier used by the plugin manager.
func (p *Plugin) Name() string {
	return PluginName
}

// Config exposes the underlying config struct for flag binding.
func (p *Plugin) Config() types.PluginConfig {
	return p.config
}

// Start wires the lifecycle listeners and spins the worker goroutine.
func (p *Plugin) Start(m types.PluginManagers) error {
	p.manager = NewManager(m.SessionManager, p.config)
	return p.manager.Start()
}

// Shutdown stops the worker goroutine and drains in-flight events.
func (p *Plugin) Shutdown() error {
	if p.manager == nil {
		return nil
	}
	return p.manager.Shutdown()
}
