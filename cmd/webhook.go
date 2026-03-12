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

var webhookCmd = &cobra.Command{
	Use:     "webhook <command>",
	Short:   "Manage board webhooks",
	Aliases: []string{"webhooks"},
}

var webhookListCmd = &cobra.Command{
	Use:     "list <board-id>",
	Short:   "List webhooks on a board",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(1),
	RunE:    runWebhookList,
}

var webhookViewCmd = &cobra.Command{
	Use:   "view <board-id> <webhook-id>",
	Short: "View a webhook",
	Args:  cobra.ExactArgs(2),
	RunE:  runWebhookView,
}

var webhookCreateCmd = &cobra.Command{
	Use:   "create <board-id>",
	Short: "Create a webhook",
	Args:  cobra.ExactArgs(1),
	RunE:  runWebhookCreate,
}

var webhookEditCmd = &cobra.Command{
	Use:   "edit <board-id> <webhook-id>",
	Short: "Update a webhook",
	Args:  cobra.ExactArgs(2),
	RunE:  runWebhookEdit,
}

var webhookDeleteCmd = &cobra.Command{
	Use:   "delete <board-id> <webhook-id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(2),
	RunE:  runWebhookDelete,
}

var webhookActivateCmd = &cobra.Command{
	Use:   "activate <board-id> <webhook-id>",
	Short: "Reactivate a deactivated webhook",
	Args:  cobra.ExactArgs(2),
	RunE:  runWebhookActivate,
}

var (
	webhookName       string
	webhookURL        string
	webhookEvents     string
	webhookShowSecret bool
)

type webhookRequest struct {
	Webhook webhookPayload `json:"webhook"`
}

type webhookPayload struct {
	Name              string   `json:"name,omitempty"`
	URL               string   `json:"url,omitempty"`
	SubscribedActions []string `json:"subscribed_actions,omitempty"`
}

func (p webhookPayload) hasUpdates() bool {
	return p.Name != "" || p.URL != "" || len(p.SubscribedActions) > 0
}

func init() {
	webhookCreateCmd.Flags().StringVar(&webhookName, "name", "", "Webhook name (required)")
	webhookCreateCmd.Flags().StringVar(&webhookURL, "url", "", "Payload URL (required)")
	webhookCreateCmd.Flags().StringVar(&webhookEvents, "events", "", "Comma-separated subscribed actions")

	webhookViewCmd.Flags().BoolVar(&webhookShowSecret, "show-secret", false, "Show the signing secret in plaintext")

	webhookEditCmd.Flags().StringVar(&webhookName, "name", "", "New webhook name")
	webhookEditCmd.Flags().StringVar(&webhookEvents, "events", "", "Comma-separated subscribed actions")

	webhookCmd.AddCommand(webhookListCmd)
	webhookCmd.AddCommand(webhookViewCmd)
	webhookCmd.AddCommand(webhookCreateCmd)
	webhookCmd.AddCommand(webhookEditCmd)
	webhookCmd.AddCommand(webhookDeleteCmd)
	webhookCmd.AddCommand(webhookActivateCmd)
	rootCmd.AddCommand(webhookCmd)
}

func runWebhookList(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	var webhooks []api.Webhook
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/boards/%s/webhooks", slug, args[0]), &webhooks); err != nil {
		return err
	}

	if len(webhooks) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No webhooks found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tNAME\tURL\tACTIVE\tEVENTS")

	for _, wh := range webhooks {
		events := strings.Join(wh.SubscribedActions, ", ")
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\n", wh.ID, wh.Name, wh.PayloadURL, wh.Active, events)
	}

	return w.Flush()
}

func runWebhookView(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	var wh api.Webhook
	if err := client.Get(cmd.Context(), fmt.Sprintf("/%s/boards/%s/webhooks/%s", slug, args[0], args[1]), &wh); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", wh.Name)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:      %s\n", wh.ID)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "URL:     %s\n", wh.PayloadURL)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Active:  %v\n", wh.Active)

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Events:  %s\n", strings.Join(wh.SubscribedActions, ", "))
	if webhookShowSecret {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Secret:  %s\n", wh.SigningSecret)
	} else {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Secret:  ****")
	}

	return nil
}

func runWebhookCreate(cmd *cobra.Command, args []string) error {
	if webhookName == "" {
		return fmt.Errorf("--name is required")
	}

	if webhookURL == "" {
		return fmt.Errorf("--url is required")
	}

	events, err := parseWebhookEvents(webhookEvents)
	if err != nil {
		return err
	}

	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	payload := webhookPayload{
		Name: webhookName,
		URL:  webhookURL,
	}
	if len(events) > 0 {
		payload.SubscribedActions = events
	}

	body, err := json.Marshal(webhookRequest{Webhook: payload})
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/boards/%s/webhooks", slug, args[0]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Webhook %q created.\n", webhookName)

	return nil
}

func runWebhookEdit(cmd *cobra.Command, args []string) error {
	events, err := parseWebhookEvents(webhookEvents)
	if err != nil {
		return err
	}

	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	payload := webhookPayload{}
	if webhookName != "" {
		payload.Name = webhookName
	}

	if len(events) > 0 {
		payload.SubscribedActions = events
	}

	if !payload.hasUpdates() {
		return fmt.Errorf("nothing to update; use --name or --events")
	}

	body, err := json.Marshal(webhookRequest{Webhook: payload})
	if err != nil {
		return err
	}

	if err := client.Patch(cmd.Context(), fmt.Sprintf("/%s/boards/%s/webhooks/%s", slug, args[0], args[1]), bytes.NewReader(body), nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Webhook updated.")

	return nil
}

func runWebhookDelete(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Delete webhook %s?", args[1]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/boards/%s/webhooks/%s", slug, args[0], args[1])); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Webhook deleted.")

	return nil
}

func runWebhookActivate(cmd *cobra.Command, args []string) error {
	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/boards/%s/webhooks/%s/activation", slug, args[0], args[1]), nil, nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Webhook reactivated.")

	return nil
}
