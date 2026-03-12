package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/njern/fz/internal/api"
	"github.com/njern/fz/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth <command>",
	Short: "Authenticate with Fizzy",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Fizzy",
	Long: `Authenticate with Fizzy using a magic link or a personal access token.

By default, initiates a magic link flow: enter your email, check your inbox,
and enter the 6-character code.

Alternatively, pipe a personal access token via stdin with --with-token.`,
	RunE: runAuthLogin,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of Fizzy",
	RunE:  runAuthLogout,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display authentication status",
	RunE:  runAuthStatus,
}

var authTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Print the authentication token",
	Long:  "Print the authentication token to stdout, useful for piping to other commands.",
	RunE:  runAuthToken,
}

var authCreateTokenCmd = &cobra.Command{
	Use:   "create-token",
	Short: "Create a personal access token",
	Long:  "Create a personal access token via the API. Requires an existing session or write-capable token.",
	RunE:  runAuthCreateToken,
}

var (
	authWithToken             bool
	authStatusCheck           bool
	authCreateTokenDesc       string
	authCreateTokenPermission string
)

func init() {
	authLoginCmd.Flags().BoolVar(&authWithToken, "with-token", false, "Read token from stdin")
	authStatusCmd.Flags().BoolVar(&authStatusCheck, "check", false, "Exit non-zero if not authenticated or the token is invalid")

	authCreateTokenCmd.Flags().StringVar(&authCreateTokenDesc, "description", "", "Token description (required)")
	authCreateTokenCmd.Flags().StringVar(&authCreateTokenPermission, "permission", "read", "Token permission: read or write")

	withConfigMode(authLoginCmd, configLoadRepairable)
	withConfigMode(authLogoutCmd, configLoadRepairable)

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authTokenCmd)
	authCmd.AddCommand(authCreateTokenCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	if authWithToken {
		return loginWithToken(cmd)
	}

	return loginWithMagicLink(cmd)
}

func loginWithToken(cmd *cobra.Command) error {
	return loginWithTokenFromReader(cmd.Context(), cmd.InOrStdin(), cmd.ErrOrStderr(), stdinIsInteractive(cmd))
}

func loginWithTokenFromReader(ctx context.Context, r io.Reader, stderr io.Writer, interactive bool) error {
	reader := bufio.NewReader(r)

	token, err := reader.ReadString('\n')
	if err != nil && token == "" {
		return fmt.Errorf("reading token from stdin: %w", err)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("empty token")
	}

	// Verify the token works by fetching identity.
	client := api.NewClient(cfg.Host, token, cliVersion)

	var identity api.Identity
	if err := client.Get(ctx, "/my/identity", &identity); err != nil {
		return fmt.Errorf("token verification failed: %w", err)
	}

	cfg.Token = token

	if err := selectDefaultAccount(&identity, reader, stderr, interactive); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stderr, "Logged in to %s\n", cfg.Host)

	return nil
}

func loginWithMagicLink(cmd *cobra.Command) error {
	reader := bufio.NewReader(cmd.InOrStdin())
	stderr := cmd.ErrOrStderr()
	httpClient := &http.Client{Timeout: 30 * time.Second}

	_, _ = fmt.Fprint(stderr, "Email address: ")

	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading email: %w", err)
	}

	email = strings.TrimSpace(email)

	pendingToken, err := auth.MagicLinkRequest(cmd.Context(), httpClient, cfg.Host, email)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(stderr, "Check your email for a magic link code.")
	_, _ = fmt.Fprint(stderr, "Code: ")

	code, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading code: %w", err)
	}

	code = strings.TrimSpace(code)

	sessionToken, err := auth.MagicLinkVerify(cmd.Context(), httpClient, cfg.Host, pendingToken, code)
	if err != nil {
		return err
	}

	// Fetch identity using the session cookie to discover accounts.
	identityJSON, err := auth.FetchIdentity(cmd.Context(), httpClient, cfg.Host, sessionToken)
	if err != nil {
		return fmt.Errorf("fetching identity: %w", err)
	}

	var identity api.Identity
	if err := json.Unmarshal(identityJSON, &identity); err != nil {
		return fmt.Errorf("decoding identity: %w", err)
	}

	if len(identity.Accounts) == 0 {
		return fmt.Errorf("no accounts found for this identity")
	}

	if err := selectDefaultAccount(&identity, reader, stderr, stdinIsInteractive(cmd)); err != nil {
		return err
	}

	// Create a long-lived access token using the session cookie.
	_, _ = fmt.Fprintln(stderr, "Creating a personal access token...")

	accessToken, err := auth.CreateAccessToken(cmd.Context(), httpClient, cfg.Host, cfg.DefaultAccount, sessionToken, "fz CLI", "write", false)
	if err != nil {
		return fmt.Errorf("creating access token: %w", err)
	}

	cfg.Token = accessToken

	if err := cfg.Save(); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stderr, "Logged in to %s\n", cfg.Host)

	return nil
}

func selectDefaultAccount(identity *api.Identity, reader *bufio.Reader, stderr io.Writer, interactive bool) error {
	if len(identity.Accounts) == 0 {
		return fmt.Errorf("no accounts found for this identity")
	}

	if cfgOverrideAccount != "" {
		override := strings.TrimPrefix(cfgOverrideAccount, "/")
		for _, acct := range identity.Accounts {
			if acct.SlugTrimmed() == override {
				cfg.DefaultAccount = override
				_, _ = fmt.Fprintf(stderr, "Default account: %s (%s)\n", acct.Name, override)

				return nil
			}
		}

		return fmt.Errorf("account %q was not found in this identity", override)
	}

	if len(identity.Accounts) == 1 {
		slug := identity.Accounts[0].SlugTrimmed()
		cfg.DefaultAccount = slug
		_, _ = fmt.Fprintf(stderr, "Default account: %s (%s)\n", identity.Accounts[0].Name, slug)

		return nil
	}

	if !interactive {
		return fmt.Errorf("multiple accounts found; re-run with --account to select one")
	}

	_, _ = fmt.Fprintln(stderr, "Select a default account:")

	for i, acct := range identity.Accounts {
		slug := acct.SlugTrimmed()
		_, _ = fmt.Fprintf(stderr, "  [%d] %s (%s)\n", i+1, acct.Name, slug)
	}

	_, _ = fmt.Fprint(stderr, "Choice: ")

	choice, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading choice: %w", err)
	}

	var idx int
	if _, err := fmt.Sscanf(strings.TrimSpace(choice), "%d", &idx); err != nil || idx < 1 || idx > len(identity.Accounts) {
		return fmt.Errorf("invalid choice")
	}

	slug := identity.Accounts[idx-1].SlugTrimmed()
	cfg.DefaultAccount = slug
	_, _ = fmt.Fprintf(stderr, "Default account: %s (%s)\n", identity.Accounts[idx-1].Name, slug)

	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	if cfgLoadErr == nil && !cfg.Authenticated() {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Not logged in.")
		return nil
	}

	confirmed, err := confirmAction(cmd, "Log out and remove the stored token?")
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if cfgLoadErr != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Stored config was invalid; rewriting it with a logged-out state.")
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Note: The server-side token has not been revoked. Revoke it manually in your Fizzy account settings if needed.")

	cfg.Token = ""

	cfg.DefaultAccount = ""
	if err := cfg.Save(); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Logged out.")

	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	if !cfg.Authenticated() {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Not logged in.")
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Run `fz auth login` to authenticate.")

		if authStatusCheck {
			return ErrNotAuthenticated
		}

		return nil
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Host:    %s\n", cfg.Host)
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Account: %s\n", cfg.DefaultAccount)
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Token:   ****")

	// Try to verify the token.
	client, err := newClient()
	if err != nil {
		return err
	}

	var identity api.Identity
	if err := client.Get(cmd.Context(), "/my/identity", &identity); err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Status:  Token is invalid or expired")

		if authStatusCheck {
			return ErrInvalidToken
		}

		return nil
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Status:  Authenticated")

	for _, acct := range identity.Accounts {
		slug := acct.SlugTrimmed()

		marker := "  "
		if slug == cfg.DefaultAccount {
			marker = "* "
		}

		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  %s%s (%s) as %s (%s)\n", marker, acct.Name, slug, acct.User.Name, acct.User.Role)
	}

	return nil
}

func runAuthToken(cmd *cobra.Command, args []string) error {
	if !cfg.Authenticated() {
		return ErrNotAuthenticated
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), cfg.Token)

	return nil
}

func runAuthCreateToken(cmd *cobra.Command, args []string) error {
	if authCreateTokenDesc == "" {
		return fmt.Errorf("--description is required")
	}

	if authCreateTokenPermission != "read" && authCreateTokenPermission != "write" {
		return fmt.Errorf("--permission must be \"read\" or \"write\"")
	}

	if !cfg.Authenticated() {
		return ErrNotAuthenticated
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	token, err := auth.CreateAccessToken(cmd.Context(), nil, cfg.Host, slug, cfg.Token, authCreateTokenDesc, authCreateTokenPermission, true)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), token)

	return nil
}
