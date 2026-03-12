package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/test/boards" {
			t.Errorf("expected /test/boards, got %s", r.URL.Path)
		}

		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept header = %q", got)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization header = %q", got)
		}

		if got := r.Header.Get("User-Agent"); !strings.HasPrefix(got, "fz-cli") {
			t.Errorf("User-Agent header = %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"My Board"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "1.0.0")

	var result struct {
		Name string `json:"name"`
	}
	if err := client.Get(context.Background(), "/test/boards", &result); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if result.Name != "My Board" {
		t.Errorf("Name = %q, want %q", result.Name, "My Board")
	}
}

func TestGet_UserAgentVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "fz-cli/1.2.3" {
			t.Errorf("User-Agent = %q, want %q", got, "fz-cli/1.2.3")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "1.2.3")

	var result any

	_ = client.Get(context.Background(), "/test", &result)
}

func TestGet_UserAgentNoVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "fz-cli" {
			t.Errorf("User-Agent = %q, want %q", got, "fz-cli")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	var result any

	_ = client.Get(context.Background(), "/test", &result)
}

func TestPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type header = %q", got)
		}

		body, _ := io.ReadAll(r.Body)

		var payload map[string]string
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}

		if payload["name"] != "New Board" {
			t.Errorf("body name = %q", payload["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"abc123"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	var result struct {
		ID string `json:"id"`
	}

	err := client.Post(context.Background(), "/test/boards", strings.NewReader(`{"name":"New Board"}`), &result)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}

	if result.ID != "abc123" {
		t.Errorf("ID = %q, want %q", result.ID, "abc123")
	}
}

func TestPost_NilResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	err := client.Post(context.Background(), "/test/action", strings.NewReader(`{}`), nil)
	if err != nil {
		t.Fatalf("Post with nil v: %v", err)
	}
}

func TestPut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	err := client.Put(context.Background(), "/test/boards/1", strings.NewReader(`{"name":"Updated"}`), nil)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
}

func TestPatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	err := client.Patch(context.Background(), "/test/boards/1", strings.NewReader(`{"name":"Patched"}`), nil)
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
}

func TestDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/test/boards/1" {
			t.Errorf("path = %q", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")
	if err := client.Delete(context.Background(), "/test/boards/1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	var result any

	err := client.Get(context.Background(), "/test/boards/missing", &result)
	if err == nil {
		t.Fatal("expected error for 404")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func TestGet_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	var result any

	err := client.Get(context.Background(), "/my/identity", &result)
	if err == nil {
		t.Fatal("expected error for 401")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
}

func TestGet_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	var result any

	err := client.Get(context.Background(), "/test/fail", &result)
	if err == nil {
		t.Fatal("expected error for 500")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
}

func TestGetAll_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"a"},{"name":"b"}]`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	var result []struct{ Name string }
	if err := client.GetAll(context.Background(), "/items", &result); err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}

	if result[0].Name != "a" || result[1].Name != "b" {
		t.Errorf("result = %+v", result)
	}
}

func TestGetAll_Paginated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`[{"name":"c"}]`))
			return
		}
		// Page 1 — include Link header pointing to page 2.
		w.Header().Set("Link", `<`+r.URL.Path+`?page=2>; rel="next"`)
		_, _ = w.Write([]byte(`[{"name":"a"},{"name":"b"}]`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	var result []struct{ Name string }
	if err := client.GetAll(context.Background(), "/items", &result); err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("len = %d, want 3", len(result))
	}

	if result[2].Name != "c" {
		t.Errorf("result[2].Name = %q, want %q", result[2].Name, "c")
	}
}

func TestGetAll_AbsoluteURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`[{"name":"d"}]`))
			return
		}
		// Link header with absolute URL (like the real Fizzy API returns).
		nextURL := "http://" + r.Host + "/items?page=2"
		w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
		_, _ = w.Write([]byte(`[{"name":"c"}]`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	var result []struct{ Name string }
	if err := client.GetAll(context.Background(), "/items", &result); err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
}

func TestGetAll_TooManyPages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return a next page link — infinite pagination.
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Link", `</items?page=next>; rel="next"`)
		_, _ = w.Write([]byte(`[{"name":"x"}]`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", "")

	var result []struct{ Name string }

	err := client.GetAll(context.Background(), "/items", &result)
	if err == nil {
		t.Fatal("expected error for too many pages")
	}

	if !strings.Contains(err.Error(), "pagination limit exceeded") {
		t.Errorf("error = %q, want pagination limit exceeded", err.Error())
	}
}

func TestRequest_NoToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("Authorization header should be empty, got %q", got)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "", "")

	resp, err := client.Request(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("Request: %v", err)
	}

	_ = resp.Body.Close()
}

func TestRequest_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	client := NewClient(srv.URL, "test-token", "")

	_, err := client.Request(ctx, http.MethodGet, "/test", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
