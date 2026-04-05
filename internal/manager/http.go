package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/mholt/archiver"

	"github.com/babarot/afx/internal/data"
	"github.com/babarot/afx/internal/runner"
	"github.com/babarot/afx/internal/state"
	"github.com/babarot/afx/internal/templates"
)

// HTTP represents
type HTTP struct {
	Name string `yaml:"name" validate:"required"`

	URL         string `yaml:"url" validate:"required,url"`
	Description string `yaml:"description"`

	Plugin  *Plugin  `yaml:"plugin"`
	Command *Command `yaml:"command"`

	DependsOn []string  `yaml:"depends-on"`
	Templates Templates `yaml:"templates"`
}

type Templates struct {
	Replacements map[string]string `yaml:"replacements"`
}

// Init is
func (c HTTP) Init() error {
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

func (c HTTP) call(ctx context.Context) error {
	log.Printf("[TRACE] Get %s\n", c.URL)
	req, err := http.NewRequest(http.MethodGet, c.URL, nil)
	if err != nil {
		return err
	}

	client := new(http.Client)
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("[DEBUG] response code: %d", resp.StatusCode)
	switch resp.StatusCode {
	case 200, 301, 302:
		// go
	case 404:
		return fmt.Errorf("%s: %d Not Found in %s", c.GetName(), resp.StatusCode, c.URL)
	default:
		return fmt.Errorf("%s: %d %s", c.GetName(), resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	_ = os.MkdirAll(c.GetHome(), os.ModePerm)
	dest := filepath.Join(c.GetHome(), filepath.Base(c.URL))

	log.Printf("[DEBUG] http: %s: copying %q to %q", c.GetName(), c.URL, dest)
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	if err := unarchiveV2(dest); err != nil {
		return fmt.Errorf("failed to unarchive: %s: %w", dest, err)
	}

	return nil
}

// Install is
func (c HTTP) Install(ctx context.Context, status chan<- runner.Status) error {
	select {
	case <-ctx.Done():
		log.Println("[DEBUG] canceled")
		return nil
	default:
		// Go installing step!
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := c.call(ctx); err != nil {
		err = fmt.Errorf("%s: failed to make HTTP request: %w", c.Name, err)
		status <- runner.Status{Name: c.GetName(), Done: true, Err: true}
		return err
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

func unarchiveV2(path string) error {
	_, err := archiver.ByExtension(path)
	if err != nil {
		log.Printf("[DEBUG] unarchiveV2: no need to unarchive. finished with nil")
		return nil
	}
	return archiver.Unarchive(path, filepath.Dir(path))
}

// Installed is
func (c HTTP) Installed() bool {
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
func (c HTTP) HasPluginBlock() bool {
	return c.Plugin != nil
}

// HasCommandBlock is
func (c HTTP) HasCommandBlock() bool {
	return c.Command != nil
}

func (c HTTP) HasReleaseBlock() bool {
	return false
}

// GetPluginBlock is
func (c HTTP) GetPluginBlock() Plugin {
	if c.HasPluginBlock() {
		return *c.Plugin
	}
	return Plugin{}
}

// GetCommandBlock is
func (c HTTP) GetCommandBlock() Command {
	if c.HasCommandBlock() {
		return *c.Command
	}
	return Command{}
}

// Uninstall is
func (c HTTP) Uninstall(ctx context.Context) error {
	var errs []error

	del := func(f string) {
		err := os.RemoveAll(f)
		if err != nil {
			errs = append(errs, err)
			return
		}
		log.Printf("[INFO] Delete %s", f)
	}

	if c.HasCommandBlock() {
		links, _ := c.Command.GetLink(c)
		for _, link := range links {
			del(link.From)
			del(link.To)
		}
	}

	del(c.GetHome())

	return errors.Join(errs...)
}

// GetName returns a name
func (c HTTP) GetName() string {
	return c.Name
}

// GetHome returns a path
func (c HTTP) GetHome() string {
	u, _ := url.Parse(c.URL)
	return filepath.Join(DataDir(), u.Host, filepath.Dir(u.Path))
}

func (c HTTP) GetDependsOn() []string {
	return c.DependsOn
}

func (c HTTP) GetResource() state.Resource {
	return getResource(c)
}

func (c *HTTP) ParseURL() {
	templated, err := templates.New(data.New(data.WithPackage(c))).
		Replace(c.Templates.Replacements).
		Apply(c.URL)
	if err != nil {
		log.Printf("[ERROR] %s: failed to parse URL", c.GetName())
		return
	}
	if templated != c.URL {
		log.Printf("[TRACE] %s: templating URL %q to %q", c.GetName(), c.URL, templated)
		c.URL = templated
	}
}

func (c HTTP) Check(ctx context.Context, status chan<- runner.Status) error {
	status <- runner.Status{Name: c.GetName(), Done: true, Err: false, Message: "(http)", NoColor: true}
	return nil
}
