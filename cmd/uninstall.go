package cmd

import (
	"fmt"
	"os"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/state"
	"github.com/b4b4r07/afx/pkg/templates"
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
		afx uninstall [args...]

		By default, it tries to uninstall all packages deleted from config file.
		If any args are given, it tries to uninstall only them.
		But it's needed also to be deleted from config file.
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
		Args:                  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.meta.init(args); err != nil {
				return err
			}

			resources := c.State.Deletions
			if len(resources) == 0 {
				c.UI.Output("No packages to uninstall")
				return nil
			}

			// not uninstall all old packages. Instead just only uninstall
			// given packages when not uninstalled yet.
			var given []state.Resource
			for _, arg := range args {
				resource, err := c.getFromDeletions(arg)
				if err != nil {
					// no hit in deletions
					continue
				}
				given = append(given, resource)
			}
			if len(given) > 0 {
				resources = given
			}

			return c.run(resources)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return c.meta.printForUpdate()
		},
	}

	return uninstallCmd
}

func (c *uninstallCmd) run(resources []state.Resource) error {
	var errs errors.Errors

	for _, resource := range resources {
		err := delete(append(resource.Paths, resource.Home)...)
		if err != nil {
			errs.Append(err)
			continue
		}
		c.State.Remove(resource.Name)
		fmt.Printf("deleted %s\n", resource.Home)
	}

	return errs.ErrorOrNil()
}

func delete(paths ...string) error {
	var errs errors.Errors
	for _, path := range paths {
		errs.Append(os.RemoveAll(path))
	}
	return errs.ErrorOrNil()
}

func (c *uninstallCmd) getFromDeletions(name string) (state.Resource, error) {
	resources := c.State.Deletions

	for _, resource := range resources {
		if resource.Name == name {
			return resource, nil
		}
	}

	return state.Resource{}, errors.New("not found")
}
