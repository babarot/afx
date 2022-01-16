package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/b4b4r07/afx/pkg/errors"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Gist represents
type Gist struct {
	Name string `yaml:"name,label"`

	Owner       string `yaml:"owner"`
	ID          string `yaml:"id"`
	Description string `yaml:"description,optional"`

	Plugin  *Plugin  `yaml:"plugin,block"`
	Command *Command `yaml:"command,block"`
}

func NewGist(owner, id string) (Gist, error) {
	type data struct {
		Description string `json:"description"`
	}
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/gists/%s", id))
	if err != nil {
		return Gist{}, err
	}
	defer resp.Body.Close()
	// if res.StatusCode != 200 {
	//   fmt.Println("StatusCode=%d", res.StatusCode)
	//   return
	// }
	var d data
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		return Gist{}, err
	}
	return Gist{
		Name:        id,
		Owner:       owner,
		ID:          id,
		Description: d.Description,
	}, nil
}

// Init is
func (c Gist) Init() error {
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
func (c Gist) Install(ctx context.Context, status chan<- Status) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

	_, err := git.PlainCloneContext(ctx, c.GetHome(), false, &git.CloneOptions{
		URL:  fmt.Sprintf("https://gist.github.com/%s/%s", c.Owner, c.ID),
		Tags: git.NoTags,
	})
	if err != nil {
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
		// TODO
	}

	delete(c.GetHome(), &errs)

	return errs.ErrorOrNil()
}

// GetName returns a name
func (c Gist) GetName() string {
	return c.Name
}

// GetHome returns a path
func (c Gist) GetHome() string {
	return filepath.Join(os.Getenv("AFX_ROOT"), "gist.github.com", c.Owner, c.ID)
}

// GetType returns a pacakge type
func (c Gist) GetType() string {
	return "gist"
}

// GetSlug returns a pacakge slug
func (c Gist) GetSlug() string {
	return fmt.Sprintf("%s/%s", c.Owner, c.ID)
}

// GetURL returns a URL related to the package
func (c Gist) GetURL() string {
	return path.Join("https://gist.github.com", c.Owner, c.ID)
}

// SetCommand sets given command to struct
func (c Gist) SetCommand(command Command) Package {
	c.Command = &command
	return c
}

// SetPlugin sets given command to struct
func (c Gist) SetPlugin(plugin Plugin) Package {
	c.Plugin = &plugin
	return c
}

// Objects returns file obejcts in the package
func (c Gist) Objects() ([]string, error) {
	var paths []string
	fs := memfs.New()
	storer := memory.NewStorage()
	r, err := git.Clone(storer, fs, &git.CloneOptions{
		URL: fmt.Sprintf("https://gist.github.com/%s/%s", c.Owner, c.ID),
	})
	if err != nil {
		return paths, err
	}
	head, err := r.Head()
	if err != nil {
		return paths, err
	}
	commit, err := r.CommitObject(head.Hash())
	if err != nil {
		return paths, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return paths, err
	}
	for _, entry := range tree.Entries {
		paths = append(paths, entry.Name)
	}
	return paths, nil
}
