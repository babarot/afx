package config

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/h2non/filetype"
	"github.com/mholt/archiver"
)

// HTTP represents
type HTTP struct {
	Name string `hcl:"name,label"`

	URL         string `hcl:"url"`
	Output      string `hcl:"output"`
	Description string `hcl:"description,optional"`

	Plugin  *Plugin  `hcl:"plugin,block"`
	Command *Command `hcl:"command,block"`
}

// Init is
func (c HTTP) Init() error {
	var errs errors.Errors
	if c.HasPluginBlock() {
		errs.Append(c.Plugin.Init(c))
	}
	if c.HasCommandBlock() {
		errs.Append(c.Command.Init(c))
	}
	return errs.ErrorOrNil()
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

	os.MkdirAll(c.GetHome(), os.ModePerm)
	dest := filepath.Join(c.GetHome(), filepath.Base(c.URL))
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	if err := unarchive(dest); err != nil {
		return errors.Wrapf(err, "failed to unarchive: %s", dest)
	}

	return nil
}

// Install is
func (c HTTP) Install(ctx context.Context, status chan<- Status) error {
	if c.Installed() {
		return nil
	}

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
		status <- Status{Path: c.GetHome(), Done: true, Err: true}
		return err
	}

	var errs errors.Errors
	if c.HasPluginBlock() {
		errs.Append(c.Plugin.Install(c))
	}
	if c.HasCommandBlock() {
		errs.Append(c.Command.Install(c))
	}

	status <- Status{Path: c.GetHome(), Done: true, Err: errs.ErrorOrNil() != nil}
	return errs.ErrorOrNil()
}

func unarchive(f string) error {
	buf, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	switch {
	case filetype.IsArchive(buf):
		if err := archiver.Unarchive(f, filepath.Dir(f)); err != nil {
			return err
		}
		return nil
	default:
		log.Printf("[INFO] %s no need to unarchive\n", f)
		return nil
	}
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
func (c HTTP) HasPluginBlock() bool {
	return c.Plugin != nil
}

// HasCommandBlock is
func (c HTTP) HasCommandBlock() bool {
	return c.Command != nil
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
	var errs errors.Errors

	delete := func(f string, errs *errors.Errors) {
		err := os.RemoveAll(f)
		if err != nil {
			errs.Append(err)
			return
		}
		log.Printf("[INFO] Delete %s\n", f)
	}

	if c.HasCommandBlock() {
		links, err := c.Command.GetLink(c)
		if err != nil {
			return err
		}
		for _, link := range links {
			delete(link.From, &errs)
			delete(link.To, &errs)
		}
	}

	if c.HasPluginBlock() {
	}

	delete(c.GetHome(), &errs)

	return errs.ErrorOrNil()
}

// GetName returns a name
func (c HTTP) GetName() string {
	return c.Name
}

// GetHome returns a path
func (c HTTP) GetHome() string {
	u, _ := url.Parse(c.URL)
	return filepath.Join(os.Getenv("AFX_ROOT"), u.Host, filepath.Dir(u.Path))
}

// GetType returns a pacakge type
func (c HTTP) GetType() string {
	return "http"
}

// GetSlug returns a pacakge slug
func (c HTTP) GetSlug() string {
	return c.Name
}

// GetURL returns a URL related to the package
func (c HTTP) GetURL() string {
	return c.URL
}

// SetCommand sets given command to struct
func (c HTTP) SetCommand(command Command) Package {
	c.Command = &command
	return c
}

// SetPlugin sets given command to struct
func (c HTTP) SetPlugin(plugin Plugin) Package {
	c.Plugin = &plugin
	return c
}

// Objects returns file obejcts in the package
func (c HTTP) Objects() ([]string, error) {
	return []string{}, nil
}
