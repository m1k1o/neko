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

type dependiencies struct {
	deps   map[string]*dependency
	logger zerolog.Logger
}

func (d *dependiencies) addPlugin(plugin types.Plugin) error {
	plug, ok := d.deps[plugin.Name()]
	if !ok {
		plug = &dependency{}
	} else if plug.plugin != nil {
		return fmt.Errorf("plugin '%s' already added", plugin.Name())
	}

	plug.plugin = plugin
	plug.logger = d.logger
	d.deps[plugin.Name()] = plug

	dplug, ok := plugin.(types.DependablePlugin)
	if !ok {
		return nil
	}

	for _, dep := range dplug.DependsOn() {
		var dependsOn *dependency
		dependsOn, ok = d.deps[dep]
		if !ok {
			dependsOn = &dependency{}
		} else if dependsOn.plugin != nil {
			// if there is a cyclical dependency, break it and return error
			if tdep := dependsOn.findPlugin(plugin.Name()); tdep != nil {
				dependsOn.dependsOn = nil
				delete(d.deps, plugin.Name())
				return fmt.Errorf("cyclical dependency detected: '%s' <-> '%s'", plugin.Name(), tdep.plugin.Name())
			}
		}

		plug.dependsOn = append(plug.dependsOn, dependsOn)
		d.deps[dep] = dependsOn
	}

	return nil
}

func (d *dependiencies) start(pm types.PluginManagers) error {
	for _, p := range d.deps {
		if err := p.start(pm); err != nil {
			return err
		}
	}
	return nil
}

func (d *dependiencies) findPlugin(name string) (*dependency, bool) {
	for _, p := range d.deps {
		if found := p.findPlugin(name); found != nil {
			return found, true
		}
	}
	return nil, false
}

func (a *dependency) findPlugin(name string) *dependency {
	if a == nil {
		return nil
	}

	if a.plugin.Name() == name {
		return a
	}

	for _, p := range a.dependsOn {
		if found := p.findPlugin(name); found != nil {
			return found
		}
	}

	return nil
}

func (d *dependiencies) forEach(f func(*dependency) error) error {
	for _, dp := range d.deps {
		if err := f(dp); err != nil {
			return err
		}
	}
	return nil
}

func (d *dependiencies) len() int {
	return len(d.deps)
}

func (a *dependency) start(pm types.PluginManagers) error {
	if a.invoked {
		return nil
	}

	a.invoked = true

	for _, do := range a.dependsOn {
		if err := do.start(pm); err != nil {
			return err
		}
	}

	err := a.plugin.Start(pm)
	a.logger.Err(err).Str("plugin", a.plugin.Name()).Msg("plugin start")
	if err != nil {
		return fmt.Errorf("plugin %s failed to start: %s", a.plugin.Name(), err)
	}

	return nil
}
