package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/b4b4r07/afx/pkg/config"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/templates"
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
			if c.parseErr != nil {
				return c.parseErr
			}
			c.Env.Ask(
				"AFX_SUDO_PASSWORD",
				"GITHUB_TOKEN",
			)
			return c.run(args)
		},
	}

	return updateCmd
}

type updateResult struct {
	Package config.Package
	Error   error
}

func (c *updateCmd) run(args []string) error {
	eg := errgroup.Group{}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	pkgs := c.State.Changes
	if len(pkgs) == 0 {
		// TODO: improve message
		log.Printf("[INFO] No packages to update")
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

	progress := config.NewProgress(pkgs)
	completion := make(chan config.Status)

	go func() {
		progress.Print(completion)
	}()

	log.Printf("[DEBUG] (update): start to run each pkg.Install()")
	results := make(chan updateResult)
	for _, pkg := range pkgs {
		pkg := pkg
		eg.Go(func() error {
			err := pkg.Install(ctx, completion)
			switch err {
			case nil:
				c.State.Update(pkg)
			}
			select {
			case results <- updateResult{Package: pkg, Error: err}:
				return nil
			case <-ctx.Done():
				return ctx.Err()
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
		log.Printf("[ERROR] failed to update: %s\n", err)
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
