package export

import (
	"context"
	"database/sql"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type CellSerializer func(any) any

type streamingCursor struct {
	rows      *sql.Rows
	colNames  []string
	columns   []model.ColumnMeta
	batchSize int
	serialize CellSerializer
	finished  bool
}

// NewStreamingCursor opens a single SELECT and reads rows in batches via NextBatch.
func NewStreamingCursor(
	ctx context.Context,
	db *sql.DB,
	selectSQL string,
	columns []model.ColumnMeta,
	batchSize int,
	serialize CellSerializer,
) (TableExportCursor, error) {
	if batchSize <= 0 {
		batchSize = 1000
	}
	if serialize == nil {
		serialize = func(v any) any { return v }
	}
	rows, err := db.QueryContext(ctx, selectSQL)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	colNames, err := rows.Columns()
	if err != nil {
		_ = rows.Close()
		return nil, model.ErrSQL(err.Error())
	}
	return &streamingCursor{
		rows:      rows,
		colNames:  colNames,
		columns:   columns,
		batchSize: batchSize,
		serialize: serialize,
	}, nil
}

func (c *streamingCursor) Columns() []model.ColumnMeta {
	return c.columns
}

func (c *streamingCursor) Close() error {
	if c.rows == nil {
		return nil
	}
	err := c.rows.Close()
	c.rows = nil
	return err
}

func (c *streamingCursor) NextBatch(ctx context.Context) (*model.TableExportBatch, error) {
	if c.finished || c.rows == nil {
		return &model.TableExportBatch{Rows: []map[string]any{}}, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var resultRows []map[string]any
	for len(resultRows) < c.batchSize {
		if !c.rows.Next() {
			c.finished = true
			break
		}
		values := make([]any, len(c.colNames))
		ptrs := make([]any, len(c.colNames))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := c.rows.Scan(ptrs...); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		row := make(map[string]any, len(c.colNames))
		for i, col := range c.colNames {
			row[col] = c.serialize(values[i])
		}
		resultRows = append(resultRows, row)
	}
	if err := c.rows.Err(); err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	if resultRows == nil {
		resultRows = []map[string]any{}
	}

	hasMore := !c.finished && len(resultRows) == c.batchSize
	return &model.TableExportBatch{Rows: resultRows, HasMore: hasMore}, nil
}
