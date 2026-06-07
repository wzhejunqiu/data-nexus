package config

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Log LogConfig `yaml:"log"`
}

type LogConfig struct {
	Level  string        `yaml:"level"`
	Output string        `yaml:"output"`
	File   LogFileConfig `yaml:"file"`
}

type LogFileConfig struct {
	Path       string `yaml:"path"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
	Compress   bool   `yaml:"compress"`
}

type CLIOverrides struct {
	LogLevel  string
	LogOutput string
	LogFile   string
	DBPath    string
}

var (
	flagLogLevel  = flag.String("log-level", "", "log level: debug|info|warn|error")
	flagLogOutput = flag.String("log-output", "", "log output: auto|console|file|both")
	flagLogFile   = flag.String("log-file", "", "log file path")
	flagDB        = flag.String("db", "", "open sqlite database on startup")
)

func ParseFlags() CLIOverrides {
	flag.Parse()
	cli := CLIOverrides{
		LogLevel:  *flagLogLevel,
		LogOutput: *flagLogOutput,
		LogFile:   *flagLogFile,
		DBPath:    *flagDB,
	}
	if cli.DBPath == "" {
		cli.DBPath = firstDatabaseArg(flag.Args())
	}
	return cli
}

func firstDatabaseArg(args []string) string {
	for _, arg := range args {
		if arg == "" || strings.HasPrefix(arg, "-") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(arg))
		switch ext {
		case ".db", ".sqlite", ".sqlite3":
			return arg
		}
	}
	return ""
}

func DefaultConfig() Config {
	return Config{
		Log: LogConfig{
			Level:  "info",
			Output: "auto",
			File: LogFileConfig{
				Path:       "",
				MaxSizeMB:  10,
				MaxBackups: 5,
				MaxAgeDays: 30,
				Compress:   true,
			},
		},
	}
}

func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".data-nexus"
	}
	return filepath.Join(home, ".data-nexus")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

func QueriesPath() string {
	return filepath.Join(ConfigDir(), "queries.json")
}

func SqlGlobalDBPath() string {
	return filepath.Join(ConfigDir(), "sql-global.db")
}

func DefaultExecutionLogConfig() model.ExecutionLogConfig {
	return model.ExecutionLogConfig{
		Driver: model.ExecutionLogSQLite,
		SQLite: &model.ExecutionLogSQLiteConfig{
			FilePath: SqlGlobalDBPath(),
		},
	}
}

func ConnectionsPath() string {
	return filepath.Join(ConfigDir(), "connections.json")
}

func Load() Config {
	cfg := DefaultConfig()
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return cfg
	}
	_ = yaml.Unmarshal(data, &cfg)
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Output == "" {
		cfg.Log.Output = "auto"
	}
	if cfg.Log.File.MaxSizeMB == 0 {
		cfg.Log.File.MaxSizeMB = 10
	}
	if cfg.Log.File.MaxBackups == 0 {
		cfg.Log.File.MaxBackups = 5
	}
	if cfg.Log.File.MaxAgeDays == 0 {
		cfg.Log.File.MaxAgeDays = 30
	}
	return cfg
}

func Save(cfg Config) error {
	if err := os.MkdirAll(ConfigDir(), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0o644)
}

func ApplyCLI(cfg Config, cli CLIOverrides) Config {
	if cli.LogLevel != "" {
		cfg.Log.Level = cli.LogLevel
	}
	if cli.LogOutput != "" {
		cfg.Log.Output = cli.LogOutput
	}
	if cli.LogFile != "" {
		cfg.Log.File.Path = cli.LogFile
	}
	return cfg
}

func DefaultLogFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(ConfigDir(), "data-nexus.log")
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Logs", "data-nexus", "data-nexus.log")
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "data-nexus", "logs", "data-nexus.log")
	default:
		return filepath.Join(home, ".local", "share", "data-nexus", "logs", "data-nexus.log")
	}
}

// IsDevMode returns true when running under wails dev (WAILS_DEV set).
func IsDevMode() bool {
	return os.Getenv("WAILS_DEV") != ""
}
