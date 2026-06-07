package model

type ExecutionLogDriverType string

const (
	ExecutionLogSQLite ExecutionLogDriverType = "sqlite"
)

type ExecutionLogConfig struct {
	Driver ExecutionLogDriverType    `yaml:"driver"`
	SQLite *ExecutionLogSQLiteConfig `yaml:"sqlite,omitempty"`
}

type ExecutionLogSQLiteConfig struct {
	FilePath string `yaml:"filePath"`
}
