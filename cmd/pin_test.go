package cmd

import (
	"net/http"
	"strings"
	"testing"
)

func TestPinList_UsesMyPinsEndpoint(t *testing.T) {
	var gotPath string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/my/pins", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pinListJSON))
	})
	testEnv(t, mux)

	result := executeCommand(t, "pin", "list")

	if result.err != nil {
		t.Fatalf("pin list: %v", result.err)
	}

	if gotPath != "/test-account/my/pins" {
		t.Errorf("path = %s, want /test-account/my/pins", gotPath)
	}

	if !strings.Contains(result.stdout, "Pinned Card") {
		t.Errorf("output missing pin title, got:\n%s", result.stdout)
	}
}
