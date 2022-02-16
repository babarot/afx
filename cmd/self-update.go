package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

type selfUpdateCmd struct {
	meta

	path string
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
	if Version == "unset" {
		fmt.Printf("Failed to update to new version\n")
		return fmt.Errorf("this version (%s/%s) is compiled from source code",
			Version, runtime.Version())
	}

	latest, found, err := selfupdate.DetectLatest("b4b4r07/afx")
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %v", err)
	}

	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from GitHub repository",
			runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(Version) {
		fmt.Printf("Current version (%s) is the latest\n", Version)
		return nil
	}

	fmt.Printf("Do you want to update to %s? (current: %s) (y/n) ", latest.Version(), Version)
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || (input != "y\n" && input != "n\n") {
		fmt.Println("Invalid input")
		return err
	}
	if input == "n\n" {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return errors.New("could not locate executable path")
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}

	fmt.Printf("Successfully updated to version %s\n", latest.Version())
	return nil
}
