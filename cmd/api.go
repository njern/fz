package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api <endpoint>",
	Short: "Make an authenticated Fizzy API request",
	Long: `Make an authenticated HTTP request to the Fizzy API and print the response.

The endpoint argument should be a path like /my/identity or /:account_slug/cards.
If the path starts with /, it's used as-is. Otherwise, the active account slug is prepended.

The default HTTP method is GET. Override with --method.`,
	Args: cobra.ExactArgs(1),
	RunE: runAPI,
}

var (
	apiMethod string
	apiInput  string
)

func init() {
	apiCmd.Flags().StringVarP(&apiMethod, "method", "X", "", "HTTP method (default: GET, or POST if --input is used)")
	apiCmd.Flags().StringVar(&apiInput, "input", "", "File to use as request body (use - for stdin)")

	rootCmd.AddCommand(apiCmd)
}

func runAPI(cmd *cobra.Command, args []string) (retErr error) {
	client, err := newClient()
	if err != nil {
		return err
	}

	endpoint := args[0]

	// If the path doesn't start with /, prepend the account slug.
	if !strings.HasPrefix(endpoint, "/") {
		slug, err := accountSlug()
		if err != nil {
			return err
		}

		endpoint = "/" + slug + "/" + endpoint
	}

	var body io.Reader

	if apiInput != "" {
		if apiInput == "-" {
			body = cmd.InOrStdin()
		} else {
			f, err := os.Open(apiInput)
			if err != nil {
				return fmt.Errorf("opening input file: %w", err)
			}

			defer func() {
				if cErr := f.Close(); retErr == nil {
					retErr = cErr
				}
			}()

			body = f
		}
	}

	method := apiMethod
	if method == "" {
		if body != nil {
			method = http.MethodPost
		} else {
			method = http.MethodGet
		}
	}

	resp, err := client.Request(cmd.Context(), strings.ToUpper(method), endpoint, body)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	// Print status to stderr.
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s %s\n", resp.Proto, resp.Status)

	// Print body to stdout.
	if _, err := io.Copy(cmd.OutOrStdout(), resp.Body); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil
}
