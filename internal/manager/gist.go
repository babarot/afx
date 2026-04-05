package manager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/babarot/afx/internal/git"
	"github.com/babarot/afx/internal/runner"
	"github.com/babarot/afx/internal/state"
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
func (c Gist) Install(ctx context.Context, status chan<- runner.Status) error {
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

	gitCmd := git.NewRunner()
	url := fmt.Sprintf("https://gist.github.com/%s/%s", c.Owner, c.ID)
	if _, err := gitCmd.Run(ctx, "clone", "--no-tags", url, c.GetHome()); err != nil {
		status <- runner.Status{Name: c.GetName(), Done: true, Err: true}
		if git.IsAuthError(err) {
			return fmt.Errorf("%s: authentication failed. Please set GITHUB_TOKEN or configure git credentials: %w", c.Name, err)
		}
		return fmt.Errorf("%s: failed to clone gist repo: %w", c.Name, err)
	}

	var errs []error
	if c.HasPluginBlock() {
		if err := c.Plugin.Install(c); err != nil {
			errs = append(errs, err)
		}
	}
	if c.HasCommandBlock() {
		if err := c.Command.Install(c); err != nil {
			errs = append(errs, err)
		}
	}

	status <- runner.Status{Name: c.GetName(), Done: true, Err: errors.Join(errs...) != nil}
	return errors.Join(errs...)
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

	if !c.HasPluginBlock() && !c.HasCommandBlock() {
		_, err := os.Stat(c.GetHome())
		list = append(list, err == nil)
	}

	return allTrue(list)
}

// HasPluginBlock is
func (c Gist) HasPluginBlock() bool {
	return c.Plugin != nil
}

// HasCommandBlock is
func (c Gist) HasCommandBlock() bool {
	return c.Command != nil
}

func (c Gist) HasReleaseBlock() bool {
	return false
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
	var errs []error

	del := func(f string) {
		err := os.RemoveAll(f)
		if err != nil {
			errs = append(errs, err)
			return
		}
		log.Printf("[INFO] Delete %s\n", f)
	}

	if c.HasCommandBlock() {
		links, err := c.Command.GetLink(c)
		if err != nil {
			return fmt.Errorf("%s: failed to get command.link: %w", c.Name, err)
		}
		for _, link := range links {
			del(link.From)
			del(link.To)
		}
	}

	del(c.GetHome())

	return errors.Join(errs...)
}

// GetName returns a name
func (c Gist) GetName() string {
	return c.Name
}

// GetHome returns a path
func (c Gist) GetHome() string {
	return filepath.Join(DataDir(), "gist.github.com", c.Owner, c.ID)
}

func (c Gist) GetDependsOn() []string {
	return c.DependsOn
}

func (c Gist) GetResource() state.Resource {
	return getResource(c)
}

func (c Gist) Check(ctx context.Context, status chan<- runner.Status) error {
	status <- runner.Status{Name: c.GetName(), Done: true, Err: false, Message: "(gist)", NoColor: true}
	return nil
}
