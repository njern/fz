package cmd

import (
	"net/http"
	"strings"
	"testing"
)

func TestAPIGet(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /my/identity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(apiGetJSON))
	})
	testEnv(t, mux)

	resetFlags(t, apiCmd)

	result := executeCommand(t, "api", "/my/identity")

	if result.err != nil {
		t.Fatalf("api get: %v", result.err)
	}

	if !strings.Contains(result.stdout, "Test Board") {
		t.Errorf("output missing response body, got:\n%s", result.stdout)
	}
}

func TestAPI_AutoPrependSlug(t *testing.T) {
	var gotPath string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	testEnv(t, mux)

	resetFlags(t, apiCmd)

	result := executeCommand(t, "api", "boards")

	if result.err != nil {
		t.Fatalf("api auto-prepend: %v", result.err)
	}

	if gotPath != "/test-account/boards" {
		t.Errorf("path = %s, want /test-account/boards", gotPath)
	}
}

func TestAPI_ErrorStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	testEnv(t, mux)

	resetFlags(t, apiCmd)

	result := executeCommand(t, "api", "/test-account/boards/missing")

	if result.err == nil {
		t.Fatal("expected error for 404 response")
	}

	if !strings.Contains(result.err.Error(), "404") {
		t.Errorf("error = %v, expected 404 mention", result.err)
	}
}
