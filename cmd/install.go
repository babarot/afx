package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/babarot/afx/internal/helpers/templates"
	"github.com/babarot/afx/internal/logging"
	afxpkg "github.com/babarot/afx/internal/pkg"
	"github.com/babarot/afx/internal/runner"
	"github.com/babarot/afx/internal/state"
)

type installCmd struct {
	metaCmd
}

var (
	// installLong is long description of fmt command
	installLong = templates.LongDesc(``)

	// installExample is examples for fmt command
	installExample = templates.Examples(`
		afx install [args...]

		By default, it tries to install all packages which are newly
		added to config file.
		If any args are given, it tries to install only them.
	`)
)

// newInstallCmd creates a new fmt command
func (m metaCmd) newInstallCmd() *cobra.Command {
	c := &installCmd{metaCmd: m}

	installCmd := &cobra.Command{
		Use:                   "install",
		Short:                 "Resume installation from paused part (idempotency)",
		Long:                  installLong,
		Example:               installExample,
		Aliases:               []string{"i"},
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MinimumNArgs(0),
		ValidArgs:             state.Keys(m.state.Additions),
		RunE: func(cmd *cobra.Command, args []string) error {
			resources := m.state.Additions
			if len(resources) == 0 {
				fmt.Println("No packages to install")
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
				"GITHUB_TOKEN":      afxpkg.HasGitHubReleaseBlock(pkgs),
				"AFX_SUDO_PASSWORD": afxpkg.HasSudoInCommandBuildSteps(pkgs),
			})

			return c.run(pkgs)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return m.printForUpdate()
		},
	}

	return installCmd
}

func (c *installCmd) run(pkgs []afxpkg.Package) error {
	log.Printf("[DEBUG] (install): start to run each pkg.Install()")

	runnerPkgs := make([]runner.Package, len(pkgs))
	for i, p := range pkgs {
		runnerPkgs[i] = p
	}

	err := runner.Execute(runnerPkgs, func(p runner.Package) runner.TaskFunc {
		pkg, _ := p.(afxpkg.Package)
		return func(ctx context.Context, completion chan<- runner.Status) error {
			err := pkg.Install(ctx, completion)
			switch err {
			case nil:
				c.state.Add(pkg)
			default:
				if !logging.IsSet() {
					log.Printf("[DEBUG] uninstall %q because installation failed", pkg.GetName())
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
