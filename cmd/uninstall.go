package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/b4b4r07/afx/pkg/errors"
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
	resources := c.State.Deletions
	if len(resources) == 0 {
		// TODO: improve message
		log.Printf("[INFO] No packages to uninstall")
		return nil
	}

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
