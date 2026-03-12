package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/njern/fz/internal/api"
	"github.com/spf13/cobra"
)

var notificationCmd = &cobra.Command{
	Use:     "notification <command>",
	Short:   "Manage notifications",
	Aliases: []string{"notifications"},
}

var notificationListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List notifications",
	Aliases: []string{"ls"},
	RunE:    runNotificationList,
}

var notificationReadCmd = &cobra.Command{
	Use:   "read <notification-id>",
	Short: "Mark a notification as read",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotificationRead,
}

var notificationUnreadCmd = &cobra.Command{
	Use:   "unread <notification-id>",
	Short: "Mark a notification as unread",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotificationUnread,
}

var notificationReadAllCmd = &cobra.Command{
	Use:   "read-all",
	Short: "Mark all notifications as read",
	RunE:  runNotificationReadAll,
}

var notificationOnlyUnread bool

func init() {
	notificationListCmd.Flags().BoolVar(&notificationOnlyUnread, "unread", false, "Only show unread notifications")

	notificationCmd.AddCommand(notificationListCmd)
	notificationCmd.AddCommand(notificationReadCmd)
	notificationCmd.AddCommand(notificationUnreadCmd)
	notificationCmd.AddCommand(notificationReadAllCmd)
	rootCmd.AddCommand(notificationCmd)
}

func runNotificationList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var notifications []api.Notification
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/notifications", slug), &notifications); err != nil {
		return err
	}

	if notificationOnlyUnread {
		filtered := notifications[:0]
		for _, n := range notifications {
			if !n.Read {
				filtered = append(filtered, n)
			}
		}

		notifications = filtered
	}

	if len(notifications) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No notifications.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "STATUS\tCARD\tTITLE\tBODY\tAGE")

	for _, n := range notifications {
		status := "read"
		if !n.Read {
			status = "UNREAD"
		}

		age := relativeTime(n.CreatedAt)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", status, n.Card.Title, n.Title, truncate(n.Body, 30), age)
	}

	return w.Flush()
}

func runNotificationRead(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/notifications/%s/reading", slug, args[0]), nil, nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Notification marked as read.")

	return nil
}

func runNotificationUnread(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/notifications/%s/reading", slug, args[0])); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Notification marked as unread.")

	return nil
}

func runNotificationReadAll(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, "Mark all notifications as read?")
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/notifications/bulk_reading", slug), nil, nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "All notifications marked as read.")

	return nil
}
