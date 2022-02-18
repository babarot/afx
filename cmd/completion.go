package cmd

import (
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/spf13/cobra"
)

type completionCmd struct {
	meta
}

var (
	// completionLong is long description of completion command
	completionLong = heredoc.Doc(`
		To load completions:

		Bash:

		  $ source <(afx completion bash)

		  # To load completions for each session, execute once:
		  # Linux:
		  $ afx completion bash > /etc/bash_completion.d/afx
		  # macOS:
		  $ afx completion bash > /usr/local/etc/bash_completion.d/afx

		Zsh:

		  # If shell completion is not already enabled in your environment,
		  # you will need to enable it.  You can execute the following once:

		  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

		  # To load completions for each session, execute once:
		  $ afx completion zsh > "${fpath[1]}/_afx"

		  # You will need to start a new shell for this setup to take effect.

		fish:

		  $ afx completion fish | source

		  # To load completions for each session, execute once:
		  $ afx completion fish > ~/.config/fish/completions/afx.fish
		`)

	// completionExample is examples for completion command
	completionExample = templates.Examples(`
		afx completion bash
		afx completion zsh
		afx completion fish
	`)
)

// newCompletionCmd creates a new completion command
func newCompletionCmd() *cobra.Command {
	c := &completionCmd{}

	completionCmd := &cobra.Command{
		Use:                   "completion [bash|zsh|fish]",
		Short:                 "Generate completion script",
		Long:                  completionLong,
		Example:               completionExample,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.meta.init(args); err != nil {
				return err
			}
			switch args[0] {
			case "bash":
				newRootCmd().GenBashCompletion(os.Stdout)
			case "zsh":
				newRootCmd().GenZshCompletion(os.Stdout)
			case "fish":
				newRootCmd().GenFishCompletion(os.Stdout, true)
			}
			return nil
		},
	}

	return completionCmd
}
