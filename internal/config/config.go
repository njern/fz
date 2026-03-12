package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultHost = "https://app.fizzy.do"
	configDir   = "fz"
	configFile  = "config.json"
)

var ErrParse = errors.New("parsing config")

// Config holds the persistent CLI configuration.
type Config struct {
	Host           string `json:"host"`
	DefaultAccount string `json:"default_account,omitempty"`

	// Token is the personal access token or session token.
	Token string `json:"token,omitempty"`

	path string `json:"-"`
}

// Dir returns the config directory path (~/.config/fz).
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determining config directory: %w", err)
	}

	return filepath.Join(base, configDir), nil
}

// New returns a config initialized with default values and the default path.
func New() (*Config, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	return &Config{
		Host: DefaultHost,
		path: filepath.Join(dir, configFile),
	}, nil
}

// Load reads the config from disk, returning defaults if the file doesn't exist.
func Load() (*Config, error) {
	cfg, err := New()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cfg.path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParse, err)
	}

	if cfg.Host == "" {
		cfg.Host = DefaultHost
	}

	return cfg, nil
}

// Save writes the config to disk.
func (c *Config) Save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// SetPath overrides the config file path. Used in tests.
func (c *Config) SetPath(p string) {
	c.path = p
}

// Authenticated returns true if a token is configured.
func (c *Config) Authenticated() bool {
	return c.Token != ""
}

// AccountSlug returns the account to use, preferring the override if set.
// Leading slashes are stripped to avoid double-slash in request paths,
// since the API identity endpoint returns slugs like "/897362094".
func (c *Config) AccountSlug(override string) (string, error) {
	if override != "" {
		return strings.TrimPrefix(override, "/"), nil
	}

	if c.DefaultAccount != "" {
		return strings.TrimPrefix(c.DefaultAccount, "/"), nil
	}

	return "", fmt.Errorf("no account specified; use --account or run `fz auth login`")
}
