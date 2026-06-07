package driver

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/driver/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type Driver interface {
	Type() model.DriverType
	Connect(ctx context.Context, cfg model.DriverConfig) error
	Close() error
	Ping(ctx context.Context) error
	ReadOnly() bool
	ListTables(ctx context.Context) ([]model.TableInfo, error)
	GetTableSchema(ctx context.Context, tableName string) (*model.TableSchema, error)
	BrowseTable(ctx context.Context, tableName string, opts model.BrowseOptions) (*model.PaginatedTableData, error)
	QueryRows(ctx context.Context, sql string, params []any, maxRows int) (*model.QueryResult, error)
	Exec(ctx context.Context, sql string, params []any) (*model.ExecResult, error)
}

func NewDriver(t model.DriverType) (Driver, error) {
	switch t {
	case model.DriverTypeSQLite:
		return sqlite.New(), nil
	default:
		return nil, model.ErrInvalidRequest("unsupported driver type")
	}
}
