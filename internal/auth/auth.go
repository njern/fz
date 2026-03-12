package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// newHTTPClient creates the shared HTTP client for the auth flow.
func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

// MagicLinkRequest initiates the magic link flow.
func MagicLinkRequest(ctx context.Context, httpClient *http.Client, host, email string) (pendingToken string, err error) {
	payload, err := json.Marshal(map[string]string{"email_address": email})
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(host, "/")+"/session", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if httpClient == nil {
		httpClient = newHTTPClient()
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("requesting magic link: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return "", fmt.Errorf("magic link request failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		PendingAuthenticationToken string `json:"pending_authentication_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return result.PendingAuthenticationToken, nil
}

// MagicLinkVerify submits the code from the magic link email.
func MagicLinkVerify(ctx context.Context, httpClient *http.Client, host, pendingToken, code string) (sessionToken string, err error) {
	payload, err := json.Marshal(map[string]string{"code": code})
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(host, "/")+"/session/magic_link", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "pending_authentication_token",
		Value: pendingToken,
	})

	if httpClient == nil {
		httpClient = newHTTPClient()
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("verifying magic link: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return "", fmt.Errorf("magic link verification failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		SessionToken string `json:"session_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return result.SessionToken, nil
}

// FetchIdentity retrieves the user's identity using a session cookie.
func FetchIdentity(ctx context.Context, httpClient *http.Client, host, sessionToken string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(host, "/")+"/my/identity", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})

	if httpClient == nil {
		httpClient = newHTTPClient()
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching identity: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("identity request failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return json.RawMessage(body), nil
}

// CreateAccessToken creates a personal access token.
// It authenticates using either a session cookie or a Bearer token, depending on
// which credential is provided. The magic-link login flow passes a session token
// (cookie auth), while fz auth create-token passes the stored Bearer PAT.
func CreateAccessToken(ctx context.Context, httpClient *http.Client, host, accountSlug, credential, description, permission string, useBearer bool) (token string, err error) {
	payload, err := json.Marshal(map[string]map[string]string{
		"access_token": {
			"description": description,
			"permission":  permission,
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/%s/my/access_tokens", strings.TrimRight(host, "/"), accountSlug),
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if useBearer {
		req.Header.Set("Authorization", "Bearer "+credential)
	} else {
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: credential,
		})
	}

	if httpClient == nil {
		httpClient = newHTTPClient()
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("creating access token: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return "", fmt.Errorf("access token creation failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return result.Token, nil
}
