package manager

import (
	"errors"
	"os"
	"path/filepath"
)

// GitHub represents GitHub repository
type GitHub struct {
	Name string `yaml:"name" validate:"required"`

	Owner       string `yaml:"owner"       validate:"required"`
	Repo        string `yaml:"repo"        validate:"required"`
	Description string `yaml:"description"`

	Branch string        `yaml:"branch"`
	Option *GitHubOption `yaml:"with"`

	Release *GitHubRelease `yaml:"release"`

	Plugin  *Plugin   `yaml:"plugin"`
	Command *Command  `yaml:"command" validate:"required_with=Release"` // TODO: (not required Release)
	As      *GitHubAs `yaml:"as"`

	DependsOn []string `yaml:"depends-on"`
}

type GitHubAs struct {
	GHExtension *GHExtension `yaml:"gh-extension"`
}

type GHExtension struct {
	Name     string `yaml:"name" validate:"required,startswith=gh-"`
	Tag      string `yaml:"tag"`
	RenameTo string `yaml:"rename-to" validate:"startswith-gh-if-not-empty,excludesall=/"`
}

type GitHubOption struct {
	Depth int `yaml:"depth"`
}

// GitHubRelease represents a GitHub release structure
type GitHubRelease struct {
	Name string `yaml:"name" validate:"required"`
	Tag  string `yaml:"tag"`

	Asset GitHubReleaseAsset `yaml:"asset"`
}

type GitHubReleaseAsset struct {
	Filename     string            `yaml:"filename"`
	Replacements map[string]string `yaml:"replacements"`
}

// Init runs initialization step related to GitHub packages
func (c GitHub) Init() error {
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

// Installed returns true the GitHub package is already installed
func (c GitHub) Installed() bool {
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

func (c GitHub) GetReleaseTag() string {
	if c.Release != nil {
		return c.Release.Tag
	}
	return "latest"
}

func (c GitHub) HasPluginBlock() bool {
	return c.Plugin != nil
}

func (c GitHub) HasCommandBlock() bool {
	return c.Command != nil
}

func (c GitHub) HasReleaseBlock() bool {
	return c.Release != nil
}

func (c GitHub) GetPluginBlock() Plugin {
	if c.HasPluginBlock() {
		return *c.Plugin
	}
	return Plugin{}
}

func (c GitHub) GetCommandBlock() Command {
	if c.HasCommandBlock() {
		return *c.Command
	}
	return Command{}
}

// GetName returns a name
func (c GitHub) GetName() string {
	return c.Name
}

// GetHome returns a path
func (c GitHub) GetHome() string {
	if c.IsGHExtension() {
		return c.As.GHExtension.GetHome()
	}
	return filepath.Join(DataDir(), "github.com", c.Owner, c.Repo)
}

func (c GitHub) GetDependsOn() []string {
	return c.DependsOn
}
