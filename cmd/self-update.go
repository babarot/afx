package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b4b4r07/afx/internal/diags"
	"github.com/b4b4r07/afx/pkg/github"
	"github.com/b4b4r07/afx/pkg/helpers/templates"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type selfUpdateCmd struct {
	metaCmd

	opt selfUpdateOpt

	annotation map[string]string
}

type selfUpdateOpt struct {
	tag bool
}

var (
	// selfUpdateLong is long description of self-update command
	selfUpdateLong = templates.LongDesc(`
		self-update requires afx is pre-compiled one to upgrade.

		Those who built afx by go install etc cannot use this feature.
		(afx --version returns unset/unset)
	`)

	// selfUpdateExample is examples for selfUpdate command
	selfUpdateExample = templates.Examples(`
		# upgrade afx to latest version
		$ afx self-update
	`)
)

// newSelfUpdateCmd creates a new selfUpdate command
func (m metaCmd) newSelfUpdateCmd() *cobra.Command {
	info := color.New(color.FgGreen).SprintFunc()
	c := &selfUpdateCmd{
		metaCmd: m,
		annotation: map[string]string{
			"0.1.11": info(`Run "afx state refresh --force" at first!`),
		},
	}

	selfUpdateCmd := &cobra.Command{
		Use:                   "self-update",
		Short:                 "Update afx itself to latest version",
		Long:                  selfUpdateLong,
		Example:               selfUpdateExample,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			m.env.AskWhen(map[string]bool{
				"GITHUB_TOKEN": true,
			})
			return c.run(args)
		},
	}

	return selfUpdateCmd
}

func (c *selfUpdateCmd) run(args []string) error {
	switch Version {
	case "unset":
		fmt.Fprintf(os.Stderr, "%s\n\n  %s\n  %s\n\n",
			"Failed to update to new version!",
			"You need to get precompiled version from GitHub releases.",
			fmt.Sprintf("This version (%s/%s) is compiled from source code.",
				Version, runtime.Version()),
		)
		return diags.New("failed to run self-update")
	}

	latest, found, err := selfupdate.DetectLatest(Repository)
	if err != nil {
		return diags.Wrap(err, "error occurred while detecting version")
	}

	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from GitHub repository",
			runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(Version) {
		fmt.Printf("Current version (%s) is the latest\n", Version)
		return nil
	}

	yes := false
	if err := survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Do you update to %s? (current version: %s)",
			latest.Version(), Version),
	}, &yes); err != nil {
		return diags.Wrap(err, "cannot get answer from console")
	}
	if !yes {
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	release, err := github.NewRelease(ctx, "b4b4r07", "afx", "v"+latest.Version(), github.WithVerbose())
	if err != nil {
		return err
	}

	asset, err := release.Download(ctx)
	if err != nil {
		return err
	}

	if err := release.Unarchive(asset); err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return diags.New("could not locate executable path")
	}

	if err := release.Install(exe); err != nil {
		return err
	}

	color.New(color.FgWhite).Printf("Successfully updated to version %s\n", latest.Version())
	return nil
}
