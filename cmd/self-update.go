package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type selfUpdateCmd struct {
	meta
}

var (
	// selfUpdateLong is long description of self-update command
	selfUpdateLong = templates.LongDesc(``)

	// selfUpdateExample is examples for selfUpdate command
	selfUpdateExample = templates.Examples(`
		afx self-update
	`)
)

// newSelfUpdateCmd creates a new selfUpdate command
func newSelfUpdateCmd() *cobra.Command {
	c := &selfUpdateCmd{}

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
			if err := c.meta.init(args); err != nil {
				return err
			}
			return c.run(args)
		},
	}

	return selfUpdateCmd
}

func (c *selfUpdateCmd) run(args []string) error {
	switch Version {
	case "unset":
		fmt.Fprintf(os.Stderr, "%s\n\n  %s\n  %s\n\n",
			color.RedString("Failed to update to new version!"),
			"You need to get precompiled version from GitHub releases.",
			fmt.Sprintf("This version (%s/%s) is compiled from source code.",
				Version, runtime.Version()),
		)
		return errors.New("failed to run self-update")
	}

	latest, found, err := selfupdate.DetectLatest(Repository)
	if err != nil {
		return errors.Wrap(err, "error occurred while detecting version")
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
		Message: fmt.Sprintf("Do you want to update to %s? (current version: %s)",
			latest.Version(), Version),
	}, &yes); err != nil {
		return errors.Wrap(err, "cannot get answer from console")
	}
	if !yes {
		// do nothing
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return errors.New("could not locate executable path")
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, latest.AssetName, exe); err != nil {
		return errors.Wrap(err, "error occurred while updating binary")
	}

	color.New(color.Bold).Printf("Successfully updated to version %s", latest.Version())
	return nil
}
