package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/njern/fz/internal/api"
	"github.com/njern/fz/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Sentinel errors for programmatic checking.
var (
	ErrNotAuthenticated = errors.New("not authenticated; run `fz auth login` first")
	ErrNoAccount        = errors.New("no account specified; use --account or run `fz auth login`")
	ErrInvalidToken     = errors.New("token is invalid or expired")
	ErrConfirmation     = errors.New("confirmation required; re-run with --yes")
)

var (
	cfgOverrideAccount string
	cfg                *config.Config
	cfgLoadErr         error
	cliVersion         string
	confirmYes         bool
)

type configLoadMode string

const (
	configAnnotationKey                 = "fz.config_mode"
	configLoadRequired   configLoadMode = "required"
	configLoadRepairable configLoadMode = "repairable"
	configLoadNone       configLoadMode = "none"
)

var rootCmd = &cobra.Command{
	Use:   "fz <command> <subcommand> [flags]",
	Short: "Work seamlessly with Fizzy from the command line",
	Long:  "fz is a CLI tool for interacting with the Fizzy project management API.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error

		cfg, cfgLoadErr, err = loadConfigForCommand(cmd)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgOverrideAccount, "account", "a", "", "Override the active account slug")
	rootCmd.PersistentFlags().BoolVarP(&confirmYes, "yes", "y", false, "Skip confirmation prompts")
}

func withConfigMode(cmd *cobra.Command, mode configLoadMode) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}

	cmd.Annotations[configAnnotationKey] = string(mode)
}

func configModeForCommand(cmd *cobra.Command) configLoadMode {
	for current := cmd; current != nil; current = current.Parent() {
		if raw, ok := current.Annotations[configAnnotationKey]; ok {
			return configLoadMode(raw)
		}
	}

	return configLoadRequired
}

func loadConfigForCommand(cmd *cobra.Command) (*config.Config, error, error) {
	mode := configModeForCommand(cmd)

	switch mode {
	case configLoadNone:
		cfg, err := config.New()
		return cfg, nil, err
	case configLoadRepairable:
		cfg, err := config.Load()
		if err == nil {
			return cfg, nil, nil
		}

		if errors.Is(err, config.ErrParse) {
			cfg, newErr := config.New()
			return cfg, err, newErr
		}

		return nil, nil, err
	default:
		cfg, err := config.Load()
		return cfg, nil, err
	}
}

// Execute runs the root command.
func Execute(version string) {
	cliVersion = version
	rootCmd.Version = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// newClient creates an authenticated API client from the loaded config.
func newClient() (*api.Client, error) {
	if !cfg.Authenticated() {
		return nil, ErrNotAuthenticated
	}

	return api.NewClient(cfg.Host, cfg.Token, cliVersion), nil
}

// accountSlug returns the active account slug, respecting the --account override.
func accountSlug() (string, error) {
	slug, err := cfg.AccountSlug(cfgOverrideAccount)
	if err != nil {
		return "", ErrNoAccount
	}

	return slug, nil
}

// accountClient creates an authenticated API client and resolves the active account slug.
func accountClient() (*api.Client, string, error) {
	client, err := newClient()
	if err != nil {
		return nil, "", err
	}

	slug, err := accountSlug()
	if err != nil {
		return nil, "", err
	}

	return client, slug, nil
}

// openInBrowser opens the given URL in the user's default browser.
func openInBrowser(url string) error {
	var (
		cmd  string
		args []string
	)

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default: // linux, freebsd, etc.
		cmd = "xdg-open"
	}

	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}

// confirmAction prompts the user for confirmation on interactive stdin.
// It returns ErrConfirmation when stdin is non-interactive and --yes was not passed.
func confirmAction(cmd *cobra.Command, prompt string) (bool, error) {
	if confirmYes {
		return true, nil
	}

	if !stdinIsInteractive(cmd) {
		return false, ErrConfirmation
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s [y/N]: ", prompt)
	reader := bufio.NewReader(cmd.InOrStdin())

	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("reading confirmation: %w", err)
	}

	answer = strings.TrimSpace(strings.ToLower(answer))

	return answer == "y" || answer == "yes", nil
}

func stdinIsInteractive(cmd *cobra.Command) bool {
	in, ok := cmd.InOrStdin().(*os.File)
	return ok && term.IsTerminal(int(in.Fd()))
}
