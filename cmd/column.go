package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/njern/fz/internal/api"
	"github.com/spf13/cobra"
)

var columnCmd = &cobra.Command{
	Use:     "column <command>",
	Short:   "Manage board columns",
	Aliases: []string{"columns"},
}

var columnListCmd = &cobra.Command{
	Use:     "list <board-id>",
	Short:   "List columns on a board",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(1),
	RunE:    runColumnList,
}

var columnCreateCmd = &cobra.Command{
	Use:   "create <board-id>",
	Short: "Create a column",
	Args:  cobra.ExactArgs(1),
	RunE:  runColumnCreate,
}

var columnEditCmd = &cobra.Command{
	Use:   "edit <board-id> <column-id>",
	Short: "Update a column",
	Args:  cobra.ExactArgs(2),
	RunE:  runColumnEdit,
}

var columnViewCmd = &cobra.Command{
	Use:   "view <board-id> <column-id>",
	Short: "View a column",
	Args:  cobra.ExactArgs(2),
	RunE:  runColumnView,
}

var columnDeleteCmd = &cobra.Command{
	Use:   "delete <board-id> <column-id>",
	Short: "Delete a column",
	Args:  cobra.ExactArgs(2),
	RunE:  runColumnDelete,
}

var (
	columnName  string
	columnColor string
)

type columnRequest struct {
	Column columnPayload `json:"column"`
}

type columnPayload struct {
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

func init() {
	columnCreateCmd.Flags().StringVar(&columnName, "name", "", "Column name (required)")
	columnCreateCmd.Flags().StringVar(&columnColor, "color", "", "Column color (e.g. Blue, Gray, Tan, Yellow, Lime, Aqua, Violet, Purple, Pink)")

	columnEditCmd.Flags().StringVar(&columnName, "name", "", "New column name")
	columnEditCmd.Flags().StringVar(&columnColor, "color", "", "New column color")

	columnCmd.AddCommand(columnListCmd)
	columnCmd.AddCommand(columnViewCmd)
	columnCmd.AddCommand(columnCreateCmd)
	columnCmd.AddCommand(columnEditCmd)
	columnCmd.AddCommand(columnDeleteCmd)
	rootCmd.AddCommand(columnCmd)
}

var colorMap = map[string]string{
	"blue":   "var(--color-card-default)",
	"gray":   "var(--color-card-1)",
	"tan":    "var(--color-card-2)",
	"yellow": "var(--color-card-3)",
	"lime":   "var(--color-card-4)",
	"aqua":   "var(--color-card-5)",
	"violet": "var(--color-card-6)",
	"purple": "var(--color-card-7)",
	"pink":   "var(--color-card-8)",
}

func resolveColor(name string) string {
	if v, ok := colorMap[strings.ToLower(name)]; ok {
		return v
	}
	// Assume it's already a CSS variable.
	return name
}

func runColumnList(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	var columns []api.Column
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/boards/%s/columns", slug, args[0]), &columns); err != nil {
		return err
	}

	if len(columns) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No columns found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(w, "ID\tNAME\tCOLOR")
	for _, c := range columns {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", c.ID, c.Name, c.Color.Name)
	}

	return w.Flush()
}

func runColumnView(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	var col api.Column
	if err := client.Get(cmd.Context(), fmt.Sprintf("/%s/boards/%s/columns/%s", slug, args[0], args[1]), &col); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", col.Name)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:      %s\n", col.ID)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Color:   %s\n", col.Color.Name)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", col.CreatedAt.Format("2006-01-02"))

	return nil
}

func runColumnCreate(cmd *cobra.Command, args []string) error {
	if columnName == "" {
		return fmt.Errorf("--name is required")
	}

	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	payload := columnPayload{Name: columnName}
	if columnColor != "" {
		payload.Color = resolveColor(columnColor)
	}

	body, err := json.Marshal(columnRequest{Column: payload})
	if err != nil {
		return err
	}

	var col api.Column
	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/boards/%s/columns", slug, args[0]), bytes.NewReader(body), &col); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Column %q created.\n", columnName)
	if col.ID != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "ID: %s\n", col.ID)
	}

	return nil
}

func runColumnEdit(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	payload := columnPayload{}
	if columnName != "" {
		payload.Name = columnName
	}

	if columnColor != "" {
		payload.Color = resolveColor(columnColor)
	}

	if payload == (columnPayload{}) {
		return fmt.Errorf("nothing to update; use --name or --color")
	}

	body, err := json.Marshal(columnRequest{Column: payload})
	if err != nil {
		return err
	}

	if err := client.Put(cmd.Context(), fmt.Sprintf("/%s/boards/%s/columns/%s", slug, args[0], args[1]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Column updated.")

	return nil
}

func runColumnDelete(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Delete column %s?", args[1]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/boards/%s/columns/%s", slug, args[0], args[1])); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Column deleted.")

	return nil
}
