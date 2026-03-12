package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config <command>",
	Short: "Manage configuration for fz",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  host       Fizzy instance URL (default: https://app.fizzy.do)
  account    Default account slug`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all configuration values",
	Aliases: []string{"ls"},
	RunE:    runConfigList,
}

func init() {
	withConfigMode(configCmd, configLoadRepairable)

	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	switch key {
	case "host":
		cfg.Host = value
	case "account":
		cfg.DefaultAccount = strings.TrimPrefix(value, "/")
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Set %s = %s\n", key, value)

	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	var value string

	switch key {
	case "host":
		value = cfg.Host
	case "account":
		value = cfg.DefaultAccount
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), value)

	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "host=%s\n", cfg.Host)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "account=%s\n", cfg.DefaultAccount)

	return nil
}
