package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"text/tabwriter"

	"github.com/guregu/null/v5"
	"github.com/njern/fz/internal/api"
	"github.com/njern/fz/internal/render"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var boardCmd = &cobra.Command{
	Use:     "board <command>",
	Short:   "Manage boards",
	Aliases: []string{"boards"},
}

var boardListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List boards",
	Aliases: []string{"ls"},
	RunE:    runBoardList,
}

var boardViewCmd = &cobra.Command{
	Use:   "view <board-id>",
	Short: "View a board",
	Args:  cobra.ExactArgs(1),
	RunE:  runBoardView,
}

var boardCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a board",
	Args:  cobra.ExactArgs(1),
	RunE:  runBoardCreate,
}

var boardDeleteCmd = &cobra.Command{
	Use:   "delete <board-id>",
	Short: "Delete a board",
	Args:  cobra.ExactArgs(1),
	RunE:  runBoardDelete,
}

var boardPublishCmd = &cobra.Command{
	Use:   "publish <board-id>",
	Short: "Publish a board",
	Args:  cobra.ExactArgs(1),
	RunE:  runBoardPublish,
}

var boardUnpublishCmd = &cobra.Command{
	Use:   "unpublish <board-id>",
	Short: "Unpublish a board",
	Args:  cobra.ExactArgs(1),
	RunE:  runBoardUnpublish,
}

var boardEditCmd = &cobra.Command{
	Use:   "edit <board-id>",
	Short: "Edit a board",
	Args:  cobra.ExactArgs(1),
	RunE:  runBoardEdit,
}

var (
	boardAutoPostpone    int
	boardEditName        string
	boardEditAccess      string
	boardEditDescription string
	boardEditUserIDs     []string
	boardClearUserIDs    bool
	boardViewJSONFlag    bool
	boardViewWebFlag     bool
	boardListJSONFlag    bool
)

type boardRequest struct {
	Board boardPayload `json:"board"`
}

type boardPayload struct {
	Name               string       `json:"name,omitempty"`
	AutoPostponePeriod *int         `json:"auto_postpone_period,omitempty"`
	AllAccess          *bool        `json:"all_access,omitempty"`
	PublicDescription  *null.String `json:"public_description,omitempty"`
	UserIDs            *[]string    `json:"user_ids,omitempty"`
}

func (p boardPayload) hasUpdates() bool {
	return p.Name != "" ||
		p.AutoPostponePeriod != nil ||
		p.AllAccess != nil ||
		p.PublicDescription != nil ||
		p.UserIDs != nil
}

func init() {
	boardCreateCmd.Flags().IntVar(&boardAutoPostpone, "auto-postpone", 0, "Days of inactivity before auto-postpone")

	boardEditCmd.Flags().StringVar(&boardEditName, "name", "", "Board name")
	boardEditCmd.Flags().IntVar(&boardAutoPostpone, "auto-postpone", 0, "Days of inactivity before auto-postpone")
	boardEditCmd.Flags().StringVar(&boardEditAccess, "access", "", "Access: all or selective")
	boardEditCmd.Flags().StringVar(&boardEditDescription, "description", "", "Public description")
	boardEditCmd.Flags().StringSliceVar(&boardEditUserIDs, "user-ids", nil, "User IDs for selective access (repeatable)")
	boardEditCmd.Flags().BoolVar(&boardClearUserIDs, "clear-user-ids", false, "Clear all selective-access user IDs")

	boardViewCmd.Flags().BoolVar(&boardViewJSONFlag, "json", false, "Output as JSON")
	boardViewCmd.Flags().BoolVar(&boardViewWebFlag, "web", false, "Open in browser")
	boardListCmd.Flags().BoolVar(&boardListJSONFlag, "json", false, "Output as JSON")

	boardCmd.AddCommand(boardListCmd)
	boardCmd.AddCommand(boardViewCmd)
	boardCmd.AddCommand(boardCreateCmd)
	boardCmd.AddCommand(boardEditCmd)
	boardCmd.AddCommand(boardDeleteCmd)
	boardCmd.AddCommand(boardPublishCmd)
	boardCmd.AddCommand(boardUnpublishCmd)
	rootCmd.AddCommand(boardCmd)
}

func runBoardList(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	var boards []api.Board
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/boards", slug), &boards); err != nil {
		return err
	}

	if boardListJSONFlag {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")

		return enc.Encode(boards)
	}

	if len(boards) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No boards found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tNAME\tACCESS\tCREATED")

	for _, b := range boards {
		access := "selective"
		if b.AllAccess {
			access = "all"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", b.ID, b.Name, access, b.CreatedAt.Format("2006-01-02"))
	}

	return w.Flush()
}

func runBoardView(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	boardPath := fmt.Sprintf("/%s/boards/%s", slug, args[0])

	var board api.Board
	if err := client.Get(cmd.Context(), boardPath, &board); err != nil {
		return err
	}

	if boardViewWebFlag {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Opening %s in browser...\n", board.URL)
		return openInBrowser(board.URL)
	}

	var columns []api.Column
	if err := client.GetAll(cmd.Context(), boardPath+"/columns", &columns); err != nil {
		return err
	}

	cards, err := fetchBoardViewCards(cmd.Context(), client, slug, args[0])
	if err != nil {
		return err
	}

	if boardViewJSONFlag {
		return printBoardJSON(cmd.OutOrStdout(), board, columns, cards)
	}

	_, _ = fmt.Fprint(cmd.OutOrStdout(), render.BoardView(board.Name, columns, cards, render.TerminalWidth()))

	return nil
}

func fetchBoardViewCards(ctx context.Context, client *api.Client, slug, boardID string) ([]api.Card, error) {
	indexes := []string{"", "not_now", "closed"}
	results := make([][]api.Card, len(indexes))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(len(indexes))

	for i, index := range indexes {
		i, index := i, index

		g.Go(func() error {
			params := url.Values{}
			params.Add("board_ids[]", boardID)
			if index != "" {
				params.Set("indexed_by", index)
			}

			var cards []api.Card
			if err := client.GetAll(ctx, fmt.Sprintf("/%s/cards?%s", slug, params.Encode()), &cards); err != nil {
				return err
			}

			results[i] = cards
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var all []api.Card

	for _, cards := range results {
		for _, card := range cards {
			if _, ok := seen[card.ID]; ok {
				continue
			}

			seen[card.ID] = struct{}{}
			all = append(all, card)
		}
	}

	return all, nil
}

func printBoardJSON(w io.Writer, board api.Board, columns []api.Column, cards []api.Card) error {
	byColumn := map[string][]api.Card{}

	var maybeCards, notNowCards, doneCards []api.Card

	for _, c := range cards {
		if c.Column != nil {
			byColumn[c.Column.ID] = append(byColumn[c.Column.ID], c)
		} else if c.Closed {
			doneCards = append(doneCards, c)
		} else if c.Postponed {
			notNowCards = append(notNowCards, c)
		} else {
			maybeCards = append(maybeCards, c)
		}
	}

	type jsonColumn struct {
		ID    string     `json:"id"`
		Name  string     `json:"name"`
		Color string     `json:"color,omitempty"`
		Cards []api.Card `json:"cards"`
	}

	out := struct {
		Board   api.Board    `json:"board"`
		Columns []jsonColumn `json:"columns"`
	}{Board: board}

	out.Columns = append(out.Columns, jsonColumn{Name: "Not Now", Color: render.BoardLaneColorComplete, Cards: notNowCards})
	out.Columns = append(out.Columns, jsonColumn{Name: "Maybe?", Color: render.BoardLaneColorMaybe, Cards: maybeCards})

	for _, col := range columns {
		out.Columns = append(out.Columns, jsonColumn{
			ID:    col.ID,
			Name:  col.Name,
			Color: col.Color.Name,
			Cards: byColumn[col.ID],
		})
	}

	out.Columns = append(out.Columns, jsonColumn{Name: "Done", Color: render.BoardLaneColorComplete, Cards: doneCards})

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(out)
}

func runBoardCreate(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	payload := boardPayload{Name: args[0]}

	if boardAutoPostpone > 0 {
		autoPostpone := boardAutoPostpone
		payload.AutoPostponePeriod = &autoPostpone
	}

	body, err := json.Marshal(boardRequest{Board: payload})
	if err != nil {
		return err
	}

	var board api.Board
	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/boards", slug), bytes.NewReader(body), &board); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Board %q created.\n", args[0])
	if board.ID != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "ID:  %s\n", board.ID)
	}

	if board.URL != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "URL: %s\n", board.URL)
	}

	return nil
}

func runBoardEdit(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	payload := boardPayload{}
	if boardEditName != "" {
		payload.Name = boardEditName
	}

	if cmd.Flags().Changed("auto-postpone") {
		autoPostpone := boardAutoPostpone
		payload.AutoPostponePeriod = &autoPostpone
	}

	if boardEditAccess != "" {
		if boardEditAccess != "all" && boardEditAccess != "selective" {
			return fmt.Errorf("--access must be \"all\" or \"selective\"")
		}

		allAccess := boardEditAccess == "all"
		payload.AllAccess = &allAccess
	}

	if cmd.Flags().Changed("description") {
		description := null.StringFrom(boardEditDescription)
		payload.PublicDescription = &description
	}

	userIDsChanged := cmd.Flags().Changed("user-ids")
	if boardClearUserIDs && userIDsChanged {
		return fmt.Errorf("use either --user-ids or --clear-user-ids, not both")
	}

	if boardClearUserIDs {
		userIDs := []string{}
		payload.UserIDs = &userIDs
	} else if userIDsChanged {
		userIDs := append([]string(nil), boardEditUserIDs...)
		payload.UserIDs = &userIDs
	}

	if !payload.hasUpdates() {
		return fmt.Errorf("nothing to update; use --name, --auto-postpone, --access, --description, or --user-ids")
	}

	body, err := json.Marshal(boardRequest{Board: payload})
	if err != nil {
		return err
	}

	if err := client.Put(cmd.Context(), fmt.Sprintf("/%s/boards/%s", slug, args[0]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Board updated.")

	return nil
}

func runBoardDelete(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Delete board %s?", args[0]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/boards/%s", slug, args[0])); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Board deleted.")

	return nil
}

func runBoardPublish(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	var board api.Board
	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/boards/%s/publication", slug, args[0]), nil, &board); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Board published: %s\n", board.PublicURL)

	return nil
}

func runBoardUnpublish(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Unpublish board %s?", args[0]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/boards/%s/publication", slug, args[0])); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Board unpublished.")

	return nil
}
