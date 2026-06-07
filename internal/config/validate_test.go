package config_test

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/config"
)

func TestValidateConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}

	cfg.Log.Level = "verbose"
	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected invalid level error")
	}

	cfg = config.DefaultConfig()
	cfg.Log.File.MaxSizeMB = 0
	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected max_size_mb error")
	}

	cfg = config.DefaultConfig()
	cfg.Log.File.Path = "../etc/passwd"
	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected path error")
	}
}

func TestFirstDatabaseArg(t *testing.T) {
	if got := config.FirstDatabaseArg([]string{"-log-level", "debug"}); got != "" {
		t.Fatalf("expected empty when no db path, got %q", got)
	}
	if got := config.FirstDatabaseArg([]string{"/tmp/app.db"}); got != "/tmp/app.db" {
		t.Fatalf("expected /tmp/app.db, got %q", got)
	}
	if got := config.FirstDatabaseArg([]string{"--db", "/tmp/a.db"}); got != "/tmp/a.db" {
		t.Fatalf("expected /tmp/a.db after flag, got %q", got)
	}
	if got := config.FirstDatabaseArg([]string{"-log-level", "debug", "/data/x.sqlite3"}); got != "/data/x.sqlite3" {
		t.Fatalf("expected sqlite path, got %q", got)
	}
}
