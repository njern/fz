package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestColumnList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards/board-1/columns", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(columnListJSON))
	})
	testEnv(t, mux)

	result := executeCommand(t, "column", "list", "board-1")

	if result.err != nil {
		t.Fatalf("column list: %v", result.err)
	}

	if !strings.Contains(result.stdout, "To Do") {
		t.Errorf("output missing column name, got:\n%s", result.stdout)
	}

	if !strings.Contains(result.stdout, "In Progress") {
		t.Errorf("output missing second column, got:\n%s", result.stdout)
	}
}

func TestColumnList_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards/board-1/columns", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(columnListEmptyJSON))
	})
	testEnv(t, mux)

	result := executeCommand(t, "column", "list", "board-1")

	if result.err != nil {
		t.Fatalf("column list empty: %v", result.err)
	}

	if !strings.Contains(result.stderr, "No columns found") {
		t.Errorf("expected 'No columns found', got:\n%s", result.stderr)
	}
}

func TestColumnCreate(t *testing.T) {
	var (
		gotPath string
		gotBody map[string]any
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /test-account/boards/board-1/columns", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"col-new","name":"Done","color":{"name":"Blue","value":"var(--color-card-default)"},"created_at":"2025-01-15T10:00:00Z"}`))
	})
	testEnv(t, mux)

	resetFlags(t, columnCreateCmd)

	result := executeCommand(t, "column", "create", "board-1", "--name", "Done")
	if result.err != nil {
		t.Fatalf("column create: %v", result.err)
	}

	if gotPath != "/test-account/boards/board-1/columns" {
		t.Errorf("path = %s", gotPath)
	}

	colPayload, ok := gotBody["column"].(map[string]any)
	if !ok {
		t.Fatalf("missing 'column' in body")
	}

	if colPayload["name"] != "Done" {
		t.Errorf("body name = %v", colPayload["name"])
	}
}

func TestColumnCreate_MissingName(t *testing.T) {
	mux := http.NewServeMux()
	testEnv(t, mux)

	resetFlags(t, columnCreateCmd)

	result := executeCommand(t, "column", "create", "board-1")
	if result.err == nil {
		t.Fatal("expected error for missing --name")
	}

	if !strings.Contains(result.err.Error(), "--name is required") {
		t.Errorf("error = %v", result.err)
	}
}

func TestColumnDelete(t *testing.T) {
	var gotMethod, gotPath string

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/boards/board-1/columns/col-1", func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path

		w.WriteHeader(http.StatusNoContent)
	})
	testEnv(t, mux)

	result := executeCommand(t, "column", "delete", "board-1", "col-1", "--yes")
	if result.err != nil {
		t.Fatalf("column delete: %v", result.err)
	}

	if gotMethod != "DELETE" {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}

	if gotPath != "/test-account/boards/board-1/columns/col-1" {
		t.Errorf("path = %s", gotPath)
	}
}
