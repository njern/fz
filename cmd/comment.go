package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/njern/fz/internal/api"
	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:     "comment <command>",
	Short:   "Manage card comments",
	Aliases: []string{"comments"},
}

var commentListCmd = &cobra.Command{
	Use:     "list <card-number>",
	Short:   "List comments on a card",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(1),
	RunE:    runCommentList,
}

var commentViewCmd = &cobra.Command{
	Use:   "view <card-number> <comment-id>",
	Short: "View a comment",
	Args:  cobra.ExactArgs(2),
	RunE:  runCommentView,
}

var commentCreateCmd = &cobra.Command{
	Use:   "create <card-number>",
	Short: "Add a comment to a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runCommentCreate,
}

var commentEditCmd = &cobra.Command{
	Use:   "edit <card-number> <comment-id>",
	Short: "Edit a comment",
	Args:  cobra.ExactArgs(2),
	RunE:  runCommentEdit,
}

var commentDeleteCmd = &cobra.Command{
	Use:   "delete <card-number> <comment-id>",
	Short: "Delete a comment",
	Args:  cobra.ExactArgs(2),
	RunE:  runCommentDelete,
}

var commentReactionCmd = &cobra.Command{
	Use:   "reaction <command>",
	Short: "Manage comment reactions",
}

var commentReactionListCmd = &cobra.Command{
	Use:     "list <card-number> <comment-id>",
	Short:   "List reactions on a comment",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(2),
	RunE:    runCommentReactionList,
}

var commentReactionCreateCmd = &cobra.Command{
	Use:   "create <card-number> <comment-id>",
	Short: "Add a reaction to a comment",
	Args:  cobra.ExactArgs(2),
	RunE:  runCommentReactionCreate,
}

var commentReactionDeleteCmd = &cobra.Command{
	Use:   "delete <card-number> <comment-id> <reaction-id>",
	Short: "Remove a reaction from a comment",
	Args:  cobra.ExactArgs(3),
	RunE:  runCommentReactionDelete,
}

var (
	commentBody         string
	commentReactionBody string
)

func init() {
	commentCreateCmd.Flags().StringVarP(&commentBody, "body", "B", "", "Comment body (required)")
	commentEditCmd.Flags().StringVarP(&commentBody, "body", "B", "", "New comment body (required)")
	commentReactionCreateCmd.Flags().StringVarP(&commentReactionBody, "body", "B", "", "Reaction text (max 16 chars)")

	commentReactionCmd.AddCommand(commentReactionListCmd)
	commentReactionCmd.AddCommand(commentReactionCreateCmd)
	commentReactionCmd.AddCommand(commentReactionDeleteCmd)

	commentCmd.AddCommand(commentListCmd)
	commentCmd.AddCommand(commentViewCmd)
	commentCmd.AddCommand(commentCreateCmd)
	commentCmd.AddCommand(commentEditCmd)
	commentCmd.AddCommand(commentDeleteCmd)
	commentCmd.AddCommand(commentReactionCmd)
	rootCmd.AddCommand(commentCmd)
}

func runCommentList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var comments []api.Comment
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/cards/%s/comments", slug, args[0]), &comments); err != nil {
		return err
	}

	if len(comments) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No comments.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tCREATOR\tCREATED\tBODY")

	for _, c := range comments {
		body := c.Body.PlainText
		if runes := []rune(body); len(runes) > 60 {
			body = string(runes[:59]) + "…"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.ID, c.Creator.Name, c.CreatedAt.Format("2006-01-02 15:04"), body)
	}

	return w.Flush()
}

func runCommentView(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var comment api.Comment
	if err := client.Get(cmd.Context(), fmt.Sprintf("/%s/cards/%s/comments/%s", slug, args[0], args[1]), &comment); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Comment %s\n", comment.ID)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Creator: %s\n", comment.Creator.Name)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", comment.CreatedAt.Format("2006-01-02 15:04"))
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated: %s\n", comment.UpdatedAt.Format("2006-01-02 15:04"))
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", comment.Body.PlainText)

	return nil
}

func runCommentCreate(cmd *cobra.Command, args []string) error {
	if commentBody == "" {
		return fmt.Errorf("--body is required")
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
		"comment": {"body": commentBody},
	})
	if err != nil {
		return err
	}

	var comment api.Comment
	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/cards/%s/comments", slug, args[0]), bytes.NewReader(body), &comment); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Comment added to card #%s.\n", args[0])
	if comment.ID != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "ID: %s\n", comment.ID)
	}

	return nil
}

func runCommentEdit(cmd *cobra.Command, args []string) error {
	if commentBody == "" {
		return fmt.Errorf("--body is required")
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
		"comment": {"body": commentBody},
	})
	if err != nil {
		return err
	}

	if err := client.Put(cmd.Context(), fmt.Sprintf("/%s/cards/%s/comments/%s", slug, args[0], args[1]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Comment updated.")

	return nil
}

func runCommentDelete(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Delete comment %s?", args[1]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/cards/%s/comments/%s", slug, args[0], args[1])); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Comment deleted.")

	return nil
}

func runCommentReactionList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var reactions []api.Reaction
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/cards/%s/comments/%s/reactions", slug, args[0], args[1]), &reactions); err != nil {
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

func runCommentReactionCreate(cmd *cobra.Command, args []string) error {
	if commentReactionBody == "" {
		return fmt.Errorf("--body is required")
	}

	if err := validateReactionContent("--body", commentReactionBody); err != nil {
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
		"reaction": {"content": commentReactionBody},
	})
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/cards/%s/comments/%s/reactions", slug, args[0], args[1]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Reaction added to comment.")

	return nil
}

func runCommentReactionDelete(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Delete reaction %s from comment %s?", args[2], args[1]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/cards/%s/comments/%s/reactions/%s", slug, args[0], args[1], args[2])); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Reaction removed from comment.")

	return nil
}
