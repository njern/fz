package cmd

import (
	"net/http"
	"strings"
	"testing"
)

func TestNotificationReadAll(t *testing.T) {
	var gotPath string

	mux := http.NewServeMux()
	mux.HandleFunc("POST /test-account/notifications/bulk_reading", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		w.WriteHeader(http.StatusNoContent)
	})
	testEnv(t, mux)

	result := executeCommand(t, "notification", "read-all", "--yes")
	if result.err != nil {
		t.Fatalf("notification read-all: %v", result.err)
	}

	if gotPath != "/test-account/notifications/bulk_reading" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestNotificationReadAll_RequiresYesWhenNonInteractive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /test-account/notifications/bulk_reading", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("bulk_reading request should not be sent without confirmation")
	})
	testEnv(t, mux)

	result := executeCommand(t, "notification", "read-all")
	if result.err == nil {
		t.Fatal("expected confirmation error")
	}

	if !strings.Contains(result.err.Error(), "--yes") {
		t.Fatalf("error = %v", result.err)
	}
}
