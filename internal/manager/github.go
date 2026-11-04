package manager

import (
	"fmt"
	"path/filepath"

	"github.com/babarot/afx/internal/gh"
	"github.com/babarot/afx/internal/state"
)

// GitHub represents GitHub repository
type GitHub struct {
	Base `yaml:",inline"`

	Owner  string `yaml:"owner" validate:"required"`
	Repo   string `yaml:"repo"  validate:"required"`
	Branch string `yaml:"branch"`

	Option  *GitHubOption  `yaml:"with"`
	Release *GitHubRelease `yaml:"release"`
	As      *GitHubAs      `yaml:"as"`

	GHRunner gh.Runner `yaml:"-"`
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
	return initPackage(c.Plugin, c.Command, c)
}

// Installed returns true if the GitHub package is already installed
func (c GitHub) Installed() bool {
	return installedPackage(c.Plugin, c.Command, c)
}

// HasReleaseBlock overrides Base to check the Release field.
func (c GitHub) HasReleaseBlock() bool {
	return c.Release != nil
}

func (c GitHub) GetReleaseTag() string {
	if c.Release != nil {
		return c.Release.Tag
	}
	return "latest"
}

// GetHome returns the installation path for this package.
func (c GitHub) GetHome() string {
	if c.IsGHExtension() {
		return c.As.GHExtension.GetHome()
	}
	return filepath.Join(DataDir(), "github.com", c.Owner, c.Repo)
}

func (c GitHub) GetResource() state.Resource {
	return getResource(c)
}

// ResourceMeta implementation

func (c GitHub) ResourceType() string {
	if c.IsGHExtension() {
		return "GitHub (gh extension)"
	}
	if c.HasReleaseBlock() {
		return "GitHub Release"
	}
	return "GitHub"
}

func (c GitHub) ResourceID() string {
	if c.HasReleaseBlock() {
		return fmt.Sprintf("github.com/release/%s/%s", c.Owner, c.Repo)
	}
	return fmt.Sprintf("github.com/%s/%s", c.Owner, c.Repo)
}

func (c GitHub) ResourceVersion() string {
	if c.HasReleaseBlock() {
		return c.Release.Tag
	}
	return ""
}

func (c GitHub) ResourceExtraPaths() []string {
	if c.IsGHExtension() {
		if alias := c.As.GHExtension.GetAliasHome(); alias != "" {
			return []string{alias}
		}
	}
	return nil
}
