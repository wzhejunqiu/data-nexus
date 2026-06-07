package export

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type TableExportCursor interface {
	Columns() []model.ColumnMeta
	NextBatch(ctx context.Context) (*model.TableExportBatch, error)
	Close() error
}
