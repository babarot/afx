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

			yes, _ := m.askRunCommand(*c, state.Keys(resources))
			if !yes {
				fmt.Println("Cancelled")
				return nil
			}

			pkgs := m.GetPackages(resources)
			m.env.AskWhen(map[string]bool{
				"GITHUB_TOKEN": config.HasGitHubReleaseBlock(pkgs),
			})

			return c.run(pkgs)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return m.printForUpdate()
		},
	}

	return checkCmd
}

type checkResult struct {
	Package config.Package
	Error   error
}

func (c *checkCmd) run(pkgs []config.Package) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	progress := config.NewProgress(pkgs)
	completion := make(chan config.Status)
	limit := make(chan struct{}, 16)
	results := make(chan checkResult)

	go func() {
		progress.Print(completion)
	}()

	log.Printf("[DEBUG] (check): start to run each pkg.Check()")
	eg := errgroup.Group{}
	for _, pkg := range pkgs {
		pkg := pkg
		eg.Go(func() error {
			limit <- struct{}{}
			defer func() { <-limit }()
			err := pkg.Check(ctx, completion)
			select {
			case results <- checkResult{Package: pkg, Error: err}:
				return nil
			case <-ctx.Done():
				return errors.Wrapf(ctx.Err(), "%s: cancelled checking", pkg.GetName())
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
		log.Printf("[ERROR] failed to check: %s", err)
		exit.Append(err)
	}

	defer func(err error) {
		if err != nil {
			c.env.Refresh()
		}
	}(exit.ErrorOrNil())

	return exit.ErrorOrNil()
}
