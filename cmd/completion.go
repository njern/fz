package cmd

import "github.com/spf13/cobra"

var completionCmd = &cobra.Command{
	Use:   "completion <shell>",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for fz.

To load completions:

Bash:
  $ source <(fz completion bash)

Zsh:
  $ source <(fz completion zsh)
  # To load completions for each session, add to your .zshrc:
  # eval "$(fz completion zsh)"

Fish:
  $ fz completion fish | source
  # To load completions for each session:
  # fz completion fish > ~/.config/fish/completions/fz.fish

PowerShell:
  PS> fz completion powershell | Out-String | Invoke-Expression`,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(cmd.OutOrStdout(), true)
		case "zsh":
			return rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		default:
			return cmd.Usage()
		}
	},
}

func init() {
	withConfigMode(completionCmd, configLoadNone)
	rootCmd.AddCommand(completionCmd)
}
