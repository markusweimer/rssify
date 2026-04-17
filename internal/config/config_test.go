package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Cache.TTL != 30*time.Minute {
		t.Errorf("expected 30m TTL, got %v", cfg.Cache.TTL)
	}
	if cfg.Scraper.UserAgent != "rssify/1.0" {
		t.Errorf("expected 'rssify/1.0', got %q", cfg.Scraper.UserAgent)
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	// Should return defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoadValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte("server:\n  port: 9090\ncache:\n  ttl: 10m\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Cache.TTL != 10*time.Minute {
		t.Errorf("expected 10m TTL, got %v", cfg.Cache.TTL)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("RSS_PORT", "3000")
	t.Setenv("RSS_CACHE_TTL", "5m")
	t.Setenv("RSS_USER_AGENT", "test-agent")

	cfg, err := Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("expected port 3000, got %d", cfg.Server.Port)
	}
	if cfg.Cache.TTL != 5*time.Minute {
		t.Errorf("expected 5m TTL, got %v", cfg.Cache.TTL)
	}
	if cfg.Scraper.UserAgent != "test-agent" {
		t.Errorf("expected 'test-agent', got %q", cfg.Scraper.UserAgent)
	}
}

func TestEnvOverridesInvalidIgnored(t *testing.T) {
	t.Setenv("RSS_PORT", "notanumber")
	t.Setenv("RSS_CACHE_TTL", "invalid")

	cfg, err := Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should keep defaults when env vars are invalid
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Cache.TTL != 30*time.Minute {
		t.Errorf("expected default 30m TTL, got %v", cfg.Cache.TTL)
	}
}
