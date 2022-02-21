package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type stateCmd struct {
	meta

	opt stateOpt
}

type stateOpt struct {
	force bool
}

var (
	// stateLong is long description of state command
	stateLong = templates.LongDesc(``)

	// stateExample is examples for state command
	stateExample = templates.Examples(``)
)

// newStateCmd creates a new state command
func newStateCmd() *cobra.Command {
	c := &stateCmd{
		opt: stateOpt{},
	}

	stateCmd := &cobra.Command{
		Use:                   "state [list|refresh|remove]",
		Short:                 "Advanced state management",
		Long:                  stateLong,
		Example:               stateExample,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MaximumNArgs(1),
		Hidden:                true,
	}

	stateListCmd := c.newStateListCmd()
	stateRefreshCmd := c.newStateRefreshCmd()
	stateRefreshCmd.Flags().BoolVarP(&c.opt.force, "force", "", false, "force update")
	stateRemoveCmd := c.newStateRemoveCmd()

	stateCmd.AddCommand(
		stateListCmd,
		stateRefreshCmd,
		stateRemoveCmd,
	)

	return stateCmd
}

func (c stateCmd) newStateListCmd() *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := c.State.List()
			if err != nil {
				return err
			}
			for _, item := range items {
				fmt.Println(item)
			}
			return nil
		},
	}
}

func (c stateCmd) newStateRefreshCmd() *cobra.Command {
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
			if c.opt.force {
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

func (c stateCmd) newStateRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "remove",
		Short:                 "Remove selected packages from state file",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Aliases:               []string{"rm"},
		Args:                  cobra.MinimumNArgs(0),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.meta.init(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var selected []string
			switch len(cmd.Flags().Args()) {
			case 0:
				resources, err := c.State.List()
				if err != nil {
					return errors.Wrap(err, "failed to list state items")
				}
				prompt := &survey.MultiSelect{
					Message:  "Choose a package:",
					Options:  resources,
					PageSize: 10,
				}
				survey.AskOne(prompt, &selected)
			default:
				// TODO: check valid or invalid
				selected = cmd.Flags().Args()
			}
			for _, resource := range selected {
				c.State.Remove(resource)
			}
			return nil
		},
	}
}
