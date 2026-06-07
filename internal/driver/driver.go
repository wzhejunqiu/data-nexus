package driver

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
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
	GetTableProfile(ctx context.Context, tableName string) (*model.TableProfile, error)
	BrowseTable(ctx context.Context, tableName string, opts model.BrowseOptions) (*model.PaginatedTableData, error)
	DetectFTSTable(ctx context.Context, tableName string) (*model.FTSInfo, error)
	Attach(ctx context.Context, filePath, alias string) error
	Detach(ctx context.Context, alias string) error
	ListAttached(ctx context.Context) ([]model.AttachedDatabase, error)
	OpenTableExport(ctx context.Context, tableName string, opts model.TableExportOptions) (export.TableExportCursor, error)
	QueryRows(ctx context.Context, sql string, params []any, maxRows int) (*model.QueryResult, error)
	Exec(ctx context.Context, sql string, params []any) (*model.ExecResult, error)
	UpdateCells(ctx context.Context, tableName string, changes []model.CellChange, schema *model.TableSchema) (int, error)
	ClassifySQL(ctx context.Context, sql string) (model.StatementKind, error)
}

func NewDriver(t model.DriverType) (Driver, error) {
	switch t {
	case model.DriverTypeSQLite:
		return sqlite.New(), nil
	default:
		return nil, model.ErrInvalidRequest("unsupported driver type")
	}
}
