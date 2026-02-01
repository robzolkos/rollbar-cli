package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for rollbar.

To load completions:

Bash:
  $ source <(rollbar completion bash)
  # Or add to ~/.bashrc:
  $ rollbar completion bash > /etc/bash_completion.d/rollbar

Zsh:
  $ source <(rollbar completion zsh)
  # Or add to ~/.zshrc:
  $ rollbar completion zsh > "${fpath[1]}/_rollbar"

Fish:
  $ rollbar completion fish | source
  # Or add to ~/.config/fish/completions:
  $ rollbar completion fish > ~/.config/fish/completions/rollbar.fish

PowerShell:
  PS> rollbar completion powershell | Out-String | Invoke-Expression
  # Or add to your PowerShell profile`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				_ = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				_ = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				_ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}

	return cmd
}
