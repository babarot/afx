package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/babarot/afx/pkg/helpers/templates"
)

var (
	// completionLong is long description of completion command
	completionLong = templates.LongDesc(``)

	// completionExample is examples for completion command
	completionExample = templates.Raw(`
		To load completions:

		Bash:
		  $ source <(afx completion bash)

		  # To load completions for each session, execute once:
		  # Linux:
		  $ afx completion bash > /etc/bash_completion.d/afx
		  # macOS:
		  $ afx completion bash > /usr/local/etc/bash_completion.d/afx

		Zsh:
		  $ source <(afx completion zsh)

		  # If shell completion is not already enabled in your environment,
		  # you will need to enable it.  You can execute the following once:

		  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

		  # To load completions for each session, execute once:
		  $ afx completion zsh > "${fpath[1]}/_afx"

		  # You will need to start a new shell for this setup to take effect.

		Fish:
		  $ afx completion fish | source

		  # To load completions for each session, execute once:
		  $ afx completion fish > ~/.config/fish/completions/afx.fish
	`)
)

// newCompletionCmd creates a new completion command
func (m metaCmd) newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "completion [bash|zsh|fish]",
		Short:                 "Generate completion script",
		Long:                  completionLong,
		Example:               completionExample,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				_ = newRootCmd(m).GenBashCompletion(os.Stdout)
			case "zsh":
				_ = newRootCmd(m).GenZshCompletion(os.Stdout)
			case "fish":
				_ = newRootCmd(m).GenFishCompletion(os.Stdout, true)
			}
			return nil
		},
	}
}
