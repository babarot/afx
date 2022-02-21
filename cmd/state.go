package cmd

import (
	"fmt"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/helpers/templates"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type stateCmd struct {
	meta
}

var (
	// stateLong is long description of state command
	stateLong = templates.LongDesc(``)

	// stateExample is examples for state command
	stateExample = templates.Examples(``)
)

// newStateCmd creates a new state command
func newStateCmd() *cobra.Command {
	stateCmd := &cobra.Command{
		Use:                   "state [list|refresh]",
		Short:                 "Advanced state management",
		Long:                  stateLong,
		Example:               stateExample,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MaximumNArgs(1),
		Hidden:                true,
	}

	stateListCmd := newStateListCmd()
	stateRefreshCmd := newStateRefreshCmd()
	stateRefreshCmd.Flags().BoolP("force", "", false, "Force update")

	stateCmd.AddCommand(
		stateListCmd,
		stateRefreshCmd,
	)

	return stateCmd
}

func newStateListCmd() *cobra.Command {
	c := &stateCmd{}

	return &cobra.Command{
		Use:                   "list",
		Short:                 "List your state items",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MaximumNArgs(0),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.meta.init(args)
		},
		Run: func(cmd *cobra.Command, args []string) {
			for _, pkg := range c.State.NoChanges {
				fmt.Println(pkg.GetName())
			}
		},
	}
}

func newStateRefreshCmd() *cobra.Command {
	c := &stateCmd{}

	return &cobra.Command{
		Use:                   "refresh",
		Short:                 "Refresh your state file",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MaximumNArgs(0),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.meta.init(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			if force {
				return c.State.New()
			}

			if err := c.State.Refresh(); err != nil {
				return errors.Wrap(err, "failed to refresh state")
			}
			fmt.Println(color.WhiteString("Successfully refreshed"))
			return nil
		},
	}
}
