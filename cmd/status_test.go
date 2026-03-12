package cmd

import (
	"net/http"
	"strings"
	"testing"
)

func TestStatus_ShowsPinnedCardsFromMyPinsEndpoint(t *testing.T) {
	var pinsPath string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /my/identity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(identityJSON))
	})
	mux.HandleFunc("GET /test-account/notifications", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(notificationsEmptyJSON))
	})
	mux.HandleFunc("GET /test-account/my/pins", func(w http.ResponseWriter, r *http.Request) {
		pinsPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pinListJSON))
	})
	mux.HandleFunc("GET /test-account/cards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	})
	testEnv(t, mux)

	result := executeCommand(t, "status")

	if result.err != nil {
		t.Fatalf("status: %v", result.err)
	}

	if pinsPath != "/test-account/my/pins" {
		t.Errorf("pins path = %s, want /test-account/my/pins", pinsPath)
	}

	if !strings.Contains(result.stdout, "Pinned Cards") {
		t.Errorf("output missing 'Pinned Cards' section, got:\n%s", result.stdout)
	}

	if !strings.Contains(result.stdout, "Pinned Card") {
		t.Errorf("output missing pin title, got:\n%s", result.stdout)
	}
}

func TestStatus_ShowsNothingNewWhenAllNotificationsAreRead(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /my/identity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(identityJSON))
	})
	mux.HandleFunc("GET /test-account/notifications", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(notificationsReadJSON))
	})
	mux.HandleFunc("GET /test-account/my/pins", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	})
	mux.HandleFunc("GET /test-account/cards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	})
	testEnv(t, mux)

	result := executeCommand(t, "status")

	if result.err != nil {
		t.Fatalf("status: %v", result.err)
	}

	if !strings.Contains(result.stdout, "Notifications") {
		t.Errorf("output missing notifications header, got:\n%s", result.stdout)
	}

	if !strings.Contains(result.stdout, "Nothing new.") {
		t.Errorf("output missing 'Nothing new.', got:\n%s", result.stdout)
	}
}
