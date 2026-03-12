package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAccessToken_AuthModes(t *testing.T) {
	tests := []struct {
		name      string
		useBearer bool
	}{
		{name: "session cookie", useBearer: false},
		{name: "bearer token", useBearer: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotPath        string
				gotCookie      string
				gotAuthHeader  string
				gotContentType string
				gotBody        map[string]map[string]string
			)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotContentType = r.Header.Get("Content-Type")

				if c, err := r.Cookie("session_token"); err == nil {
					gotCookie = c.Value
				}

				gotAuthHeader = r.Header.Get("Authorization")

				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &gotBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"token":"new-pat-123"}`))
			}))
			t.Cleanup(srv.Close)

			token, err := CreateAccessToken(context.Background(), nil, srv.URL, "my-org", "cred-value", "test token", "write", tt.useBearer)
			if err != nil {
				t.Fatalf("CreateAccessToken: %v", err)
			}

			if token != "new-pat-123" {
				t.Errorf("token = %q, want %q", token, "new-pat-123")
			}

			if gotPath != "/my-org/my/access_tokens" {
				t.Errorf("path = %s, want /my-org/my/access_tokens", gotPath)
			}

			if gotContentType != "application/json" {
				t.Errorf("content-type = %s, want application/json", gotContentType)
			}

			// Verify payload.
			at := gotBody["access_token"]
			if at["description"] != "test token" {
				t.Errorf("description = %q, want %q", at["description"], "test token")
			}

			if at["permission"] != "write" {
				t.Errorf("permission = %q, want %q", at["permission"], "write")
			}

			// Verify auth mechanism.
			if tt.useBearer {
				if gotAuthHeader != "Bearer cred-value" {
					t.Errorf("Authorization = %q, want %q", gotAuthHeader, "Bearer cred-value")
				}

				if gotCookie != "" {
					t.Error("unexpected session_token cookie with Bearer auth")
				}
			} else {
				if gotCookie != "cred-value" {
					t.Errorf("session_token cookie = %q, want %q", gotCookie, "cred-value")
				}

				if gotAuthHeader != "" {
					t.Error("unexpected Authorization header with cookie auth")
				}
			}
		})
	}
}
