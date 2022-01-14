package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/b4b4r07/afx/pkg/logging"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/spf13/cobra"
)

var (
	rootLong = templates.LongDesc(`Package manager for everything`)
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
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                "pkg",
		Short:              "Package manager for everything",
		Long:               rootLong,
		SilenceErrors:      true,
		DisableSuggestions: false,
		Version:            fmt.Sprintf("%s (%s/%s)", Version, BuildTag, BuildSHA),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
	}

	// var (
	// 	version bool
	// )
	//
	// p := rootCmd.PersistentFlags()
	// p.BoolVar(&version, "version", false, "show version")
	//
	// if version {
	// 	fmt.Printf("%s (%s) / %s\n", Version, BuildTag, BuildSHA)
	// }

	rootCmd.AddCommand(newInstallCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newRemoveCmd())
	rootCmd.AddCommand(newGetCmd())

	return rootCmd
}

// Execute is
func Execute() error {
	logWriter, err := logging.LogOutput()
	if err != nil {
		return err
	}
	log.SetOutput(logWriter)

	log.Printf("[INFO] afx version: %s", Version)
	log.Printf("[INFO] Go runtime version: %s", runtime.Version())
	log.Printf("[INFO] Build tag/SHA: %s/%s", BuildTag, BuildSHA)
	log.Printf("[INFO] CLI args: %#v", os.Args)

	defer log.Printf("[DEBUG] root command execution finished")
	return newRootCmd().Execute()
}
