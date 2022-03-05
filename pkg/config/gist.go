package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/b4b4r07/afx/internal/diags"
	"github.com/b4b4r07/afx/pkg/state"
	git "gopkg.in/src-d/go-git.v4"
)

// Gist represents
type Gist struct {
	Name string `yaml:"name" validate:"required"`

	Owner       string `yaml:"owner" validate:"required"`
	ID          string `yaml:"id" validate:"required"`
	Description string `yaml:"description"`

	Plugin  *Plugin  `yaml:"plugin"`
	Command *Command `yaml:"command"`

	DependsOn []string `yaml:"depends-on"`
}

// Init is
func (c Gist) Init() error {
	var errs diags.Error
	if c.HasPluginBlock() {
		errs.Append(c.Plugin.Init(c))
	}
	if c.HasCommandBlock() {
		errs.Append(c.Command.Init(c))
	}
	return errs.ErrorOrNil()
}

// Install is
func (c Gist) Install(ctx context.Context, status chan<- Status) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("[DEBUG] canceled")
		return nil
	default:
		// Go installing step!
	}

	if _, err := os.Stat(c.GetHome()); err == nil {
		log.Printf("[DEBUG] %s: removed because already exists before clone gist: %s",
			c.GetName(), c.GetHome())
		os.RemoveAll(c.GetHome())
	}

	_, err := git.PlainCloneContext(ctx, c.GetHome(), false, &git.CloneOptions{
		URL:  fmt.Sprintf("https://gist.github.com/%s/%s", c.Owner, c.ID),
		Tags: git.NoTags,
	})
	if err != nil {
		status <- Status{Name: c.GetName(), Done: true, Err: true}
		return diags.Wrapf(err, "%s: failed to clone gist repo", c.Name)
	}

	var errs diags.Error
	if c.HasPluginBlock() {
		errs.Append(c.Plugin.Install(c))
	}
	if c.HasCommandBlock() {
		errs.Append(c.Command.Install(c))
	}

	status <- Status{Name: c.GetName(), Done: true, Err: errs.ErrorOrNil() != nil}
	return errs.ErrorOrNil()
}

// Installed is
func (c Gist) Installed() bool {
	var list []bool

	if c.HasPluginBlock() {
		list = append(list, c.Plugin.Installed(c))
	}

	if c.HasCommandBlock() {
		list = append(list, c.Command.Installed(c))
	}

	switch {
	case c.HasPluginBlock():
	case c.HasCommandBlock():
	default:
		_, err := os.Stat(c.GetHome())
		list = append(list, err == nil)
	}

	return check(list)
}

// HasPluginBlock is
func (c Gist) HasPluginBlock() bool {
	return c.Plugin != nil
}

// HasCommandBlock is
func (c Gist) HasCommandBlock() bool {
	return c.Command != nil
}

// GetPluginBlock is
func (c Gist) GetPluginBlock() Plugin {
	if c.HasPluginBlock() {
		return *c.Plugin
	}
	return Plugin{}
}

// GetCommandBlock is
func (c Gist) GetCommandBlock() Command {
	if c.HasCommandBlock() {
		return *c.Command
	}
	return Command{}
}

// Uninstall is
func (c Gist) Uninstall(ctx context.Context) error {
	var errs diags.Error

	delete := func(f string) error {
		err := os.RemoveAll(f)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Delete %s\n", f)
		return nil
	}

	if c.HasCommandBlock() {
		links, _ := c.Command.GetLink(c)
		for _, link := range links {
			errs.Append(delete(link.From))
			errs.Append(delete(link.To))
		}
	}

	if c.HasPluginBlock() {
		// TODO
	}

	errs.Append(delete(c.GetHome()))
	return errs.ErrorOrNil()
}

// GetName returns a name
func (c Gist) GetName() string {
	return c.Name
}

// GetHome returns a path
func (c Gist) GetHome() string {
	return filepath.Join(os.Getenv("HOME"), ".afx", "gist.github.com", c.Owner, c.ID)
}

func (c Gist) GetDependsOn() []string {
	return c.DependsOn
}

func (c Gist) GetResource() state.Resource {
	return getResource(c)
}
