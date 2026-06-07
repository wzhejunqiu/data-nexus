package logger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wzhejunqiu/data-nexus/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Manager struct {
	mu     sync.RWMutex
	logger *zap.Logger
	level  zap.AtomicLevel
}

func NewManager(cfg config.LogConfig, devMode bool) (*Manager, error) {
	level := zap.NewAtomicLevelAt(parseLevel(cfg.Level))
	m := &Manager{level: level}
	log, err := buildLogger(cfg, devMode, level)
	if err != nil {
		return nil, err
	}
	m.logger = log
	return m, nil
}

func (m *Manager) Logger() *zap.Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.logger
}

func (m *Manager) SetLevel(level string) {
	m.level.SetLevel(parseLevel(level))
}

func buildLogger(cfg config.LogConfig, devMode bool, level zap.AtomicLevel) (*zap.Logger, error) {
	output := cfg.Output
	if output == "auto" {
		if devMode {
			output = "console"
		} else {
			output = "file"
		}
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var cores []zapcore.Core

	if output == "console" || output == "both" {
		consoleEnc := zapcore.NewConsoleEncoder(encCfg)
		cores = append(cores, zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stdout), level))
	}
	if output == "file" || output == "both" {
		jsonEnc := zapcore.NewJSONEncoder(encCfg)
		path := cfg.File.Path
		if path == "" {
			path = config.DefaultLogFilePath()
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    cfg.File.MaxSizeMB,
			MaxBackups: cfg.File.MaxBackups,
			MaxAge:     cfg.File.MaxAgeDays,
			Compress:   cfg.File.Compress,
		})
		cores = append(cores, zapcore.NewCore(jsonEnc, w, level))
	}

	if len(cores) == 0 {
		consoleEnc := zapcore.NewConsoleEncoder(encCfg)
		cores = append(cores, zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stdout), level))
	}

	return zap.New(zapcore.NewTee(cores...), zap.AddCaller()), nil
}

func parseLevel(s string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// New preserves the previous API for tests and one-off use.
func New(cfg config.LogConfig, devMode bool) (*zap.Logger, error) {
	mgr, err := NewManager(cfg, devMode)
	if err != nil {
		return nil, err
	}
	return mgr.Logger(), nil
}
