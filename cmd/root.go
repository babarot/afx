package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/helpers/templates"
	"github.com/b4b4r07/afx/pkg/logging"
	"github.com/b4b4r07/afx/pkg/update"
	"github.com/spf13/cobra"
)

var Repository string = "b4b4r07/afx"

var (
	rootLong = templates.LongDesc(`Package manager for CLI`)
)

var (
	// Version is the version number
	Version = "unset"

	// BuildTag set during build to git tag, if any
	BuildTag = "unset"

	// BuildSHA is the git sha set during build
	BuildSHA = "unset"
)

// newRootCmd returns the root command
func newRootCmd(m metaCmd) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                "afx",
		Short:              "Package manager for CLI",
		Long:               rootLong,
		SilenceErrors:      true,
		DisableSuggestions: false,
		Version:            fmt.Sprintf("%s (%s/%s)", Version, BuildTag, BuildSHA),
		PreRun: func(cmd *cobra.Command, args []string) {
			uriCh := make(chan *update.ReleaseInfo)
			go func() {
				log.Printf("[DEBUG] (goroutine): checking new updates...")
				release, err := checkForUpdate(Version)
				if err != nil {
					log.Printf("[ERROR] (goroutine): cannot check for new updates: %s", err)
				}
				uriCh <- release
			}()

			if cmd.Runnable() {
				cmd.Help()
			}

			printForUpdate(uriCh)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Just define this function to prevent c.Runnable() from becoming false.
			// if c.Runnable() is true, just c.Help() is called and then stopped.
			return nil
		},
	}

	rootCmd.AddCommand(
		m.newInitCmd(),
		m.newInstallCmd(),
		m.newUninstallCmd(),
		m.newUpdateCmd(),
		m.newSelfUpdateCmd(),
		m.newShowCmd(),
		m.newCompletionCmd(),
		m.newStateCmd(),
	)

	return rootCmd
}

func Execute() error {
	logWriter, err := logging.LogOutput()
	if err != nil {
		return errors.Wrap(err, "%s: failed to set logger")
	}
	log.SetOutput(logWriter)

	log.Printf("[INFO] afx version: %s", Version)
	log.Printf("[INFO] Go runtime version: %s", runtime.Version())
	log.Printf("[INFO] Build tag/SHA: %s/%s", BuildTag, BuildSHA)
	log.Printf("[INFO] CLI args: %#v", os.Args)

	meta := metaCmd{}
	if err := meta.init(); err != nil {
		return errors.Wrap(err, "failed to initialize afx")
	}

	defer log.Printf("[INFO] root command execution finished")
	return newRootCmd(meta).Execute()
}
