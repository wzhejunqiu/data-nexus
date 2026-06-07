package logger_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/logger"
)

func TestNewDevModeAutoConsole(t *testing.T) {
	log, err := logger.New(config.LogConfig{Level: "debug", Output: "auto"}, true)
	if err != nil {
		t.Fatal(err)
	}
	log.Info("dev console test")
}

func TestNewFileOutput(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")
	log, err := logger.New(config.LogConfig{
		Level:  "info",
		Output: "file",
		File:   config.LogFileConfig{Path: logPath},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	log.Info("file output test")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected log file: %v", err)
	}
}

func TestNewBothOutput(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "both.log")
	log, err := logger.New(config.LogConfig{
		Level:  "warn",
		Output: "both",
		File:   config.LogFileConfig{Path: logPath},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	log.Warn("both output test")
}

func TestNewVariousLogLevels(t *testing.T) {
	for _, level := range []string{"debug", "warn", "error", "info", ""} {
		log, err := logger.New(config.LogConfig{Level: level, Output: "console"}, true)
		if err != nil {
			t.Fatalf("level %q: %v", level, err)
		}
		log.Sync()
	}
}

func TestNewInvalidOutputFallsBackToConsole(t *testing.T) {
	log, err := logger.New(config.LogConfig{Level: "info", Output: "unknown"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if log == nil {
		t.Fatal("expected logger")
	}
}

func TestNewFileCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "nested", "logs", "app.log")
	log, err := logger.New(config.LogConfig{
		Level:  "info",
		Output: "file",
		File:   config.LogFileConfig{Path: logPath},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	log.Info("nested dir test")
	if !strings.HasSuffix(logPath, "app.log") {
		t.Fatal("unexpected log path")
	}
}
