package config

import (
	"fmt"
	"strings"
)

var validLogLevels = map[string]struct{}{
	"debug": {}, "info": {}, "warn": {}, "error": {},
}

var validLogOutputs = map[string]struct{}{
	"auto": {}, "console": {}, "file": {}, "both": {},
}

func Validate(cfg Config) error {
	level := strings.ToLower(strings.TrimSpace(cfg.Log.Level))
	if _, ok := validLogLevels[level]; !ok {
		return fmt.Errorf("invalid log level: %s", cfg.Log.Level)
	}
	output := strings.ToLower(strings.TrimSpace(cfg.Log.Output))
	if _, ok := validLogOutputs[output]; !ok {
		return fmt.Errorf("invalid log output: %s", cfg.Log.Output)
	}
	if cfg.Log.File.MaxSizeMB < 1 {
		return fmt.Errorf("max_size_mb must be at least 1")
	}
	if cfg.Log.File.MaxBackups < 1 {
		return fmt.Errorf("max_backups must be at least 1")
	}
	if cfg.Log.File.MaxAgeDays < 1 {
		return fmt.Errorf("max_age_days must be at least 1")
	}
	path := strings.TrimSpace(cfg.Log.File.Path)
	if path != "" && strings.Contains(path, "..") {
		return fmt.Errorf("log file path must not contain double-dot segments")
	}
	return nil
}

func FirstDatabaseArg(args []string) string {
	return firstDatabaseArg(args)
}
