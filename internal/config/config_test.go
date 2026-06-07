package config_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/config"
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
