package cmd

import (
	"net/http"
	"strings"
	"testing"
)

func TestWebhookCreate_RejectsInvalidEvents(t *testing.T) {
	testEnv(t, http.NewServeMux())

	resetFlags(t, webhookCreateCmd)

	result := executeCommand(t, "webhook", "create", "board-1", "--name", "hook", "--url", "https://example.com", "--events", "invalid_event")
	if result.err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(result.err.Error(), "--events must be one of") {
		t.Fatalf("error = %v", result.err)
	}
}

func TestWebhookEdit_RejectsEmptyEvents(t *testing.T) {
	testEnv(t, http.NewServeMux())

	resetFlags(t, webhookEditCmd)

	result := executeCommand(t, "webhook", "edit", "board-1", "hook-1", "--events", "card_closed,")
	if result.err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(result.err.Error(), "--events must not contain empty values") {
		t.Fatalf("error = %v", result.err)
	}
}
