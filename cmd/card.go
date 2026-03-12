package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/njern/fz/internal/api"
	"github.com/spf13/cobra"
)

var cardCmd = &cobra.Command{
	Use:     "card <command>",
	Short:   "Manage cards",
	Aliases: []string{"cards"},
}

// List flags.
var (
	cardListBoard      string
	cardListTags       []string
	cardListAssignee   string
	cardListStatus     string
	cardListSort       string
	cardListSearch     []string
	cardListUnassigned bool
	cardListCreated    string
	cardListClosed     string
)

// Create flags.
var (
	cardCreateBoard string
	cardCreateTitle string
	cardCreateBody  string
	cardCreateDraft bool
	cardCreateTags  []string
)

// Edit flags.
var (
	cardEditTitle string
	cardEditBody  string
)

// Triage flags.
var cardTriageColumn string

// Assign flags.
var cardAssignUser string

// Tag flags.
var cardTagTitle string

// Reaction flags.
var cardReactionBody string

var cardListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List cards",
	Aliases: []string{"ls"},
	RunE:    runCardList,
}

var cardViewCmd = &cobra.Command{
	Use:   "view <number>",
	Short: "View a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardView,
}

var cardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a card",
	RunE:  runCardCreate,
}

var cardEditCmd = &cobra.Command{
	Use:   "edit <number>",
	Short: "Edit a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardEdit,
}

var cardDeleteCmd = &cobra.Command{
	Use:   "delete <number>",
	Short: "Delete a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardDelete,
}

var cardCloseCmd = &cobra.Command{
	Use:   "close <number>",
	Short: "Close a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardClose,
}

var cardReopenCmd = &cobra.Command{
	Use:   "reopen <number>",
	Short: "Reopen a closed card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardReopen,
}

var cardTriageCmd = &cobra.Command{
	Use:   "triage <number>",
	Short: "Move a card into a column",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardTriage,
}

var cardUnTriageCmd = &cobra.Command{
	Use:   "untriage <number>",
	Short: "Send a card back to triage",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardUntriage,
}

var cardPostponeCmd = &cobra.Command{
	Use:   "postpone <number>",
	Short: "Move a card to Not Now",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardPostpone,
}

var cardAssignCmd = &cobra.Command{
	Use:   "assign <number>",
	Short: "Toggle user assignment on a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardAssign,
}

var cardTagCmd = &cobra.Command{
	Use:   "tag <number>",
	Short: "Toggle a tag on a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardTag,
}

var cardPinCmd = &cobra.Command{
	Use:   "pin <number>",
	Short: "Pin a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardPin,
}

var cardUnpinCmd = &cobra.Command{
	Use:   "unpin <number>",
	Short: "Unpin a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardUnpin,
}

var cardWatchCmd = &cobra.Command{
	Use:   "watch <number>",
	Short: "Watch a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardWatch,
}

var cardUnwatchCmd = &cobra.Command{
	Use:   "unwatch <number>",
	Short: "Unwatch a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardUnwatch,
}

var cardGildCmd = &cobra.Command{
	Use:   "gild <number>",
	Short: "Mark a card as golden",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardGild,
}

var cardUngildCmd = &cobra.Command{
	Use:   "ungild <number>",
	Short: "Remove golden status from a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardUngild,
}

var cardRemoveImageCmd = &cobra.Command{
	Use:   "remove-image <number>",
	Short: "Remove the header image from a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardRemoveImage,
}

var cardReactionCmd = &cobra.Command{
	Use:   "reaction <command>",
	Short: "Manage card reactions",
}

var cardReactionListCmd = &cobra.Command{
	Use:     "list <number>",
	Short:   "List reactions on a card",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(1),
	RunE:    runCardReactionList,
}

var cardReactionCreateCmd = &cobra.Command{
	Use:   "create <number>",
	Short: "Add a reaction to a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardReactionCreate,
}

var cardReactionDeleteCmd = &cobra.Command{
	Use:   "delete <number> <reaction-id>",
	Short: "Remove a reaction from a card",
	Args:  cobra.ExactArgs(2),
	RunE:  runCardReactionDelete,
}

var (
	cardViewComments bool
	cardListJSONFlag bool
	cardViewJSONFlag bool
	cardViewWebFlag  bool
)

func init() {
	// List flags.
	cardListCmd.Flags().StringVarP(&cardListBoard, "board", "b", "", "Filter by board ID")
	cardListCmd.Flags().StringSliceVarP(&cardListTags, "tag", "t", nil, "Filter by tag ID (repeatable)")
	cardListCmd.Flags().StringVar(&cardListAssignee, "assignee", "", "Filter by assignee user ID")
	cardListCmd.Flags().StringVar(&cardListStatus, "status", "", "Filter: closed, not_now, stalled, golden, postponing_soon")
	cardListCmd.Flags().StringVar(&cardListSort, "sort", "", "Sort: latest, newest, oldest")
	cardListCmd.Flags().StringSliceVarP(&cardListSearch, "search", "S", nil, "Search terms")
	cardListCmd.Flags().BoolVar(&cardListUnassigned, "unassigned", false, "Only unassigned cards")
	cardListCmd.Flags().StringVar(&cardListCreated, "created", "", "Filter by creation: today, thisweek, etc.")
	cardListCmd.Flags().StringVar(&cardListClosed, "closed", "", "Filter by closure: today, thisweek, etc.")
	cardListCmd.Flags().BoolVar(&cardListJSONFlag, "json", false, "Output as JSON")

	// View flags.
	cardViewCmd.Flags().BoolVarP(&cardViewComments, "comments", "c", false, "Show comments")
	cardViewCmd.Flags().BoolVar(&cardViewJSONFlag, "json", false, "Output as JSON")
	cardViewCmd.Flags().BoolVar(&cardViewWebFlag, "web", false, "Open in browser")

	// Create flags.
	cardCreateCmd.Flags().StringVarP(&cardCreateBoard, "board", "b", "", "Board ID (required)")
	cardCreateCmd.Flags().StringVarP(&cardCreateTitle, "title", "t", "", "Card title (required)")
	cardCreateCmd.Flags().StringVarP(&cardCreateBody, "body", "B", "", "Card description")
	cardCreateCmd.Flags().BoolVar(&cardCreateDraft, "draft", false, "Create as draft")
	cardCreateCmd.Flags().StringSliceVar(&cardCreateTags, "tag-id", nil, "Tag IDs (repeatable)")

	// Edit flags.
	cardEditCmd.Flags().StringVarP(&cardEditTitle, "title", "t", "", "New title")
	cardEditCmd.Flags().StringVarP(&cardEditBody, "body", "B", "", "New description")

	// Triage flags.
	cardTriageCmd.Flags().StringVarP(&cardTriageColumn, "column", "c", "", "Column ID (required)")

	// Assign flags.
	cardAssignCmd.Flags().StringVar(&cardAssignUser, "assignee", "", "User ID to assign/unassign")

	// Tag flags.
	cardTagCmd.Flags().StringVar(&cardTagTitle, "tag", "", "Tag title to toggle")

	// Reaction flags.
	cardReactionCreateCmd.Flags().StringVarP(&cardReactionBody, "body", "B", "", "Reaction text (max 16 chars)")

	cardReactionCmd.AddCommand(cardReactionListCmd)
	cardReactionCmd.AddCommand(cardReactionCreateCmd)
	cardReactionCmd.AddCommand(cardReactionDeleteCmd)

	cardCmd.AddCommand(cardListCmd)
	cardCmd.AddCommand(cardViewCmd)
	cardCmd.AddCommand(cardCreateCmd)
	cardCmd.AddCommand(cardEditCmd)
	cardCmd.AddCommand(cardDeleteCmd)
	cardCmd.AddCommand(cardCloseCmd)
	cardCmd.AddCommand(cardReopenCmd)
	cardCmd.AddCommand(cardTriageCmd)
	cardCmd.AddCommand(cardUnTriageCmd)
	cardCmd.AddCommand(cardPostponeCmd)
	cardCmd.AddCommand(cardAssignCmd)
	cardCmd.AddCommand(cardTagCmd)
	cardCmd.AddCommand(cardPinCmd)
	cardCmd.AddCommand(cardUnpinCmd)
	cardCmd.AddCommand(cardWatchCmd)
	cardCmd.AddCommand(cardUnwatchCmd)
	cardCmd.AddCommand(cardGildCmd)
	cardCmd.AddCommand(cardUngildCmd)
	cardCmd.AddCommand(cardRemoveImageCmd)
	cardCmd.AddCommand(cardReactionCmd)
	rootCmd.AddCommand(cardCmd)
}

func runCardList(cmd *cobra.Command, args []string) error {
	if err := validateCardListFilters(); err != nil {
		return err
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	params := url.Values{}
	if cardListBoard != "" {
		params.Add("board_ids[]", cardListBoard)
	}

	for _, t := range cardListTags {
		params.Add("tag_ids[]", t)
	}

	if cardListAssignee != "" {
		params.Add("assignee_ids[]", cardListAssignee)
	}

	if cardListStatus != "" {
		params.Add("indexed_by", cardListStatus)
	}

	if cardListSort != "" {
		params.Add("sorted_by", cardListSort)
	}

	for _, s := range cardListSearch {
		params.Add("terms[]", s)
	}

	if cardListUnassigned {
		params.Add("assignment_status", "unassigned")
	}

	if cardListCreated != "" {
		params.Add("creation", cardListCreated)
	}

	if cardListClosed != "" {
		params.Add("closure", cardListClosed)
	}

	path := fmt.Sprintf("/%s/cards", slug)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var cards []api.Card
	if err := client.GetAll(cmd.Context(), path, &cards); err != nil {
		return err
	}

	if cardListJSONFlag {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")

		return enc.Encode(cards)
	}

	if len(cards) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No cards found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "#\tTITLE\tSTATUS\tBOARD\tTAGS")

	for _, c := range cards {
		tags := strings.Join(c.Tags, ", ")
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", c.Number, c.Title, c.Status, c.Board.Name, tags)
	}

	return w.Flush()
}

func runCardView(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var card api.Card
	if err := client.Get(cmd.Context(), fmt.Sprintf("/%s/cards/%s", slug, args[0]), &card); err != nil {
		return err
	}

	if cardViewWebFlag {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Opening %s in browser...\n", card.URL)
		return openInBrowser(card.URL)
	}

	if cardViewJSONFlag {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")

		return enc.Encode(card)
	}

	printCard(cmd.OutOrStdout(), &card)

	if cardViewComments {
		_, _ = fmt.Fprintln(cmd.OutOrStdout())

		var comments []api.Comment
		if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/cards/%s/comments", slug, args[0]), &comments); err != nil {
			return err
		}

		if len(comments) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No comments.")
		}

		for _, c := range comments {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "--- %s (%s) ---\n", c.Creator.Name, c.CreatedAt.Format("2006-01-02 15:04"))
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), c.Body.PlainText)
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		}
	}

	return nil
}

func printCard(w io.Writer, card *api.Card) {
	_, _ = fmt.Fprintf(w, "#%d %s\n", card.Number, card.Title)
	_, _ = fmt.Fprintf(w, "Status:  %s\n", card.Status)

	_, _ = fmt.Fprintf(w, "Board:   %s\n", card.Board.Name)
	if card.Column != nil {
		_, _ = fmt.Fprintf(w, "Column:  %s\n", card.Column.Name)
	}

	_, _ = fmt.Fprintf(w, "Creator: %s\n", card.Creator.Name)
	if len(card.Tags) > 0 {
		_, _ = fmt.Fprintf(w, "Tags:    %s\n", strings.Join(card.Tags, ", "))
	}

	if card.Golden {
		_, _ = fmt.Fprintln(w, "Golden:  yes")
	}

	_, _ = fmt.Fprintf(w, "Created: %s\n", card.CreatedAt.Format("2006-01-02 15:04"))
	_, _ = fmt.Fprintf(w, "URL:     %s\n", card.URL)

	if card.Description != "" {
		_, _ = fmt.Fprintf(w, "\n%s\n", card.Description)
	}

	if len(card.Steps) > 0 {
		_, _ = fmt.Fprintln(w)

		for _, s := range card.Steps {
			check := " "
			if s.Completed {
				check = "x"
			}

			_, _ = fmt.Fprintf(w, "  [%s] %s  (%s)\n", check, s.Content, s.ID)
		}
	}
}

func runCardCreate(cmd *cobra.Command, args []string) error {
	if cardCreateBoard == "" {
		return fmt.Errorf("--board is required")
	}

	if cardCreateTitle == "" {
		return fmt.Errorf("--title is required")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	cardPayload := map[string]any{
		"title": cardCreateTitle,
	}
	if cardCreateBody != "" {
		cardPayload["description"] = cardCreateBody
	}

	if cardCreateDraft {
		cardPayload["status"] = "drafted"
	}

	if len(cardCreateTags) > 0 {
		cardPayload["tag_ids"] = cardCreateTags
	}

	body, err := json.Marshal(map[string]any{"card": cardPayload})
	if err != nil {
		return err
	}

	var card api.Card
	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/boards/%s/cards", slug, cardCreateBoard), bytes.NewReader(body), &card); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card %q created.\n", cardCreateTitle)
	if card.Number > 0 {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "#:   %d\n", card.Number)
	}

	if card.URL != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "URL: %s\n", card.URL)
	}

	return nil
}

func runCardEdit(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	cardPayload := map[string]any{}
	if cardEditTitle != "" {
		cardPayload["title"] = cardEditTitle
	}

	if cardEditBody != "" {
		cardPayload["description"] = cardEditBody
	}

	if len(cardPayload) == 0 {
		return fmt.Errorf("nothing to update; use --title or --body")
	}

	body, err := json.Marshal(map[string]any{"card": cardPayload})
	if err != nil {
		return err
	}

	var card api.Card
	if err := client.Put(cmd.Context(), fmt.Sprintf("/%s/cards/%s", slug, args[0]), bytes.NewReader(body), &card); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card #%s updated.\n", args[0])

	return nil
}

func runCardDelete(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Delete card #%s?", args[0]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/cards/%s", slug, args[0])); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card #%s deleted.\n", args[0])

	return nil
}

func runCardClose(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/cards/%s/closure", slug, args[0]), nil, nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card #%s closed.\n", args[0])

	return nil
}

func runCardReopen(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/cards/%s/closure", slug, args[0])); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card #%s reopened.\n", args[0])

	return nil
}

func runCardTriage(cmd *cobra.Command, args []string) error {
	if cardTriageColumn == "" {
		return fmt.Errorf("--column is required")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{"column_id": cardTriageColumn})
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/cards/%s/triage", slug, args[0]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card #%s triaged.\n", args[0])

	return nil
}

func runCardUntriage(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/cards/%s/triage", slug, args[0])); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card #%s sent back to triage.\n", args[0])

	return nil
}

func runCardPostpone(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/cards/%s/not_now", slug, args[0]), nil, nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card #%s moved to Not Now.\n", args[0])

	return nil
}

func runCardAssign(cmd *cobra.Command, args []string) error {
	if cardAssignUser == "" {
		return fmt.Errorf("--assignee is required")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{"assignee_id": cardAssignUser})
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/cards/%s/assignments", slug, args[0]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card #%s assignment toggled.\n", args[0])

	return nil
}

func runCardTag(cmd *cobra.Command, args []string) error {
	if cardTagTitle == "" {
		return fmt.Errorf("--tag is required")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{"tag_title": cardTagTitle})
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/cards/%s/taggings", slug, args[0]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Tag %q toggled on card #%s.\n", cardTagTitle, args[0])

	return nil
}

func runCardPin(cmd *cobra.Command, args []string) error {
	return cardAction(cmd, "pin", args[0])
}

func runCardUnpin(cmd *cobra.Command, args []string) error {
	return cardActionDelete(cmd, "pin", args[0])
}

func runCardWatch(cmd *cobra.Command, args []string) error {
	return cardAction(cmd, "watch", args[0])
}

func runCardUnwatch(cmd *cobra.Command, args []string) error {
	return cardActionDelete(cmd, "watch", args[0])
}

func runCardGild(cmd *cobra.Command, args []string) error {
	return cardAction(cmd, "goldness", args[0])
}

func runCardUngild(cmd *cobra.Command, args []string) error {
	return cardActionDelete(cmd, "goldness", args[0])
}

// cardAction performs a POST to /:slug/cards/:number/:action.
func cardAction(cmd *cobra.Command, action, cardNumber string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/cards/%s/%s", slug, cardNumber, action), nil, nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card #%s: %s.\n", cardNumber, action)

	return nil
}

// cardActionDelete performs a DELETE to /:slug/cards/:number/:action.
func cardActionDelete(cmd *cobra.Command, action, cardNumber string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/cards/%s/%s", slug, cardNumber, action)); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Card #%s: %s removed.\n", cardNumber, action)

	return nil
}

func runCardRemoveImage(cmd *cobra.Command, args []string) error {
	confirmed, err := confirmAction(cmd, fmt.Sprintf("Remove the header image from card #%s?", args[0]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	return cardActionDelete(cmd, "image", args[0])
}

func runCardReactionList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var reactions []api.Reaction
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/cards/%s/reactions", slug, args[0]), &reactions); err != nil {
		return err
	}

	if len(reactions) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No reactions.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(w, "ID\tCONTENT\tREACTER")
	for _, r := range reactions {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", r.ID, r.Content, r.Reacter.Name)
	}

	return w.Flush()
}

func runCardReactionCreate(cmd *cobra.Command, args []string) error {
	if cardReactionBody == "" {
		return fmt.Errorf("--body is required")
	}

	if err := validateReactionContent("--body", cardReactionBody); err != nil {
		return err
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]map[string]string{
		"reaction": {"content": cardReactionBody},
	})
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/cards/%s/reactions", slug, args[0]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Reaction added to card #%s.\n", args[0])

	return nil
}

func runCardReactionDelete(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Delete reaction %s from card #%s?", args[1], args[0]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/cards/%s/reactions/%s", slug, args[0], args[1])); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Reaction removed from card #%s.\n", args[0])

	return nil
}
