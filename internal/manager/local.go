package manager

import (
	"context"
	"fmt"
	"os"

	pathutil "github.com/babarot/afx/internal/helpers/path"
	"github.com/babarot/afx/internal/runner"
	"github.com/babarot/afx/internal/state"
)

// Local represents a local directory package.
type Local struct {
	Base `yaml:",inline"`

	Directory string `yaml:"directory" validate:"required"`
}

func (c Local) Init() error { return initPackage(c.Plugin, c.Command, c) }

// Installed always returns true for local packages.
func (c Local) Installed() bool { return true }

func (c Local) GetHome() string {
	return pathutil.ExpandTilda(os.ExpandEnv(c.Directory))
}

func (c Local) GetResource() state.Resource { return getResource(c) }

func (c Local) Install(ctx context.Context, status chan<- runner.Status) error {
	return nil
}

func (c Local) Uninstall(ctx context.Context) error {
	return nil
}

func (c Local) Check(ctx context.Context, status chan<- runner.Status) error {
	status <- runner.Status{Name: c.GetName(), Done: true, Err: false, Message: "(local)", NoColor: true}
	return nil
}

// ResourceMeta implementation

func (c Local) ResourceType() string         { return "Local" }
func (c Local) ResourceID() string           { return fmt.Sprintf("local/%s", c.Directory) }
func (c Local) ResourceVersion() string      { return "" }
func (c Local) ResourceExtraPaths() []string { return nil }
