package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestBoardList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(boardListJSON))
	})
	testEnv(t, mux)

	result := executeCommand(t, "board", "list")

	if result.err != nil {
		t.Fatalf("board list: %v", result.err)
	}

	if !strings.Contains(result.stdout, "Test Board") {
		t.Errorf("output missing board name, got:\n%s", result.stdout)
	}

	if !strings.Contains(result.stdout, "Another Board") {
		t.Errorf("output missing second board, got:\n%s", result.stdout)
	}

	if !strings.Contains(result.stdout, "board-1") {
		t.Errorf("output missing board ID, got:\n%s", result.stdout)
	}
}

func TestBoardList_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(boardListEmptyJSON))
	})
	testEnv(t, mux)

	result := executeCommand(t, "board", "list")

	if result.err != nil {
		t.Fatalf("board list empty: %v", result.err)
	}

	if !strings.Contains(result.stderr, "No boards found") {
		t.Errorf("expected 'No boards found', got:\n%s", result.stderr)
	}
}

func TestBoardView_JSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards/board-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(boardViewJSON))
	})
	mux.HandleFunc("GET /test-account/boards/board-1/columns", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(columnListJSON))
	})
	mux.HandleFunc("GET /test-account/cards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cardListJSON))
	})
	testEnv(t, mux)

	resetFlags(t, boardViewCmd)

	result := executeCommand(t, "board", "view", "board-1", "--json")

	if result.err != nil {
		t.Fatalf("board view --json: %v", result.err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(result.stdout), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, result.stdout)
	}

	board, ok := parsed["board"].(map[string]any)
	if !ok {
		t.Fatalf("missing 'board' key in output")
	}

	if board["name"] != "Test Board" {
		t.Errorf("board name = %v", board["name"])
	}
}

func TestBoardView_JSON_IncludesBuiltInColumnsWhenEmpty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards/board-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(boardViewJSON))
	})
	mux.HandleFunc("GET /test-account/boards/board-1/columns", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(columnListEmptyJSON))
	})
	mux.HandleFunc("GET /test-account/cards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cardListEmptyJSON))
	})
	testEnv(t, mux)

	resetFlags(t, boardViewCmd)

	result := executeCommand(t, "board", "view", "board-1", "--json")
	if result.err != nil {
		t.Fatalf("board view --json: %v", result.err)
	}

	var parsed struct {
		Columns []struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"columns"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, result.stdout)
	}

	if len(parsed.Columns) != 3 {
		t.Fatalf("columns len = %d, want 3; output:\n%s", len(parsed.Columns), result.stdout)
	}

	gotNames := []string{parsed.Columns[0].Name, parsed.Columns[1].Name, parsed.Columns[2].Name}
	wantNames := []string{"Not Now", "Maybe?", "Done"}

	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Fatalf("column %d name = %q, want %q", i, gotNames[i], wantNames[i])
		}
	}

	if parsed.Columns[0].Color != "Gray" {
		t.Fatalf("Not Now color = %q, want %q", parsed.Columns[0].Color, "Gray")
	}

	if parsed.Columns[1].Color != "Blue" {
		t.Fatalf("Maybe color = %q, want %q", parsed.Columns[1].Color, "Blue")
	}

	if parsed.Columns[2].Color != "Gray" {
		t.Fatalf("Done color = %q, want %q", parsed.Columns[2].Color, "Gray")
	}
}

func TestBoardView_JSON_FetchesBuiltInLaneCards(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards/board-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(boardViewJSON))
	})
	mux.HandleFunc("GET /test-account/boards/board-1/columns", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(columnListEmptyJSON))
	})
	mux.HandleFunc("GET /test-account/cards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Query().Get("indexed_by") {
		case "not_now":
			_, _ = w.Write([]byte(`[
			  {
			    "id": "card-not-now",
			    "number": 2,
			    "title": "Later Card",
			    "status": "published",
			    "description": "",
			    "tags": [],
			    "closed": false,
			    "postponed": true,
			    "golden": false,
			    "last_active_at": "2025-03-02T12:00:00Z",
			    "created_at": "2025-02-02T09:00:00Z",
			    "url": "https://app.fizzy.do/test-account/cards/2",
			    "board": {"id": "board-1", "name": "Test Board", "all_access": true, "created_at": "2025-01-15T10:00:00Z", "url": "", "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""}},
			    "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""},
			    "comments_url": ""
			  }
			]`))
		case "closed":
			_, _ = w.Write([]byte(`[
			  {
			    "id": "card-done",
			    "number": 3,
			    "title": "Done Card",
			    "status": "published",
			    "description": "",
			    "tags": [],
			    "closed": true,
			    "postponed": false,
			    "golden": false,
			    "last_active_at": "2025-03-03T12:00:00Z",
			    "created_at": "2025-02-03T09:00:00Z",
			    "url": "https://app.fizzy.do/test-account/cards/3",
			    "board": {"id": "board-1", "name": "Test Board", "all_access": true, "created_at": "2025-01-15T10:00:00Z", "url": "", "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""}},
			    "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""},
			    "comments_url": ""
			  }
			]`))
		default:
			_, _ = w.Write([]byte(`[
			  {
			    "id": "card-maybe",
			    "number": 1,
			    "title": "Maybe Card",
			    "status": "published",
			    "description": "",
			    "tags": [],
			    "closed": false,
			    "postponed": false,
			    "golden": false,
			    "last_active_at": "2025-03-01T12:00:00Z",
			    "created_at": "2025-02-01T09:00:00Z",
			    "url": "https://app.fizzy.do/test-account/cards/1",
			    "board": {"id": "board-1", "name": "Test Board", "all_access": true, "created_at": "2025-01-15T10:00:00Z", "url": "", "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""}},
			    "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""},
			    "comments_url": ""
			  }
			]`))
		}
	})
	testEnv(t, mux)

	resetFlags(t, boardViewCmd)

	result := executeCommand(t, "board", "view", "board-1", "--json")
	if result.err != nil {
		t.Fatalf("board view --json: %v", result.err)
	}

	var parsed struct {
		Columns []struct {
			Name  string `json:"name"`
			Color string `json:"color"`
			Cards []struct {
				Title string `json:"title"`
			} `json:"cards"`
		} `json:"columns"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, result.stdout)
	}

	if len(parsed.Columns) != 3 {
		t.Fatalf("columns len = %d, want 3; output:\n%s", len(parsed.Columns), result.stdout)
	}

	if parsed.Columns[0].Name != "Not Now" || len(parsed.Columns[0].Cards) != 1 || parsed.Columns[0].Cards[0].Title != "Later Card" {
		t.Fatalf("unexpected Not Now column: %#v", parsed.Columns[0])
	}
	if parsed.Columns[0].Color != "Gray" {
		t.Fatalf("unexpected Not Now color: %#v", parsed.Columns[0])
	}

	if parsed.Columns[1].Name != "Maybe?" || len(parsed.Columns[1].Cards) != 1 || parsed.Columns[1].Cards[0].Title != "Maybe Card" {
		t.Fatalf("unexpected Maybe column: %#v", parsed.Columns[1])
	}
	if parsed.Columns[1].Color != "Blue" {
		t.Fatalf("unexpected Maybe color: %#v", parsed.Columns[1])
	}

	if parsed.Columns[2].Name != "Done" || len(parsed.Columns[2].Cards) != 1 || parsed.Columns[2].Cards[0].Title != "Done Card" {
		t.Fatalf("unexpected Done column: %#v", parsed.Columns[2])
	}
	if parsed.Columns[2].Color != "Gray" {
		t.Fatalf("unexpected Done color: %#v", parsed.Columns[2])
	}
}

func TestBoardView_JSON_KeepsCustomColumnsBeforeDone(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards/board-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(boardViewJSON))
	})
	mux.HandleFunc("GET /test-account/boards/board-1/columns", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(columnListJSON))
	})
	mux.HandleFunc("GET /test-account/cards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cardListEmptyJSON))
	})
	testEnv(t, mux)

	resetFlags(t, boardViewCmd)

	result := executeCommand(t, "board", "view", "board-1", "--json")
	if result.err != nil {
		t.Fatalf("board view --json: %v", result.err)
	}

	var parsed struct {
		Columns []struct {
			Name string `json:"name"`
		} `json:"columns"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, result.stdout)
	}

	gotNames := []string{
		parsed.Columns[0].Name,
		parsed.Columns[1].Name,
		parsed.Columns[2].Name,
		parsed.Columns[3].Name,
		parsed.Columns[4].Name,
	}
	wantNames := []string{"Not Now", "Maybe?", "To Do", "In Progress", "Done"}

	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Fatalf("column %d name = %q, want %q; output:\n%s", i, gotNames[i], wantNames[i], result.stdout)
		}
	}
}

func TestBoardView_Text_IncludesBuiltInColumnsWhenEmpty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards/board-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(boardViewJSON))
	})
	mux.HandleFunc("GET /test-account/boards/board-1/columns", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(columnListEmptyJSON))
	})
	mux.HandleFunc("GET /test-account/cards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cardListEmptyJSON))
	})
	testEnv(t, mux)

	resetFlags(t, boardViewCmd)

	result := executeCommand(t, "board", "view", "board-1")
	if result.err != nil {
		t.Fatalf("board view: %v", result.err)
	}

	for _, name := range []string{"Not Now (0)", "Maybe? (0)", "Done (0)"} {
		if !strings.Contains(result.stdout, name) {
			t.Fatalf("output missing %q:\n%s", name, result.stdout)
		}
	}
}

func TestBoardView_Text_KeepsCustomColumnsBeforeDone(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-account/boards/board-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(boardViewJSON))
	})
	mux.HandleFunc("GET /test-account/boards/board-1/columns", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(columnListJSON))
	})
	mux.HandleFunc("GET /test-account/cards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cardListEmptyJSON))
	})
	testEnv(t, mux)

	resetFlags(t, boardViewCmd)

	result := executeCommand(t, "board", "view", "board-1")
	if result.err != nil {
		t.Fatalf("board view: %v", result.err)
	}

	notNow := strings.Index(result.stdout, "Not Now (0)")
	maybe := strings.Index(result.stdout, "Maybe? (0)")
	todo := strings.Index(result.stdout, "To Do (0)")
	inProgress := strings.Index(result.stdout, "In Progress (0)")
	done := strings.Index(result.stdout, "Done (0)")

	if !(notNow < maybe && maybe < todo && todo < inProgress && inProgress < done) {
		t.Fatalf("unexpected column order:\n%s", result.stdout)
	}
}

func TestBoardCreate(t *testing.T) {
	var (
		gotMethod, gotPath string
		gotBody            map[string]any
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /test-account/boards", func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"board-new","name":"New Board","url":"https://app.fizzy.do/test-account/boards/board-new","all_access":true,"created_at":"2025-01-15T10:00:00Z","creator":{"id":"user-1","name":"Test User","role":"admin","active":true,"email_address":"","created_at":"2025-01-01T00:00:00Z","url":""}}`))
	})
	testEnv(t, mux)

	result := executeCommand(t, "board", "create", "New Board")
	if result.err != nil {
		t.Fatalf("board create: %v", result.err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %s, want POST", gotMethod)
	}

	if gotPath != "/test-account/boards" {
		t.Errorf("path = %s", gotPath)
	}

	boardPayload, ok := gotBody["board"].(map[string]any)
	if !ok {
		t.Fatalf("missing 'board' in body")
	}

	if boardPayload["name"] != "New Board" {
		t.Errorf("body name = %v", boardPayload["name"])
	}
}

func TestBoardDelete(t *testing.T) {
	var gotMethod, gotPath string

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/boards/board-1", func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path

		w.WriteHeader(http.StatusNoContent)
	})
	testEnv(t, mux)

	result := executeCommand(t, "board", "delete", "board-1", "--yes")
	if result.err != nil {
		t.Fatalf("board delete: %v", result.err)
	}

	if gotMethod != "DELETE" {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}

	if gotPath != "/test-account/boards/board-1" {
		t.Errorf("path = %s", gotPath)
	}
}

func TestBoardDelete_RequiresYesWhenNonInteractive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/boards/board-1", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("delete request should not be sent without confirmation")
	})
	testEnv(t, mux)

	result := executeCommand(t, "board", "delete", "board-1")
	if result.err == nil {
		t.Fatal("expected confirmation error")
	}

	if !strings.Contains(result.err.Error(), "--yes") {
		t.Fatalf("error = %v", result.err)
	}
}

func TestBoardUnpublish_RequiresYesWhenNonInteractive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /test-account/boards/board-1/publication", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unpublish request should not be sent without confirmation")
	})
	testEnv(t, mux)

	result := executeCommand(t, "board", "unpublish", "board-1")
	if result.err == nil {
		t.Fatal("expected confirmation error")
	}

	if !strings.Contains(result.err.Error(), "--yes") {
		t.Fatalf("error = %v", result.err)
	}
}

func TestBoardEdit_ClearsDescriptionAndUserIDs(t *testing.T) {
	var gotBody map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /test-account/boards/board-1", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		w.WriteHeader(http.StatusNoContent)
	})
	testEnv(t, mux)

	resetFlags(t, boardEditCmd)

	result := executeCommand(t, "board", "edit", "board-1", "--description", "", "--clear-user-ids")
	if result.err != nil {
		t.Fatalf("board edit: %v", result.err)
	}

	boardPayload, ok := gotBody["board"].(map[string]any)
	if !ok {
		t.Fatalf("missing board payload: %#v", gotBody)
	}

	if value, ok := boardPayload["public_description"]; !ok || value != "" {
		t.Fatalf("public_description = %#v, want empty string", value)
	}

	userIDs, ok := boardPayload["user_ids"].([]any)
	if !ok {
		t.Fatalf("user_ids = %#v, want empty array", boardPayload["user_ids"])
	}

	if len(userIDs) != 0 {
		t.Fatalf("user_ids len = %d, want 0", len(userIDs))
	}
}
