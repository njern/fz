package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/njern/fz/internal/config"
)

func TestCompletion_IgnoresMalformedConfig(t *testing.T) {
	configPath := writeMalformedDefaultConfig(t)
	_ = configPath

	result := executeCommand(t, "completion", "bash")
	if result.err != nil {
		t.Fatalf("completion bash: %v", result.err)
	}

	if !strings.Contains(result.stdout, "__start_fz") {
		t.Fatalf("completion output missing bash function header:\n%s", result.stdout)
	}
}

func TestConfigSet_RepairsMalformedConfig(t *testing.T) {
	configPath := writeMalformedDefaultConfig(t)

	result := executeCommand(t, "config", "set", "host", "https://example.test")
	if result.err != nil {
		t.Fatalf("config set: %v", result.err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("saved config is invalid JSON: %v", err)
	}

	if cfg.Host != "https://example.test" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "https://example.test")
	}
}

func TestAuthLogout_RepairsMalformedConfig(t *testing.T) {
	configPath := writeMalformedDefaultConfig(t)

	result := executeCommand(t, "auth", "logout", "--yes")
	if result.err != nil {
		t.Fatalf("auth logout: %v", result.err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("saved config is invalid JSON: %v", err)
	}

	if cfg.Token != "" {
		t.Fatalf("Token = %q, want empty", cfg.Token)
	}

	if cfg.DefaultAccount != "" {
		t.Fatalf("DefaultAccount = %q, want empty", cfg.DefaultAccount)
	}

	if cfg.Host != config.DefaultHost {
		t.Fatalf("Host = %q, want %q", cfg.Host, config.DefaultHost)
	}
}

func writeMalformedDefaultConfig(t *testing.T) string {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir, err := config.Dir()
	if err != nil {
		t.Fatalf("config.Dir: %v", err)
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	return path
}
