package cmd

import (
	"os"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/helpers/spin"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/spf13/cobra"
)

type removeCmd struct {
	meta
}

var (
	// removeLong is long description of remove command
	removeLong = templates.LongDesc(``)

	// removeExample is examples for remove command
	removeExample = templates.Examples(`
		# Normal
		afx remove
	`)
)

// newRemoveCmd creates a new remove command
func newRemoveCmd() *cobra.Command {
	c := &removeCmd{}

	removeCmd := &cobra.Command{
		Use:                   "remove",
		Short:                 "Remove installed packages",
		Long:                  removeLong,
		Example:               removeExample,
		Aliases:               []string{"rm"},
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

	return removeCmd
}

func (c *removeCmd) run(args []string) error {
	pkg, err := c.Prompt()
	if err != nil {
		return err
	}

	s := spin.New("Removing "+pkg.GetSlug()+" %s", spin.WithDoneMessage("Removed\n"))
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
