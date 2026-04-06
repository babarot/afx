package manager

import (
	"errors"
	"os"
)

// Base contains fields common to all package types.
type Base struct {
	Name        string   `yaml:"name" validate:"required"`
	Description string   `yaml:"description"`
	Plugin      *Plugin  `yaml:"plugin"`
	Command     *Command `yaml:"command"`
	DependsOn   []string `yaml:"depends-on"`
}

// GetName returns the package name.
func (b Base) GetName() string { return b.Name }

// GetDependsOn returns the list of dependency names.
func (b Base) GetDependsOn() []string { return b.DependsOn }

// HasPluginBlock returns true if a plugin block is configured.
func (b Base) HasPluginBlock() bool { return b.Plugin != nil }

// HasCommandBlock returns true if a command block is configured.
func (b Base) HasCommandBlock() bool { return b.Command != nil }

// HasReleaseBlock returns false by default. GitHub overrides this.
func (b Base) HasReleaseBlock() bool { return false }

// GetPluginBlock returns the plugin configuration.
func (b Base) GetPluginBlock() Plugin {
	if b.Plugin != nil {
		return *b.Plugin
	}
	return Plugin{}
}

// GetCommandBlock returns the command configuration.
func (b Base) GetCommandBlock() Command {
	if b.Command != nil {
		return *b.Command
	}
	return Command{}
}

// initPackage runs the common initialization logic for plugin and command blocks.
func initPackage(plugin *Plugin, command *Command, pkg Package) error {
	var errs []error
	if plugin != nil {
		if err := plugin.Init(pkg); err != nil {
			errs = append(errs, err)
		}
	}
	if command != nil {
		if err := command.Init(pkg); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// installedPackage checks whether the package is installed by examining
// plugin sources, command symlinks, or home directory existence.
func installedPackage(plugin *Plugin, command *Command, pkg Package) bool {
	var list []bool
	if plugin != nil {
		list = append(list, plugin.Installed(pkg))
	}
	if command != nil {
		list = append(list, command.Installed(pkg))
	}
	if plugin == nil && command == nil {
		_, err := os.Stat(pkg.GetHome())
		list = append(list, err == nil)
	}
	return allTrue(list)
}
