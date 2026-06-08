package executionlog

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/executionlog/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type Store interface {
	Type() model.ExecutionLogDriverType
	Close() error
	Insert(ctx context.Context, record model.SqlExecutionRecord) error
	ListQueryHistory(ctx context.Context, connectionID string, limit int) ([]string, error)
	ListExecutions(ctx context.Context, connectionID string, limit int) ([]model.SqlExecutionRecord, error)
	ListAllExecutions(ctx context.Context, limit int) ([]model.SqlExecutionRecord, error)
}

func NewStore(cfg model.ExecutionLogConfig) (Store, error) {
	switch cfg.Driver {
	case model.ExecutionLogSQLite, "":
		return sqlite.NewStore(cfg.SQLite)
	default:
		return nil, model.ErrInvalidRequest("unsupported execution log driver")
	}
}
