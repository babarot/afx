package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b4b4r07/afx/internal/diags"
	"github.com/b4b4r07/afx/pkg/helpers/templates"
	"github.com/b4b4r07/afx/pkg/state"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type stateCmd struct {
	metaCmd

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
func (m metaCmd) newStateCmd() *cobra.Command {
	c := &stateCmd{metaCmd: m}

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

	stateCmd.AddCommand(
		c.newStateListCmd(),
		c.newStateRefreshCmd(),
		c.newStateRemoveCmd(),
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
		Args:                  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			resources, err := c.state.List()
			if err != nil {
				return err
			}
			for _, resource := range resources {
				fmt.Println(resource.Name)
			}
			return nil
		},
	}
}

func (c stateCmd) newStateRefreshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "refresh",
		Short:                 "Refresh your state file",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if c.opt.force {
				return c.state.New()
			}
			if err := c.state.Refresh(); err != nil {
				return diags.Wrap(err, "failed to refresh state")
			}
			fmt.Println(color.WhiteString("Successfully refreshed"))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&c.opt.force, "force", "", false, "force update")
	return cmd
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
		ValidArgs:             state.Keys(c.state.NoChanges),
		RunE: func(cmd *cobra.Command, args []string) error {
			var resources []state.Resource
			switch len(cmd.Flags().Args()) {
			case 0:
				rs, err := c.state.List()
				if err != nil {
					return diags.Wrap(err, "failed to list state items")
				}
				var items []string
				for _, r := range rs {
					items = append(items, r.Name)
				}
				var selected string
				if err := survey.AskOne(&survey.Select{
					Message: "Choose a package:",
					Options: items,
				}, &selected); err != nil {
					return diags.Wrap(err, "failed to get input from console")
				}
				resource, err := c.state.Get(selected)
				if err != nil {
					return diags.Wrapf(err, "%s: failed to get state file", selected)
				}
				resources = append(resources, resource)
			default:
				for _, arg := range cmd.Flags().Args() {
					resource, err := c.state.Get(arg)
					if err != nil {
						return diags.Wrapf(err, "%s: failed to get state file", arg)
					}
					resources = append(resources, resource)
				}
			}
			for _, resource := range resources {
				c.state.Remove(resource)
			}
			return nil
		},
	}
}
