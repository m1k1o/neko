package plugins

import (
	"fmt"

	"github.com/rs/zerolog"

	"github.com/demodesk/neko/pkg/types"
)

type dependency struct {
	plugin    types.Plugin
	dependsOn []*dependency
	invoked   bool
	logger    zerolog.Logger
}

func (a *dependency) findPlugin(name string) (*dependency, bool) {
	if a == nil {
		return nil, false
	}

	if a.plugin.Name() == name {
		return a, true
	}

	for _, dep := range a.dependsOn {
		plug, ok := dep.findPlugin(name)
		if ok {
			return plug, true
		}
	}

	return nil, false
}

func (a *dependency) startPlugin(pm types.PluginManagers) error {
	if a.invoked {
		return nil
	}

	a.invoked = true

	for _, do := range a.dependsOn {
		if err := do.startPlugin(pm); err != nil {
			return fmt.Errorf("plugin's '%s' dependency: %w", a.plugin.Name(), err)
		}
	}

	err := a.plugin.Start(pm)
	if err != nil {
		return fmt.Errorf("plugin '%s' failed to start: %w", a.plugin.Name(), err)
	}

	a.logger.Info().Str("plugin", a.plugin.Name()).Msg("plugin started")
	return nil
}

type dependiencies struct {
	deps   map[string]*dependency
	logger zerolog.Logger
}

func (d *dependiencies) addPlugin(plugin types.Plugin) error {
	pluginName := plugin.Name()

	plug, ok := d.deps[pluginName]
	if !ok {
		plug = &dependency{}
	} else if plug.plugin != nil {
		return fmt.Errorf("plugin '%s' already added", pluginName)
	}

	plug.plugin = plugin
	plug.logger = d.logger
	d.deps[pluginName] = plug

	dplug, ok := plugin.(types.DependablePlugin)
	if !ok {
		return nil
	}

	for _, depName := range dplug.DependsOn() {
		dependsOn, ok := d.deps[depName]
		if !ok {
			dependsOn = &dependency{}
		} else if dependsOn.plugin != nil {
			// if there is a cyclical dependency, break it and return error
			if tdep, ok := dependsOn.findPlugin(pluginName); ok {
				dependsOn.dependsOn = nil
				delete(d.deps, pluginName)
				return fmt.Errorf("cyclical dependency detected: '%s' <-> '%s'", pluginName, tdep.plugin.Name())
			}
		}

		plug.dependsOn = append(plug.dependsOn, dependsOn)
		d.deps[depName] = dependsOn
	}

	return nil
}

func (d *dependiencies) findPlugin(name string) (*dependency, bool) {
	for _, dep := range d.deps {
		plug, ok := dep.findPlugin(name)
		if ok {
			return plug, true
		}
	}
	return nil, false
}

func (d *dependiencies) start(pm types.PluginManagers) error {
	for _, dep := range d.deps {
		if err := dep.startPlugin(pm); err != nil {
			return err
		}
	}
	return nil
}

func (d *dependiencies) forEach(f func(*dependency) error) error {
	for _, dep := range d.deps {
		if err := f(dep); err != nil {
			return err
		}
	}
	return nil
}

func (d *dependiencies) len() int {
	return len(d.deps)
}
