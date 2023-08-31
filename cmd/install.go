package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/babarot/afx/pkg/config"
	"github.com/babarot/afx/pkg/errors"
	"github.com/babarot/afx/pkg/helpers/templates"
	"github.com/babarot/afx/pkg/logging"
	"github.com/babarot/afx/pkg/state"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
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
				fmt.Println("Cancelled")
				return nil
			}

			pkgs := m.GetPackages(resources)
			m.env.AskWhen(map[string]bool{
				"GITHUB_TOKEN":      config.HasGitHubReleaseBlock(pkgs),
				"AFX_SUDO_PASSWORD": config.HasSudoInCommandBuildSteps(pkgs),
			})

			return c.run(pkgs)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return m.printForUpdate()
		},
	}

	return installCmd
}

type installResult struct {
	Package config.Package
	Error   error
}

func (c *installCmd) run(pkgs []config.Package) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	progress := config.NewProgress(pkgs)
	completion := make(chan config.Status)
	limit := make(chan struct{}, 16)
	results := make(chan installResult)

	go func() {
		progress.Print(completion)
	}()

	log.Printf("[DEBUG] (install): start to run each pkg.Install()")
	eg := errgroup.Group{}
	for _, pkg := range pkgs {
		pkg := pkg
		eg.Go(func() error {
			limit <- struct{}{}
			defer func() { <-limit }()
			err := pkg.Install(ctx, completion)
			switch err {
			case nil:
				c.state.Add(pkg)
			default:
				if !logging.IsSet() {
					log.Printf("[DEBUG] uninstall %q because installation failed", pkg.GetName())
					pkg.Uninstall(ctx)
				}
			}
			select {
			case results <- installResult{Package: pkg, Error: err}:
				return nil
			case <-ctx.Done():
				return errors.Wrapf(ctx.Err(), "%s: cancelled installation", pkg.GetName())
			}
		})
	}

	go func() {
		eg.Wait()
		close(results)
	}()

	var exit errors.Errors
	for result := range results {
		exit.Append(result.Error)
	}

	if err := eg.Wait(); err != nil {
		log.Printf("[ERROR] failed to install: %s", err)
		exit.Append(err)
	}

	defer func(err error) {
		if err != nil {
			c.env.Refresh()
		}
	}(exit.ErrorOrNil())

	return exit.ErrorOrNil()
}
