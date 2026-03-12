package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// maxPages is the maximum number of pages GetAll will follow before returning
// an error. This prevents infinite loops from circular Link headers.
const maxPages = 200

// maxErrorBody is the maximum number of bytes read from error response bodies.
const maxErrorBody = 1 << 20 // 1 MiB

// ErrTooManyPages is returned when GetAll exceeds the maximum page limit.
var ErrTooManyPages = errors.New("pagination limit exceeded")

// Client is an authenticated HTTP client for the Fizzy API.
type Client struct {
	httpClient *http.Client
	host       string
	token      string
	userAgent  string
}

// NewClient creates a new API client.
func NewClient(host, token, version string) *Client {
	ua := "fz-cli"
	if version != "" {
		ua += "/" + version
	}

	return &Client{
		host:  strings.TrimRight(host, "/"),
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: ua,
	}
}

// Request performs an authenticated HTTP request and returns the raw response.
func (c *Client) Request(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	return c.requestWithContentType(ctx, method, path, body, "")
}

// requestWithContentType builds and executes a request, optionally overriding
// the Content-Type header. If contentType is empty and body is non-nil,
// "application/json" is used.
func (c *Client) requestWithContentType(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
	url := c.host + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

// Get performs a GET request and decodes the JSON response into v.
func (c *Client) Get(ctx context.Context, path string, v any) error {
	resp, err := c.Request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if err := checkStatus(resp); err != nil {
		return err
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// GetAll performs paginated GET requests, following Link rel="next" headers,
// and collects all pages into v (which must be a pointer to a slice).
func (c *Client) GetAll(ctx context.Context, path string, v any) error {
	var all []json.RawMessage

	for page := 0; path != ""; page++ {
		if page >= maxPages {
			return fmt.Errorf("%w: stopped after %d pages", ErrTooManyPages, maxPages)
		}

		resp, err := c.Request(ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}

		if err := checkStatus(resp); err != nil {
			_ = resp.Body.Close()
			return err
		}

		var items []json.RawMessage

		err = json.NewDecoder(resp.Body).Decode(&items)
		_ = resp.Body.Close()

		if err != nil {
			return err
		}

		all = append(all, items...)
		path = nextPagePath(resp.Header.Get("Link"), c.host)
	}

	combined, err := json.Marshal(all)
	if err != nil {
		return err
	}

	return json.Unmarshal(combined, v)
}

// linkNextRe matches Link: <url>; rel="next" headers.
var linkNextRe = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

// nextPagePath extracts the path (relative to host) from a Link header's rel="next" URL.
// Returns "" if there is no next page.
func nextPagePath(linkHeader, host string) string {
	m := linkNextRe.FindStringSubmatch(linkHeader)
	if m == nil {
		return ""
	}

	url := m[1]
	if strings.HasPrefix(url, host) {
		return url[len(host):]
	}
	// Already a relative path.
	if strings.HasPrefix(url, "/") {
		return url
	}

	return ""
}

// Post performs a POST request with a JSON body and decodes the response into v.
// Pass nil for v if no response body is expected.
func (c *Client) Post(ctx context.Context, path string, body io.Reader, v any) error {
	resp, err := c.Request(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if err := checkStatus(resp); err != nil {
		return err
	}

	if v != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

// Put performs a PUT request with a JSON body and decodes the response into v.
// Pass nil for v if no response body is expected.
func (c *Client) Put(ctx context.Context, path string, body io.Reader, v any) error {
	resp, err := c.Request(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if err := checkStatus(resp); err != nil {
		return err
	}

	if v != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

// PutMultipart performs a PUT request with a custom content type (e.g. multipart/form-data)
// and decodes the JSON response into v. Pass nil for v if no response body is expected.
func (c *Client) PutMultipart(ctx context.Context, path, contentType string, body io.Reader, v any) error {
	resp, err := c.requestWithContentType(ctx, http.MethodPut, path, body, contentType)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if err := checkStatus(resp); err != nil {
		return err
	}

	if v != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	resp, err := c.Request(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	return checkStatus(resp)
}

// Patch performs a PATCH request with a JSON body and decodes the response into v.
func (c *Client) Patch(ctx context.Context, path string, body io.Reader, v any) error {
	resp, err := c.Request(ctx, http.MethodPatch, path, body)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if err := checkStatus(resp); err != nil {
		return err
	}

	if v != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

// APIError represents an error response from the Fizzy API.
type APIError struct {
	Body       string
	StatusCode int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Body)
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))

	return &APIError{
		StatusCode: resp.StatusCode,
		Body:       strings.TrimSpace(string(body)),
	}
}
