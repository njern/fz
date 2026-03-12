package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/njern/fz/internal/api"
	"github.com/spf13/cobra"
)

var pinCmd = &cobra.Command{
	Use:     "pin <command>",
	Short:   "View your pinned cards",
	Aliases: []string{"pins"},
}

var pinListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List your pinned cards",
	Aliases: []string{"ls"},
	RunE:    runPinList,
}

func init() {
	pinCmd.AddCommand(pinListCmd)
	rootCmd.AddCommand(pinCmd)
}

func runPinList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var pins []api.Card
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/my/pins", slug), &pins); err != nil {
		return err
	}

	if len(pins) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No pinned cards.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(w, "#\tTITLE\tBOARD\tSTATUS")
	for _, c := range pins {
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", c.Number, c.Title, c.Board.Name, c.Status)
	}

	return w.Flush()
}
