package cmd

import (
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/njern/fz/internal/api"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show your notifications and assigned cards",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	// Fetch notifications, pinned cards, and identity in parallel would be nice,
	// but for simplicity we do them sequentially.
	var identity api.Identity
	if err := client.Get(ctx, "/my/identity", &identity); err != nil {
		return err
	}

	// Find the current account and user.
	acct := findAccount(identity.Accounts, slug)

	accountName := ""

	var currentUser *api.User

	if acct != nil {
		accountName = acct.Name
		currentUser = &acct.User
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Account: %s (/%s)\n\n", accountName, slug)

	// Notifications.
	var notifications []api.Notification
	if err := client.GetAll(ctx, fmt.Sprintf("/%s/notifications", slug), &notifications); err != nil {
		return err
	}

	unread := 0

	for _, n := range notifications {
		if !n.Read {
			unread++
		}
	}

	if unread > 0 {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Notifications (%d unread)\n", unread)
	} else {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Notifications")
	}

	printedUnread := false

	for _, n := range notifications {
		if n.Read {
			continue
		}

		printedUnread = true
		age := relativeTime(n.CreatedAt)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  #%-6s %-30s %-20s %s\n", cardNumberFromURL(n.Card.URL), truncate(n.Card.Title, 30), truncate(n.Body, 20), age)
	}

	if !printedUnread {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  Nothing new.")
	}

	// Pinned cards.
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	var pins []api.Card
	if err := client.GetAll(ctx, fmt.Sprintf("/%s/my/pins", slug), &pins); err != nil {
		return fmt.Errorf("fetching pinned cards: %w", err)
	}

	if len(pins) > 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Pinned Cards")
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

		for _, c := range pins {
			col := ""
			if c.Column != nil {
				col = c.Column.Name
			}

			_, _ = fmt.Fprintf(w, "  #%d\t%s\t%s\n", c.Number, c.Title, col)
		}

		if err := w.Flush(); err != nil {
			return err
		}
	}

	// Assigned cards (if we know the user).
	if currentUser != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout())

		var assigned []api.Card

		params := url.Values{}
		params.Add("assignee_ids[]", currentUser.ID)

		path := fmt.Sprintf("/%s/cards?%s", slug, params.Encode())
		if err := client.GetAll(ctx, path, &assigned); err != nil {
			return fmt.Errorf("fetching assigned cards: %w", err)
		}

		if len(assigned) > 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Assigned to You")
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

			for _, c := range assigned {
				col := ""
				if c.Column != nil {
					col = c.Column.Name
				}

				_, _ = fmt.Fprintf(w, "  #%d\t%s\t%s\t%s\n", c.Number, c.Title, c.Board.Name, col)
			}

			if err := w.Flush(); err != nil {
				return err
			}
		}
	}

	return nil
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}

	return string(runes[:n-1]) + "…"
}

func findAccount(accounts []api.Account, slug string) *api.Account {
	for i, acct := range accounts {
		if acct.SlugTrimmed() == slug {
			return &accounts[i]
		}
	}

	return nil
}

func cardNumberFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "?"
}
