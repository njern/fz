package cmd

import (
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoginWithToken_AcceptsEOFWithoutNewline(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /my/identity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(identityJSON))
	})
	testEnv(t, mux)
	cfg.SetPath(filepath.Join(t.TempDir(), "config.json"))

	// Token with no trailing newline, simulating: printf %s "$TOKEN" | fz auth login --with-token
	r := strings.NewReader("test-pat-token")

	err := loginWithTokenFromReader(context.Background(), r, io.Discard, false)
	if err != nil {
		t.Fatalf("loginWithTokenFromReader: %v", err)
	}

	if cfg.Token != "test-pat-token" {
		t.Errorf("cfg.Token = %q, want %q", cfg.Token, "test-pat-token")
	}

	if cfg.DefaultAccount != "test-account" {
		t.Errorf("cfg.DefaultAccount = %q, want %q", cfg.DefaultAccount, "test-account")
	}
}

func TestLoginWithToken_RequiresAccountWhenMultipleAccountsAndNonInteractive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /my/identity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(multiAccountIdentityJSON))
	})
	testEnv(t, mux)
	cfg.SetPath(filepath.Join(t.TempDir(), "config.json"))

	err := loginWithTokenFromReader(context.Background(), strings.NewReader("test-pat-token\n"), io.Discard, false)
	if err == nil {
		t.Fatal("expected error for multi-account non-interactive login")
	}

	if !strings.Contains(err.Error(), "--account") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoginWithToken_UsesAccountOverride(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /my/identity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(multiAccountIdentityJSON))
	})
	testEnv(t, mux)
	cfg.SetPath(filepath.Join(t.TempDir(), "config.json"))

	cfgOverrideAccount = "other-account"

	err := loginWithTokenFromReader(context.Background(), strings.NewReader("test-pat-token\n"), io.Discard, false)
	if err != nil {
		t.Fatalf("loginWithTokenFromReader: %v", err)
	}

	if cfg.DefaultAccount != "other-account" {
		t.Fatalf("cfg.DefaultAccount = %q, want %q", cfg.DefaultAccount, "other-account")
	}
}

func TestAuthCreateToken_UsesBearerToken(t *testing.T) {
	var gotAuthHeader string

	mux := http.NewServeMux()
	mux.HandleFunc("POST /test-account/my/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"token":"new-token-abc"}`))
	})
	testEnv(t, mux)

	resetFlags(t, authCreateTokenCmd)

	result := executeCommand(t, "auth", "create-token", "--description", "my token", "--permission", "write")

	if result.err != nil {
		t.Fatalf("auth create-token: %v", result.err)
	}

	if gotAuthHeader != "Bearer test-token" {
		t.Errorf("Authorization = %q, want %q", gotAuthHeader, "Bearer test-token")
	}

	if !strings.Contains(result.stdout, "new-token-abc") {
		t.Errorf("output missing new token, got:\n%s", result.stdout)
	}
}

func TestAuthStatusCheck_ReturnsErrorWhenNotLoggedIn(t *testing.T) {
	testEnv(t, http.NewServeMux())

	cfg.Token = ""

	resetFlags(t, authStatusCmd)

	result := executeCommand(t, "auth", "status", "--check")

	if !errors.Is(result.err, ErrNotAuthenticated) {
		t.Fatalf("error = %v, want %v", result.err, ErrNotAuthenticated)
	}

	if !strings.Contains(result.stderr, "Not logged in.") {
		t.Fatalf("stderr = %q", result.stderr)
	}
}

func TestAuthStatusCheck_ReturnsErrorWhenTokenIsInvalid(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /my/identity", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	})
	testEnv(t, mux)

	resetFlags(t, authStatusCmd)

	result := executeCommand(t, "auth", "status", "--check")

	if !errors.Is(result.err, ErrInvalidToken) {
		t.Fatalf("error = %v, want %v", result.err, ErrInvalidToken)
	}

	if !strings.Contains(result.stderr, "Token is invalid or expired") {
		t.Fatalf("stderr = %q", result.stderr)
	}
}

func TestAuthStatusCheck_ReturnsTransportErrorWhenVerificationFails(t *testing.T) {
	srv := testEnv(t, http.NewServeMux())
	srv.Close()

	resetFlags(t, authStatusCmd)

	result := executeCommand(t, "auth", "status", "--check")

	if result.err == nil {
		t.Fatal("expected transport error")
	}

	if errors.Is(result.err, ErrInvalidToken) {
		t.Fatalf("error = %v, should not map transport failures to %v", result.err, ErrInvalidToken)
	}

	if strings.Contains(result.stderr, "Token is invalid or expired") {
		t.Fatalf("stderr = %q", result.stderr)
	}

	if !strings.Contains(result.stderr, "Authentication check failed") {
		t.Fatalf("stderr = %q", result.stderr)
	}
}

func TestAuthLogout_RequiresYesWhenNonInteractive(t *testing.T) {
	testEnv(t, http.NewServeMux())

	result := executeCommand(t, "auth", "logout")
	if !errors.Is(result.err, ErrConfirmation) {
		t.Fatalf("error = %v, want %v", result.err, ErrConfirmation)
	}
}

const multiAccountIdentityJSON = `{
  "accounts": [
    {
      "id": "acct-1",
      "name": "Test Org",
      "slug": "/test-account",
      "created_at": "2025-01-01T00:00:00Z",
      "user": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "test@example.com", "created_at": "2025-01-01T00:00:00Z", "url": ""}
    },
    {
      "id": "acct-2",
      "name": "Other Org",
      "slug": "/other-account",
      "created_at": "2025-01-01T00:00:00Z",
      "user": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "test@example.com", "created_at": "2025-01-01T00:00:00Z", "url": ""}
    }
  ]
}`
