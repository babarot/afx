package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/b4b4r07/afx/pkg/config"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/helpers/templates"
	"github.com/b4b4r07/afx/pkg/state"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
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

	return updateCmd
}

type updateResult struct {
	Package config.Package
	Error   error
}

func (c *updateCmd) run(pkgs []config.Package) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	progress := config.NewProgress(pkgs)
	completion := make(chan config.Status)
	limit := make(chan struct{}, 16)
	results := make(chan updateResult)

	go func() {
		progress.Print(completion)
	}()

	log.Printf("[DEBUG] (update): start to run each pkg.Install()")
	eg := errgroup.Group{}
	for _, pkg := range pkgs {
		pkg := pkg
		eg.Go(func() error {
			limit <- struct{}{}
			defer func() { <-limit }()
			os.RemoveAll(pkg.GetHome()) // delete before updating
			err := pkg.Install(ctx, completion)
			switch err {
			case nil:
				c.state.Update(pkg)
			default:
				log.Printf("[DEBUG] uninstall %q because updating failed", pkg.GetName())
				pkg.Uninstall(ctx)
			}
			select {
			case results <- updateResult{Package: pkg, Error: err}:
				return nil
			case <-ctx.Done():
				return errors.Wrapf(ctx.Err(), "%s: cancelled updating", pkg.GetName())
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
		log.Printf("[ERROR] failed to update: %s", err)
		exit.Append(err)
	}

	defer func(err error) {
		if err != nil {
			c.env.Refresh()
		}
	}(exit.ErrorOrNil())

	return exit.ErrorOrNil()
}
