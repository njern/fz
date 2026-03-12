package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/njern/fz/internal/api"
	"github.com/spf13/cobra"
)

var stepCmd = &cobra.Command{
	Use:     "step <command>",
	Short:   "Manage card steps (to-dos)",
	Aliases: []string{"steps"},
}

var stepListCmd = &cobra.Command{
	Use:     "list <card-number>",
	Short:   "List steps on a card",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(1),
	RunE:    runStepList,
}

var stepViewCmd = &cobra.Command{
	Use:   "view <card-number> <step-id>",
	Short: "View a step",
	Args:  cobra.ExactArgs(2),
	RunE:  runStepView,
}

var stepCreateCmd = &cobra.Command{
	Use:   "create <card-number>",
	Short: "Add a step to a card",
	Args:  cobra.ExactArgs(1),
	RunE:  runStepCreate,
}

var stepEditCmd = &cobra.Command{
	Use:   "edit <card-number> <step-id>",
	Short: "Edit a step",
	Args:  cobra.ExactArgs(2),
	RunE:  runStepEdit,
}

var stepDeleteCmd = &cobra.Command{
	Use:   "delete <card-number> <step-id>",
	Short: "Delete a step",
	Args:  cobra.ExactArgs(2),
	RunE:  runStepDelete,
}

var (
	stepContent   string
	stepCompleted bool
)

func init() {
	stepCreateCmd.Flags().StringVar(&stepContent, "content", "", "Step content (required)")
	stepCreateCmd.Flags().BoolVar(&stepCompleted, "completed", false, "Mark as completed")

	stepEditCmd.Flags().StringVar(&stepContent, "content", "", "New step content")
	stepEditCmd.Flags().BoolVar(&stepCompleted, "completed", false, "Mark as completed")

	stepCmd.AddCommand(stepListCmd)
	stepCmd.AddCommand(stepViewCmd)
	stepCmd.AddCommand(stepCreateCmd)
	stepCmd.AddCommand(stepEditCmd)
	stepCmd.AddCommand(stepDeleteCmd)
	rootCmd.AddCommand(stepCmd)
}

func runStepList(cmd *cobra.Command, args []string) error {
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

	if len(card.Steps) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No steps.")
		return nil
	}

	for _, s := range card.Steps {
		check := " "
		if s.Completed {
			check = "x"
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s  (%s)\n", check, s.Content, s.ID)
	}

	return nil
}

func runStepView(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var step api.Step
	if err := client.Get(cmd.Context(), fmt.Sprintf("/%s/cards/%s/steps/%s", slug, args[0], args[1]), &step); err != nil {
		return err
	}

	check := " "
	if step.Completed {
		check = "x"
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", check, step.Content)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\n", step.ID)

	return nil
}

func runStepCreate(cmd *cobra.Command, args []string) error {
	if stepContent == "" {
		return fmt.Errorf("--content is required")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	payload := map[string]any{
		"content": stepContent,
	}
	if stepCompleted {
		payload["completed"] = true
	}

	body, err := json.Marshal(map[string]any{"step": payload})
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/cards/%s/steps", slug, args[0]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Step added to card #%s.\n", args[0])

	return nil
}

func runStepEdit(cmd *cobra.Command, args []string) error {
	payload := map[string]any{}
	if stepContent != "" {
		payload["content"] = stepContent
	}

	if cmd.Flags().Changed("completed") {
		payload["completed"] = stepCompleted
	}

	if len(payload) == 0 {
		return fmt.Errorf("nothing to update; use --content or --completed")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]any{"step": payload})
	if err != nil {
		return err
	}

	if err := client.Put(cmd.Context(), fmt.Sprintf("/%s/cards/%s/steps/%s", slug, args[0], args[1]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Step updated.")

	return nil
}

func runStepDelete(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Delete step %s?", args[1]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/cards/%s/steps/%s", slug, args[0], args[1])); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Step deleted.")

	return nil
}
