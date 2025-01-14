package cmd

import (
	"fmt"
	"os"

	"github.com/babarot/afx/pkg/errors"
	"github.com/babarot/afx/pkg/helpers/templates"
	"github.com/babarot/afx/pkg/state"
	"github.com/spf13/cobra"
)

type uninstallCmd struct {
	metaCmd
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
func (m metaCmd) newUninstallCmd() *cobra.Command {
	c := &uninstallCmd{metaCmd: m}

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
		ValidArgs:             state.Keys(m.state.Deletions),
		RunE: func(cmd *cobra.Command, args []string) error {
			resources := m.state.Deletions
			if len(resources) == 0 {
				fmt.Println("No packages to uninstall")
				return nil
			}

			var tmp []state.Resource
			for _, arg := range args {
				resource, ok := state.Map(resources)[arg]
				if !ok {
					return fmt.Errorf("%s: no such package to be uninstalled", arg)
				}
				tmp = append(tmp, resource)
			}
			if len(tmp) > 0 {
				resources = tmp
			}

			yes, _ := m.askRunCommand(*c, state.Keys(resources))
			if !yes {
				fmt.Println("Cancelled")
				return nil
			}

			return c.run(resources)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return m.printForUpdate()
		},
	}

	return uninstallCmd
}

func (c *uninstallCmd) run(resources []state.Resource) error {
	var errs errors.Errors

	delete := func(paths ...string) error {
		var errs errors.Errors
		for _, path := range paths {
			errs.Append(os.RemoveAll(path))
		}
		return errs.ErrorOrNil()
	}

	for _, resource := range resources {
		err := delete(append(resource.Paths, resource.Home)...)
		if err != nil {
			errs.Append(err)
			continue
		}
		c.state.Remove(resource)
		fmt.Printf("deleted %s\n", resource.Home)
	}

	return errs.ErrorOrNil()
}
