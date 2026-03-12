package cmd

import (
	"net/http"
	"strings"
	"testing"
)

func TestStepDelete(t *testing.T) {
	var gotPath string

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/cards/1/steps/step-1", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		w.WriteHeader(http.StatusNoContent)
	})
	testEnv(t, mux)

	result := executeCommand(t, "step", "delete", "1", "step-1", "--yes")
	if result.err != nil {
		t.Fatalf("step delete: %v", result.err)
	}

	if gotPath != "/test-account/cards/1/steps/step-1" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestStepDelete_RequiresYesWhenNonInteractive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/cards/1/steps/step-1", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("delete request should not be sent without confirmation")
	})
	testEnv(t, mux)

	result := executeCommand(t, "step", "delete", "1", "step-1")
	if result.err == nil {
		t.Fatal("expected confirmation error")
	}

	if !strings.Contains(result.err.Error(), "--yes") {
		t.Fatalf("error = %v", result.err)
	}
}
