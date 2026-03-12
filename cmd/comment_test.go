package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCommentList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/cards/1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(commentListJSON))
	})
	testEnv(t, mux)

	result := executeCommand(t, "comment", "list", "1")

	if result.err != nil {
		t.Fatalf("comment list: %v", result.err)
	}

	if !strings.Contains(result.stdout, "comment-1") {
		t.Errorf("output missing comment ID, got:\n%s", result.stdout)
	}

	if !strings.Contains(result.stdout, "Test User") {
		t.Errorf("output missing creator, got:\n%s", result.stdout)
	}
}

func TestCommentList_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/cards/1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(commentListEmptyJSON))
	})
	testEnv(t, mux)

	result := executeCommand(t, "comment", "list", "1")

	if result.err != nil {
		t.Fatalf("comment list empty: %v", result.err)
	}

	if !strings.Contains(result.stderr, "No comments") {
		t.Errorf("expected 'No comments', got:\n%s", result.stderr)
	}
}

func TestCommentCreate(t *testing.T) {
	var (
		gotPath string
		gotBody map[string]any
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /test-account/cards/1/comments", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"comment-new","created_at":"2025-02-10T09:00:00Z","updated_at":"2025-02-10T09:00:00Z","body":{"plain_text":"Test comment","html":"<p>Test comment</p>"},"creator":{"id":"user-1","name":"Test User","role":"admin","active":true,"email_address":"","created_at":"2025-01-01T00:00:00Z","url":""},"url":""}`))
	})
	testEnv(t, mux)

	resetFlags(t, commentCreateCmd)

	result := executeCommand(t, "comment", "create", "1", "--body", "Test comment")
	if result.err != nil {
		t.Fatalf("comment create: %v", result.err)
	}

	if gotPath != "/test-account/cards/1/comments" {
		t.Errorf("path = %s", gotPath)
	}

	commentPayload, ok := gotBody["comment"].(map[string]any)
	if !ok {
		t.Fatalf("missing 'comment' in body")
	}

	if commentPayload["body"] != "Test comment" {
		t.Errorf("body = %v", commentPayload["body"])
	}
}

func TestCommentCreate_MissingBody(t *testing.T) {
	mux := http.NewServeMux()
	testEnv(t, mux)

	resetFlags(t, commentCreateCmd)

	result := executeCommand(t, "comment", "create", "1")
	if result.err == nil {
		t.Fatal("expected error for missing --body")
	}

	if !strings.Contains(result.err.Error(), "--body is required") {
		t.Errorf("error = %v", result.err)
	}
}

func TestCommentDelete(t *testing.T) {
	var gotMethod, gotPath string

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/cards/1/comments/comment-1", func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path

		w.WriteHeader(http.StatusNoContent)
	})
	testEnv(t, mux)

	result := executeCommand(t, "comment", "delete", "1", "comment-1", "--yes")
	if result.err != nil {
		t.Fatalf("comment delete: %v", result.err)
	}

	if gotMethod != "DELETE" {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}

	if gotPath != "/test-account/cards/1/comments/comment-1" {
		t.Errorf("path = %s", gotPath)
	}
}

func TestCommentReactionCreate_RejectsTooLongBody(t *testing.T) {
	testEnv(t, http.NewServeMux())

	resetFlags(t, commentReactionCreateCmd)

	result := executeCommand(t, "comment", "reaction", "create", "1", "comment-1", "--body", strings.Repeat("a", 17))
	if result.err == nil {
		t.Fatal("expected error for too-long reaction body")
	}

	if !strings.Contains(result.err.Error(), "--body must be at most 16 characters") {
		t.Fatalf("error = %v", result.err)
	}
}

func TestCommentReactionDelete_RequiresYesWhenNonInteractive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/cards/1/comments/comment-1/reactions/reaction-1", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("reaction delete request should not be sent without confirmation")
	})
	testEnv(t, mux)

	result := executeCommand(t, "comment", "reaction", "delete", "1", "comment-1", "reaction-1")
	if result.err == nil {
		t.Fatal("expected confirmation error")
	}

	if !strings.Contains(result.err.Error(), "--yes") {
		t.Fatalf("error = %v", result.err)
	}
}
