package cmd

import (
	"os"

	"github.com/b4b4r07/afx/pkg/config"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/k0kubun/pp"
	"github.com/spf13/cobra"
)

type uninstallCmd struct {
	meta
}

var (
	// uninstallLong is long description of uninstall command
	uninstallLong = templates.LongDesc(``)

	// uninstallExample is examples for uninstall command
	uninstallExample = templates.Examples(`
		# Normal
		afx uninstall
	`)
)

// newUninstallCmd creates a new uninstall command
func newUninstallCmd() *cobra.Command {
	c := &uninstallCmd{}

	uninstallCmd := &cobra.Command{
		Use:                   "uninstall",
		Short:                 "Uninstall installed packages",
		Long:                  uninstallLong,
		Example:               uninstallExample,
		Aliases:               []string{"rm", "un"},
		SuggestFor:            []string{"delete"},
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.meta.init(args); err != nil {
				return err
			}
			if c.parseErr != nil {
				return c.parseErr
			}
			return c.run(args)
		},
	}

	return uninstallCmd
}

func (c *uninstallCmd) run(args []string) error {
	pp.Println("should install", c.State.CheckInstall())
	pp.Println("should uninstall", c.State.CheckUninstall())
	return nil

	var pkgs []config.Package
	for _, resource := range c.State.Resources {
		if !resource.Valid() {
			pkgs = append(pkgs, c.get(resource.Name))
		}
	}

	var errs errors.Errors

	for _, pkg := range pkgs {
		errs.Append(os.RemoveAll(pkg.GetHome()))

		switch {
		case pkg.HasPluginBlock():
		case pkg.HasCommandBlock():
		}
	}

	return errs.ErrorOrNil()
}
