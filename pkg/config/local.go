package config

import (
	"context"

	"github.com/b4b4r07/afx/pkg/errors"
)

// Local represents
type Local struct {
	Name string `hcl:"name,label"`

	Directory   string `hcl:"directory"`
	Description string `hcl:"description,optional"`

	Plugin  *Plugin  `hcl:"plugin,block"`
	Command *Command `hcl:"command,block"`
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
	return c.Directory
}

// GetType returns a pacakge type
func (c Local) GetType() string {
	return "local"
}

// GetSlug returns a pacakge type
func (c Local) GetSlug() string {
	return c.Name
}

// GetURL returns a URL related to the package
func (c Local) GetURL() string {
	return ""
}

// SetCommand sets given command to struct
func (c Local) SetCommand(command Command) Package {
	c.Command = &command
	return c
}

// SetPlugin sets given command to struct
func (c Local) SetPlugin(plugin Plugin) Package {
	c.Plugin = &plugin
	return c
}

// Objects returns file obejcts in the package
func (c Local) Objects() ([]string, error) {
	return []string{}, nil
}
