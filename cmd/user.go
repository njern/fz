package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/njern/fz/internal/api"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:     "user <command>",
	Short:   "Manage account users",
	Aliases: []string{"users"},
}

var userListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List active users",
	Aliases: []string{"ls"},
	RunE:    runUserList,
}

var userViewCmd = &cobra.Command{
	Use:   "view <user-id>",
	Short: "View a user",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserView,
}

var userEditCmd = &cobra.Command{
	Use:   "edit <user-id>",
	Short: "Edit a user",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserEdit,
}

var userDeactivateCmd = &cobra.Command{
	Use:   "deactivate <user-id>",
	Short: "Deactivate a user",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserDeactivate,
}

var (
	userEditName   string
	userEditAvatar string
)

func init() {
	userEditCmd.Flags().StringVar(&userEditName, "name", "", "New display name")
	userEditCmd.Flags().StringVar(&userEditAvatar, "avatar", "", "Path to avatar image file")

	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userViewCmd)
	userCmd.AddCommand(userEditCmd)
	userCmd.AddCommand(userDeactivateCmd)
	rootCmd.AddCommand(userCmd)
}

func runUserList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var users []api.User
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/users", slug), &users); err != nil {
		return err
	}

	if len(users) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No users found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(w, "ID\tNAME\tROLE\tEMAIL")
	for _, u := range users {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", u.ID, u.Name, u.Role, u.EmailAddress)
	}

	return w.Flush()
}

func runUserView(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	var user api.User
	if err := client.Get(cmd.Context(), fmt.Sprintf("/%s/users/%s", slug, args[0]), &user); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", user.Name)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:    %s\n", user.ID)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Role:  %s\n", user.Role)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Email: %s\n", user.EmailAddress)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "URL:   %s\n", user.URL)

	return nil
}

func runUserEdit(cmd *cobra.Command, args []string) error {
	if userEditName == "" && userEditAvatar == "" {
		return fmt.Errorf("at least one of --name or --avatar is required")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/%s/users/%s", slug, args[0])

	if userEditAvatar != "" {
		if err := putUserMultipart(cmd.Context(), client, path); err != nil {
			return err
		}
	} else {
		body, err := json.Marshal(map[string]map[string]string{
			"user": {"name": userEditName},
		})
		if err != nil {
			return err
		}

		if err := client.Put(cmd.Context(), path, bytes.NewReader(body), nil); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "User updated.")

	return nil
}

func putUserMultipart(ctx context.Context, client *api.Client, path string) error {
	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	if userEditName != "" {
		if err := w.WriteField("user[name]", userEditName); err != nil {
			return fmt.Errorf("writing name field: %w", err)
		}
	}

	f, err := os.Open(userEditAvatar)
	if err != nil {
		return fmt.Errorf("opening avatar file: %w", err)
	}

	defer func() { _ = f.Close() }()

	part, err := w.CreateFormFile("user[avatar]", filepath.Base(userEditAvatar))
	if err != nil {
		return fmt.Errorf("creating avatar form file: %w", err)
	}

	if _, err := io.Copy(part, f); err != nil {
		return fmt.Errorf("writing avatar data: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("closing multipart writer: %w", err)
	}

	return client.PutMultipart(ctx, path, w.FormDataContentType(), &buf, nil)
}

func runUserDeactivate(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	slug, err := accountSlug()
	if err != nil {
		return err
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Deactivate user %s?", args[0]))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	if err := client.Delete(cmd.Context(), fmt.Sprintf("/%s/users/%s", slug, args[0])); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "User deactivated.")

	return nil
}
