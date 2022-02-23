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
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type updateCmd struct {
	meta
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
func newUpdateCmd() *cobra.Command {
	c := &updateCmd{}

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
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.meta.init(args); err != nil {
				return err
			}

			pkgs := c.State.Changes
			if len(pkgs) == 0 {
				fmt.Println("No packages to update")
				return nil
			}

			// not update all packages. Instead just only update
			// given packages when not updated yet.
			var given []config.Package
			for _, arg := range args {
				pkg, err := c.getFromChanges(arg)
				if err != nil {
					// no hit in changes
					continue
				}
				given = append(given, pkg)
			}
			if len(given) > 0 {
				pkgs = given
			}

			yes, _ := c.askRunCommand(*c, getNameInPackages(pkgs))
			if !yes {
				fmt.Println("Cancelled")
				return nil
			}

			c.Env.AskWhen(map[string]bool{
				"GITHUB_TOKEN":      config.HasGitHubReleaseBlock(pkgs),
				"AFX_SUDO_PASSWORD": config.HasSudoInCommandBuildSteps(pkgs),
			})

			return c.run(pkgs)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return c.meta.printForUpdate()
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
			err := pkg.Install(ctx, completion)
			switch err {
			case nil:
				c.State.Update(pkg)
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
			c.Env.Refresh()
		}
	}(exit.ErrorOrNil())

	return exit.ErrorOrNil()
}

func (c *updateCmd) getFromChanges(name string) (config.Package, error) {
	pkgs := c.State.Changes

	for _, pkg := range pkgs {
		if pkg.GetName() == name {
			return pkg, nil
		}
	}

	return nil, errors.New("not found")
}
