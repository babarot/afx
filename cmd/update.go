package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/babarot/afx/internal/helpers/templates"
	manager "github.com/babarot/afx/internal/manager"
	"github.com/babarot/afx/internal/runner"
	"github.com/babarot/afx/internal/state"
)

type updateCmd struct {
	metaCmd
}

var (
	// updateLong is long description of fmt command
	updateLong = templates.LongDesc(``)

	// updateExample is examples for fmt command
	updateExample = templates.Examples(`
		afx update [args...]

		By default, it tries to update packages only if something are
		changed in config file.
		If any args are given, it tries to update only them.
	`)
)

// newUpdateCmd creates a new fmt command
func (m metaCmd) newUpdateCmd() *cobra.Command {
	c := &updateCmd{metaCmd: m}

	updateCmd := &cobra.Command{
		Use:                   "update",
		Short:                 "Update installed package if version etc is changed",
		Long:                  updateLong,
		Example:               updateExample,
		Aliases:               []string{"u"},
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MinimumNArgs(0),
		ValidArgs:             state.Keys(m.state.Changes),
		RunE: func(cmd *cobra.Command, args []string) error {
			resources := m.state.Changes
			if len(resources) == 0 {
				fmt.Println("No packages to update")
				return nil
			}

			var tmp []state.Resource
			for _, arg := range args {
				resource, ok := state.Map(resources)[arg]
				if !ok {
					return fmt.Errorf("%s: no such package in config", arg)
				}
				tmp = append(tmp, resource)
			}
			if len(tmp) > 0 {
				resources = tmp
			}

			yes, _ := m.askRunCommand(*c, state.Keys(resources))
			if !yes {
				fmt.Println("Canceled")
				return nil
			}

			pkgs := m.GetPackages(resources)
			m.env.AskWhen(map[string]bool{
				"GITHUB_TOKEN":      manager.HasGitHubReleaseBlock(pkgs),
				"AFX_SUDO_PASSWORD": manager.HasSudoInCommandBuildSteps(pkgs),
			})

			return c.run(pkgs)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return m.printForUpdate()
		},
	}

	return updateCmd
}

func (c *updateCmd) run(pkgs []manager.Package) error {
	log.Printf("[DEBUG] (update): start to run each pkg.Install()")

	runnerPkgs := make([]runner.Package, len(pkgs))
	for i, p := range pkgs {
		runnerPkgs[i] = p
	}

	err := runner.Execute(runnerPkgs, func(p runner.Package) runner.TaskFunc {
		pkg, _ := p.(manager.Package)
		return func(ctx context.Context, completion chan<- runner.Status) error {
			home := pkg.GetHome()
			backup := home + ".bak"

			// Backup existing installation before updating
			if _, err := os.Stat(home); err == nil {
				_ = os.RemoveAll(backup) // remove stale backup if any
				if err := os.Rename(home, backup); err != nil {
					log.Printf("[WARN] %s: failed to backup, falling back to remove: %v", pkg.GetName(), err)
					_ = os.RemoveAll(home)
				}
			}

			err := pkg.Install(ctx, completion)
			switch err {
			case nil:
				c.state.Update(pkg)
				_ = os.RemoveAll(backup) // clean up backup on success
			default:
				// Restore backup on failure
				if _, statErr := os.Stat(backup); statErr == nil {
					log.Printf("[DEBUG] restoring %q from backup after failed update", pkg.GetName())
					_ = os.RemoveAll(home)
					_ = os.Rename(backup, home)
				} else {
					log.Printf("[DEBUG] uninstall %q because updating failed (no backup)", pkg.GetName())
					_ = pkg.Uninstall(ctx)
				}
			}
			return err
		}
	})

	if err != nil {
		_ = c.env.Refresh()
	}
	return err
}
