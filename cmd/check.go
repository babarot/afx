package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/babarot/afx/internal/helpers/templates"
	manager "github.com/babarot/afx/internal/manager"
	"github.com/babarot/afx/internal/runner"
	"github.com/babarot/afx/internal/state"
)

type checkCmd struct {
	metaCmd
}

var (
	// checkLong is long description of fmt command
	checkLong = templates.LongDesc(``)

	// checkExample is examples for fmt command
	checkExample = templates.Examples(`
		afx check [args...]

		By default, it tries to check packages if new version is
		available or not.
		If any args are given, it tries to check only them.
	`)
)

// newCheckCmd creates a new fmt command
func (m metaCmd) newCheckCmd() *cobra.Command {
	c := &checkCmd{metaCmd: m}

	checkCmd := &cobra.Command{
		Use:                   "check",
		Short:                 "Check new updates on each package",
		Long:                  checkLong,
		Example:               checkExample,
		Aliases:               []string{"c"},
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MinimumNArgs(0),
		ValidArgs:             state.Keys(m.state.NoChanges),
		RunE: func(cmd *cobra.Command, args []string) error {
			resources := m.state.NoChanges
			if len(resources) == 0 {
				fmt.Println("No packages to check")
				return nil
			}

			var tmp []state.Resource
			for _, arg := range args {
				resource, ok := state.Map(resources)[arg]
				if !ok {
					return fmt.Errorf("%s: not installed yet", arg)
				}
				tmp = append(tmp, resource)
			}
			if len(tmp) > 0 {
				resources = tmp
			}

			yes, err := m.askRunCommand(*c, state.Keys(resources))
			if err != nil {
				return fmt.Errorf("failed to confirm: %w", err)
			}
			if !yes {
				fmt.Println("Canceled")
				return nil
			}

			pkgs := m.GetPackages(resources)
			m.env.AskWhen(map[string]bool{
				"GITHUB_TOKEN": manager.HasGitHubReleaseBlock(pkgs),
			})

			return c.run(pkgs)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return m.printForUpdate()
		},
	}

	return checkCmd
}

func (c *checkCmd) run(pkgs []manager.Package) error {
	log.Printf("[DEBUG] (check): start to run each pkg.Check()")

	runnerPkgs := make([]runner.Package, len(pkgs))
	for i, p := range pkgs {
		runnerPkgs[i] = p
	}

	err := runner.Execute(runnerPkgs, func(p runner.Package) runner.TaskFunc {
		pkg, _ := p.(manager.Package)
		return func(ctx context.Context, completion chan<- runner.Status) error {
			return pkg.Check(ctx, completion)
		}
	})

	if err != nil {
		_ = c.env.Refresh()
	}
	return err
}
