package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"gopkg.in/yaml.v3"
)

func TestDefaultLogFilePath(t *testing.T) {
	path := config.DefaultLogFilePath()
	switch runtime.GOOS {
	case "darwin":
		if !contains(path, "Library/Logs/data-nexus") {
			t.Fatalf("unexpected darwin log path: %s", path)
		}
	case "windows":
		if !contains(path, "data-nexus") || !contains(path, "logs") {
			t.Fatalf("unexpected windows log path: %s", path)
		}
	default:
		if !contains(path, ".local/share/data-nexus/logs") {
			t.Fatalf("unexpected linux log path: %s", path)
		}
	}
	if filepath.Base(path) != "data-nexus.log" {
		t.Fatalf("expected data-nexus.log basename, got %s", path)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func withHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.Log.Level != "info" || cfg.Log.Output != "auto" {
		t.Fatalf("unexpected defaults: %+v", cfg.Log)
	}
	if cfg.Log.File.MaxSizeMB != 10 || cfg.Log.File.MaxBackups != 5 || cfg.Log.File.MaxAgeDays != 30 || !cfg.Log.File.Compress {
		t.Fatalf("unexpected file defaults: %+v", cfg.Log.File)
	}
}

func TestConfigPaths(t *testing.T) {
	home := withHome(t)
	if config.ConfigDir() != filepath.Join(home, ".data-nexus") {
		t.Fatalf("unexpected config dir: %s", config.ConfigDir())
	}
	if config.ConfigPath() != filepath.Join(home, ".data-nexus", "config.yaml") {
		t.Fatal("unexpected config path")
	}
	if config.CatalogDBPath() != filepath.Join(home, ".data-nexus", "catalog.db") {
		t.Fatal("unexpected catalog db path")
	}
	if config.QueriesPath() != filepath.Join(home, ".data-nexus", "queries.json") {
		t.Fatal("unexpected queries path")
	}
	if config.SqlGlobalDBPath() != filepath.Join(home, ".data-nexus", "sql-global.db") {
		t.Fatal("unexpected sql global db path")
	}
	if config.VaultDir() != filepath.Join(home, ".data-nexus", "vault") {
		t.Fatal("unexpected vault dir")
	}
}

func TestDefaultExecutionLogConfig(t *testing.T) {
	home := withHome(t)
	cfg := config.DefaultExecutionLogConfig()
	if cfg.Driver != model.ExecutionLogSQLite || cfg.SQLite == nil {
		t.Fatalf("unexpected execution log config: %+v", cfg)
	}
	want := filepath.Join(home, ".data-nexus", "sql-global.db")
	if cfg.SQLite.FilePath != want {
		t.Fatalf("expected %s, got %s", want, cfg.SQLite.FilePath)
	}
}

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	withHome(t)
	cfg := config.Load()
	defaults := config.DefaultConfig()
	if cfg.Log.Level != defaults.Log.Level {
		t.Fatalf("expected default level, got %s", cfg.Log.Level)
	}
}

func TestLoadValidYAML(t *testing.T) {
	home := withHome(t)
	dir := filepath.Join(home, ".data-nexus")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := yaml.Marshal(map[string]any{
		"log": map[string]any{
			"level":  "debug",
			"output": "console",
			"file": map[string]any{
				"path": "/tmp/custom.log",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.ConfigPath(), data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Load()
	if cfg.Log.Level != "debug" || cfg.Log.Output != "console" {
		t.Fatalf("unexpected loaded config: %+v", cfg.Log)
	}
	if cfg.Log.File.Path != "/tmp/custom.log" {
		t.Fatalf("unexpected file path: %s", cfg.Log.File.Path)
	}
}

func TestLoadEmptyFieldsBackfilled(t *testing.T) {
	home := withHome(t)
	dir := filepath.Join(home, ".data-nexus")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.ConfigPath(), []byte("log: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Load()
	if cfg.Log.Level != "info" || cfg.Log.Output != "auto" {
		t.Fatalf("expected backfilled defaults, got %+v", cfg.Log)
	}
	if cfg.Log.File.MaxSizeMB != 10 {
		t.Fatalf("expected default max size, got %d", cfg.Log.File.MaxSizeMB)
	}
}

func TestApplyCLI(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg = config.ApplyCLI(cfg, config.CLIOverrides{
		LogLevel:  "error",
		LogOutput: "file",
		LogFile:   "/tmp/cli.log",
	})
	if cfg.Log.Level != "error" || cfg.Log.Output != "file" || cfg.Log.File.Path != "/tmp/cli.log" {
		t.Fatalf("unexpected applied config: %+v", cfg.Log)
	}
}

func TestIsDevMode(t *testing.T) {
	t.Setenv("WAILS_DEV", "")
	if config.IsDevMode() {
		t.Fatal("expected false without WAILS_DEV")
	}
	t.Setenv("WAILS_DEV", "1")
	if !config.IsDevMode() {
		t.Fatal("expected true with WAILS_DEV set")
	}
}

func TestConfigSave(t *testing.T) {
	home := withHome(t)
	cfg := config.DefaultConfig()
	cfg.Log.Level = "warn"
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded := config.Load()
	if loaded.Log.Level != "warn" {
		t.Fatalf("expected warn, got %s", loaded.Log.Level)
	}
	_ = home
}

func TestDefaultLogFilePathNoHome(t *testing.T) {
	t.Setenv("HOME", "")
	path := config.DefaultLogFilePath()
	if !contains(path, "data-nexus.log") {
		t.Fatalf("unexpected fallback log path: %s", path)
	}
}

func TestLoadCorruptYAMLUsesDefaults(t *testing.T) {
	home := withHome(t)
	dir := filepath.Join(home, ".data-nexus")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.ConfigPath(), []byte(":\n\tbad"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Load()
	if cfg.Log.Level != "info" {
		t.Fatalf("expected default level after corrupt yaml, got %s", cfg.Log.Level)
	}
}
