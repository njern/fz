package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/njern/fz/internal/api"
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:     "tag <command>",
	Short:   "Manage tags",
	Aliases: []string{"tags"},
}

var tagListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all tags in the account",
	Aliases: []string{"ls"},
	RunE:    runTagList,
}

func init() {
	tagCmd.AddCommand(tagListCmd)
	rootCmd.AddCommand(tagCmd)
}

func runTagList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var tags []api.Tag
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/tags", slug), &tags); err != nil {
		return err
	}

	if len(tags) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No tags found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(w, "ID\tTITLE")
	for _, t := range tags {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", t.ID, t.Title)
	}

	return w.Flush()
}
