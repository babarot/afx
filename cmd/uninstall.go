package cmd

import (
	"os"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/helpers/spin"
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
	pkg, err := c.Prompt()
	if err != nil {
		return err
	}

	s := spin.New("Removing "+pkg.GetSlug()+" %s", spin.WithDoneMessage("Uninstalled\n"))
	s.Start()
	defer s.Stop()

	var errs errors.Errors

	switch {
	case pkg.HasPluginBlock():
		// TODO: think what to do in this type (plugin)
	case pkg.HasCommandBlock():
		command := pkg.GetCommandBlock()
		links, err := command.GetLink(pkg)
		if err != nil {
			return err
		}
		for _, link := range links {
			errs.Append(os.Remove(link.From))
			errs.Append(os.Remove(link.To))
		}
	}

	errs.Append(os.RemoveAll(pkg.GetHome()))
	return errs.ErrorOrNil()
}
