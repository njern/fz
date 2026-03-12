package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCardList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/cards", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("board_ids[]"); got != "board-1" {
			t.Errorf("board_ids[] = %q, want board-1", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cardListJSON))
	})
	testEnv(t, mux)

	resetFlags(t, cardListCmd)

	result := executeCommand(t, "card", "list", "-b", "board-1")

	if result.err != nil {
		t.Fatalf("card list: %v", result.err)
	}

	if !strings.Contains(result.stdout, "First Card") {
		t.Errorf("output missing card title, got:\n%s", result.stdout)
	}

	if !strings.Contains(result.stdout, "Second Card") {
		t.Errorf("output missing second card, got:\n%s", result.stdout)
	}
}

func TestCardList_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/cards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cardListEmptyJSON))
	})
	testEnv(t, mux)

	resetFlags(t, cardListCmd)

	result := executeCommand(t, "card", "list")

	if result.err != nil {
		t.Fatalf("card list empty: %v", result.err)
	}

	if !strings.Contains(result.stderr, "No cards found") {
		t.Errorf("expected 'No cards found', got:\n%s", result.stderr)
	}
}

func TestCardList_RejectsInvalidFilters(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "status", args: []string{"card", "list", "--status", "invalid"}, want: "--status must be one of"},
		{name: "sort", args: []string{"card", "list", "--sort", "invalid"}, want: "--sort must be one of"},
		{name: "created", args: []string{"card", "list", "--created", "invalid"}, want: "--created must be one of"},
		{name: "closed", args: []string{"card", "list", "--closed", "invalid"}, want: "--closed must be one of"},
	}

	testEnv(t, http.NewServeMux())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags(t, cardListCmd)

			result := executeCommand(t, tt.args...)
			if result.err == nil {
				t.Fatal("expected validation error")
			}

			if !strings.Contains(result.err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", result.err, tt.want)
			}
		})
	}
}

func TestCardView(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/cards/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cardViewJSON))
	})
	testEnv(t, mux)

	resetFlags(t, cardViewCmd)

	result := executeCommand(t, "card", "view", "1")

	if result.err != nil {
		t.Fatalf("card view: %v", result.err)
	}

	if !strings.Contains(result.stdout, "#1 First Card") {
		t.Errorf("output missing card header, got:\n%s", result.stdout)
	}

	if !strings.Contains(result.stdout, "Card description here") {
		t.Errorf("output missing description, got:\n%s", result.stdout)
	}

	if !strings.Contains(result.stdout, "[x] Step one") {
		t.Errorf("output missing completed step, got:\n%s", result.stdout)
	}

	if !strings.Contains(result.stdout, "[ ] Step two") {
		t.Errorf("output missing incomplete step, got:\n%s", result.stdout)
	}
}

func TestCardView_WithComments(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/cards/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cardViewJSON))
	})
	mux.HandleFunc("GET /test-account/cards/1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(commentListJSON))
	})
	testEnv(t, mux)

	resetFlags(t, cardViewCmd)

	result := executeCommand(t, "card", "view", "1", "--comments")

	if result.err != nil {
		t.Fatalf("card view --comments: %v", result.err)
	}

	if !strings.Contains(result.stdout, "This is a test comment") {
		t.Errorf("output missing comment text, got:\n%s", result.stdout)
	}
}

func TestCardCreate(t *testing.T) {
	var (
		gotPath string
		gotBody map[string]any
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /test-account/boards/board-1/cards", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"card-new","number":42,"title":"New Card","status":"new","description":"","tags":[],"closed":false,"golden":false,"last_active_at":"2025-03-01T12:00:00Z","created_at":"2025-03-01T12:00:00Z","url":"https://app.fizzy.do/test-account/cards/42","board":{"id":"board-1","name":"Test Board","all_access":true,"created_at":"2025-01-15T10:00:00Z","url":"","creator":{"id":"user-1","name":"Test User","role":"admin","active":true,"email_address":"","created_at":"2025-01-01T00:00:00Z","url":""}},"creator":{"id":"user-1","name":"Test User","role":"admin","active":true,"email_address":"","created_at":"2025-01-01T00:00:00Z","url":""},"comments_url":""}`))
	})
	testEnv(t, mux)

	resetFlags(t, cardCreateCmd)

	result := executeCommand(t, "card", "create", "-b", "board-1", "-t", "New Card")
	if result.err != nil {
		t.Fatalf("card create: %v", result.err)
	}

	if gotPath != "/test-account/boards/board-1/cards" {
		t.Errorf("path = %s", gotPath)
	}

	cardPayload, ok := gotBody["card"].(map[string]any)
	if !ok {
		t.Fatalf("missing 'card' in body")
	}

	if cardPayload["title"] != "New Card" {
		t.Errorf("body title = %v", cardPayload["title"])
	}
}

func TestCardCreate_MissingFlags(t *testing.T) {
	mux := http.NewServeMux()
	testEnv(t, mux)

	resetFlags(t, cardCreateCmd)

	result := executeCommand(t, "card", "create", "-t", "Card")
	if result.err == nil {
		t.Fatal("expected error for missing --board")
	}

	resetFlags(t, cardCreateCmd)

	result = executeCommand(t, "card", "create", "-b", "board-1")
	if result.err == nil {
		t.Fatal("expected error for missing --title")
	}
}

func TestCardReactionCreate_RejectsTooLongBody(t *testing.T) {
	testEnv(t, http.NewServeMux())

	resetFlags(t, cardReactionCreateCmd)

	result := executeCommand(t, "card", "reaction", "create", "1", "--body", strings.Repeat("a", 17))
	if result.err == nil {
		t.Fatal("expected error for too-long reaction body")
	}

	if !strings.Contains(result.err.Error(), "--body must be at most 16 characters") {
		t.Fatalf("error = %v", result.err)
	}
}

func TestCardClose(t *testing.T) {
	var gotMethod, gotPath string

	mux := http.NewServeMux()
	mux.HandleFunc("POST /test-account/cards/1/closure", func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path

		w.WriteHeader(http.StatusNoContent)
	})
	testEnv(t, mux)

	result := executeCommand(t, "card", "close", "1")
	if result.err != nil {
		t.Fatalf("card close: %v", result.err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %s, want POST", gotMethod)
	}

	if gotPath != "/test-account/cards/1/closure" {
		t.Errorf("path = %s", gotPath)
	}
}

func TestCardReopen(t *testing.T) {
	var gotMethod, gotPath string

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/cards/1/closure", func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path

		w.WriteHeader(http.StatusNoContent)
	})
	testEnv(t, mux)

	result := executeCommand(t, "card", "reopen", "1")
	if result.err != nil {
		t.Fatalf("card reopen: %v", result.err)
	}

	if gotMethod != "DELETE" {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}

	if gotPath != "/test-account/cards/1/closure" {
		t.Errorf("path = %s", gotPath)
	}
}

func TestCardRemoveImage_RequiresYesWhenNonInteractive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/cards/1/image", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("remove-image request should not be sent without confirmation")
	})
	testEnv(t, mux)

	result := executeCommand(t, "card", "remove-image", "1")
	if result.err == nil {
		t.Fatal("expected confirmation error")
	}

	if !strings.Contains(result.err.Error(), "--yes") {
		t.Fatalf("error = %v", result.err)
	}
}

func TestCardReactionDelete_RequiresYesWhenNonInteractive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/cards/1/reactions/reaction-1", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("reaction delete request should not be sent without confirmation")
	})
	testEnv(t, mux)

	result := executeCommand(t, "card", "reaction", "delete", "1", "reaction-1")
	if result.err == nil {
		t.Fatal("expected confirmation error")
	}

	if !strings.Contains(result.err.Error(), "--yes") {
		t.Fatalf("error = %v", result.err)
	}
}
