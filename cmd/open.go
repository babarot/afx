package cmd

import (
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

type openCmd struct {
	meta
}

var (
	// openLong is long description of open command
	openLong = templates.LongDesc(``)

	// openExample is examples for open command
	openExample = templates.Examples(`
		# Normal
		pkg open
	`)
)

// newOpenCmd creates a new open command
func newOpenCmd() *cobra.Command {
	c := &openCmd{}

	openCmd := &cobra.Command{
		Use:                   "open",
		Short:                 "Open a page related to the package",
		Long:                  openLong,
		Example:               openExample,
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

	return openCmd
}

func (c *openCmd) run(args []string) error {
	pkg, err := c.Prompt()
	if err != nil {
		return err
	}
	return browser.OpenURL(pkg.GetURL())
}
