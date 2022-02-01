package config

import (
	"context"
	"os"

	"github.com/b4b4r07/afx/pkg/errors"
)

// Local represents
type Local struct {
	Name string `yaml:"name"`

	Directory   string `yaml:"directory"`
	Description string `yaml:"description"`

	Plugin  *Plugin  `yaml:"plugin"`
	Command *Command `yaml:"command"`
}

// Init is
func (c Local) Init() error {
	var errs errors.Errors
	if c.HasPluginBlock() {
		errs.Append(c.Plugin.Init(c))
	}
	if c.HasCommandBlock() {
		errs.Append(c.Command.Init(c))
	}
	return errs.ErrorOrNil()
}

// Install is
func (c Local) Install(ctx context.Context, status chan<- Status) error {
	return nil
}

// Installed is
func (c Local) Installed() bool {
	return true
}

// HasPluginBlock is
func (c Local) HasPluginBlock() bool {
	return c.Plugin != nil
}

// HasCommandBlock is
func (c Local) HasCommandBlock() bool {
	return c.Command != nil
}

// GetPluginBlock is
func (c Local) GetPluginBlock() Plugin {
	if c.HasPluginBlock() {
		return *c.Plugin
	}
	return Plugin{}
}

// GetCommandBlock is
func (c Local) GetCommandBlock() Command {
	if c.HasCommandBlock() {
		return *c.Command
	}
	return Command{}
}

// Uninstall is
func (c Local) Uninstall(ctx context.Context) error {
	return nil
}

// GetName returns a name
func (c Local) GetName() string {
	return c.Name
}

// GetHome returns a path
func (c Local) GetHome() string {
	return expandTilda(os.ExpandEnv(c.Directory))
}

// GetType returns a pacakge type
func (c Local) GetType() string {
	return "local"
}

// GetURL returns a URL related to the package
func (c Local) GetURL() string {
	return ""
}
