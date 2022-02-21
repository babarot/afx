package cmd

import (
	"fmt"
	"log"

	"github.com/b4b4r07/afx/pkg/helpers/templates"
	"github.com/spf13/cobra"
)

type initCmd struct {
	meta

	path string
}

var (
	// initLong is long description of init command
	initLong = templates.LongDesc(``)

	// initExample is examples for init command
	initExample = templates.Examples(`
		# show a source file to start packages installed by afx
		afx init

		# enable plugins/commands in current shell
		source <(afx init)

		# automatically load configurations
		Bash:
		  echo 'source <(afx init)' ~/.bashrc
		Zsh:
		  echo 'source <(afx init)' ~/.zshrc
		Fish:
		  echo 'afx init | source' ~/.config/fish/config.fish
	`)
)

// newInitCmd creates a new init command
func newInitCmd() *cobra.Command {
	c := &initCmd{}

	initCmd := &cobra.Command{
		Use:                   "init",
		Short:                 "Initialize installed packages",
		Long:                  initLong,
		Example:               initExample,
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

	return initCmd
}

func (c *initCmd) run(args []string) error {
	for _, pkg := range c.Packages {
		// if has init config,
		// print pkg name, e.g. fmt.Printf("# %s\n", pkg.GetHome())
		// endif
		if err := pkg.Init(); err != nil {
			log.Printf("[ERROR] %s: failed to init pacakge: %v\n", pkg.GetName(), err)
			// do not return err to continue to load even if failed
			continue
		}
	}

	if c.path != "" {
		fmt.Printf("export PATH=$PATH:%s\n", c.path)
	}

	return nil
}
