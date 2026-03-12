package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/njern/fz/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// testEnv sets up a mock HTTP server and overrides global state for command tests.
// It saves and restores cfg and PersistentPreRunE via t.Cleanup.
// Tests using this must NOT call t.Parallel().
func testEnv(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	origCfg := cfg
	origCfgLoadErr := cfgLoadErr
	origPreRun := rootCmd.PersistentPreRunE
	origAccount := cfgOverrideAccount
	origVersion := cliVersion

	t.Cleanup(func() {
		cfg = origCfg
		cfgLoadErr = origCfgLoadErr
		rootCmd.PersistentPreRunE = origPreRun
		cfgOverrideAccount = origAccount
		cliVersion = origVersion
	})

	cliVersion = "test"

	cfg = &config.Config{
		Host:           srv.URL,
		Token:          "test-token",
		DefaultAccount: "test-account",
	}

	// Skip real config loading in tests.
	rootCmd.PersistentPreRunE = nil
	cfgOverrideAccount = ""
	cfgLoadErr = nil

	return srv
}

type commandResult struct {
	stdout string
	stderr string
	err    error
}

// executeCommand runs a cobra command with the given args using Cobra-managed streams.
func executeCommand(t *testing.T, args ...string) commandResult {
	t.Helper()
	return executeCommandWithInput(t, strings.NewReader(""), args...)
}

// executeCommandWithInput runs a cobra command with the provided stdin.
func executeCommandWithInput(t *testing.T, input io.Reader, args ...string) commandResult {
	t.Helper()

	if input == nil {
		input = strings.NewReader("")
	}

	var stdout, stderr bytes.Buffer

	origOut := rootCmd.OutOrStdout()
	origErr := rootCmd.ErrOrStderr()
	origIn := rootCmd.InOrStdin()

	resetFlagSet(rootCmd.PersistentFlags())
	resetFlagSet(rootCmd.Flags())

	rootCmd.SetArgs(args)
	rootCmd.SetIn(input)
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	err := rootCmd.Execute()

	rootCmd.SetOut(origOut)
	rootCmd.SetErr(origErr)
	rootCmd.SetIn(origIn)
	rootCmd.SetArgs(nil)

	return commandResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
		err:    err,
	}
}

func resetFlagSet(flagSet *pflag.FlagSet) {
	flagSet.VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
}

// resetFlags resets cobra flag values that persist between test runs.
// Call this in tests that exercise commands with flags.
func resetFlags(t *testing.T, cmds ...*cobra.Command) {
	t.Helper()

	for _, cmd := range cmds {
		resetFlagSet(cmd.Flags())
	}
}
