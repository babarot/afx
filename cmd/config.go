package cmd

import (
	"context"

	"github.com/b4b4r07/afx/pkg/helpers/shell"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/spf13/cobra"
)

type configCmd struct {
	meta
}

var (
	// configLong is long description of config command
	configLong = templates.LongDesc(``)

	// configExample is examples for config command
	configExample = templates.Examples(`
		# Normal
		pkg config
	`)
)

// newConfigCmd creates a new config command
func newConfigCmd() *cobra.Command {
	c := &configCmd{}

	configCmd := &cobra.Command{
		Use:                   "config",
		Short:                 "Configure HCL files",
		Long:                  configLong,
		Example:               configExample,
		Aliases:               []string{"cfg", "configure"},
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.meta.init(args); err != nil {
				return err
			}
			if c.parseErr != nil {
				return c.parseErr
			}
			return c.run(args)
		},
	}

	return configCmd
}

func (c *configCmd) run(args []string) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	vim := shell.New("vim", path)
	return vim.Run(context.Background())
}
