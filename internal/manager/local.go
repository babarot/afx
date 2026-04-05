package manager

import (
	"context"
	"errors"
	"os"

	pathutil "github.com/babarot/afx/internal/helpers/path"
	"github.com/babarot/afx/internal/runner"
	"github.com/babarot/afx/internal/state"
)

// Local represents
type Local struct {
	Name string `yaml:"name" validate:"required"`

	Directory   string `yaml:"directory" validate:"required"`
	Description string `yaml:"description"`

	Plugin  *Plugin  `yaml:"plugin"`
	Command *Command `yaml:"command"`

	DependsOn []string `yaml:"depends-on"`
}

// Init is
func (c Local) Init() error {
	var errs []error
	if c.HasPluginBlock() {
		if err := c.Plugin.Init(c); err != nil {
			errs = append(errs, err)
		}
	}
	if c.HasCommandBlock() {
		if err := c.Command.Init(c); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Install is
func (c Local) Install(ctx context.Context, status chan<- runner.Status) error {
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

func (c Local) HasReleaseBlock() bool {
	return false
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
	return pathutil.ExpandTilda(os.ExpandEnv(c.Directory))
}

func (c Local) GetDependsOn() []string {
	return c.DependsOn
}

func (c Local) GetResource() state.Resource {
	return getResource(c)
}

func (c Local) Check(ctx context.Context, status chan<- runner.Status) error {
	status <- runner.Status{Name: c.GetName(), Done: true, Err: false, Message: "(local)", NoColor: true}
	return nil
}
