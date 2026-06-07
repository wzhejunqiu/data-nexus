package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
	"github.com/wzhejunqiu/data-nexus/internal/driver/remote"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type postgresExportCursor struct {
	db        *sql.DB
	fromRef   string
	key       model.StableRowKey
	batchSize int
	columns   []model.ColumnMeta
	lastKey   []any
	started   bool
	finished  bool
}

func newExportCursor(d *Driver, ctx context.Context, tableName string, opts model.TableExportOptions) (export.TableExportCursor, error) {
	ref, err := d.parseTableRef(tableName)
	if err != nil {
		return nil, err
	}
	if _, err := d.lookupTableType(ctx, ref); err != nil {
		return nil, err
	}
	schema, err := d.GetTableSchema(ctx, ref.Qualified)
	if err != nil {
		return nil, err
	}
	key, err := export.ResolveStableRowKey(schema, model.DriverTypePostgres, false)
	if err != nil {
		return nil, err
	}
	columns, err := d.exportColumnMeta(ctx, ref)
	if err != nil {
		return nil, err
	}
	return &postgresExportCursor{
		db:        d.db,
		fromRef:   ref.FromRef,
		key:       key,
		batchSize: opts.NormalizedBatchSize(),
		columns:   columns,
	}, nil
}

func (d *Driver) exportColumnMeta(ctx context.Context, ref tableRef) ([]model.ColumnMeta, error) {
	rows, err := d.db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s LIMIT 0", ref.FromRef))
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	names, err := rows.Columns()
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	meta := make([]model.ColumnMeta, len(names))
	for i, name := range names {
		meta[i] = model.ColumnMeta{Name: name, DataType: colTypes[i].DatabaseTypeName()}
	}
	return meta, rows.Err()
}

func (c *postgresExportCursor) Columns() []model.ColumnMeta {
	return c.columns
}

func (c *postgresExportCursor) Close() error {
	return nil
}

func (c *postgresExportCursor) NextBatch(ctx context.Context) (*model.TableExportBatch, error) {
	if c.finished {
		return &model.TableExportBatch{Rows: []map[string]any{}}, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	query, args, err := c.buildKeysetQuery()
	if err != nil {
		return nil, err
	}
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	colNames, err := rows.Columns()
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}

	var resultRows []map[string]any
	var lastKey []any
	for rows.Next() {
		values := make([]any, len(colNames))
		ptrs := make([]any, len(colNames))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		row := make(map[string]any, len(c.columns))
		for i, col := range colNames {
			row[col] = remote.SerializeCellValue(values[i])
		}
		resultRows = append(resultRows, row)
		lastKey = c.extractKeyFromRow(row)
	}
	if err := rows.Err(); err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	if resultRows == nil {
		resultRows = []map[string]any{}
	}

	c.started = true
	hasMore := len(resultRows) == c.batchSize
	if len(resultRows) > 0 {
		c.lastKey = lastKey
	}
	if !hasMore {
		c.finished = true
	}

	return &model.TableExportBatch{Rows: resultRows, HasMore: hasMore}, nil
}

func (c *postgresExportCursor) extractKeyFromRow(row map[string]any) []any {
	vals := make([]any, len(c.key.Columns))
	for i, col := range c.key.Columns {
		vals[i] = row[col]
	}
	return vals
}

func (c *postgresExportCursor) buildKeysetQuery() (string, []any, error) {
	orderParts := make([]string, len(c.key.Columns))
	for i, col := range c.key.Columns {
		orderParts[i] = quoteIdent(col) + " ASC"
	}
	orderClause := " ORDER BY " + strings.Join(orderParts, ", ")

	var whereClause string
	var args []any
	if c.started && len(c.lastKey) > 0 {
		if len(c.key.Columns) == 1 {
			whereClause = fmt.Sprintf(" WHERE %s > $1", quoteIdent(c.key.Columns[0]))
			args = append(args, c.lastKey[0])
		} else {
			quotedCols := make([]string, len(c.key.Columns))
			placeholders := make([]string, len(c.key.Columns))
			for i, col := range c.key.Columns {
				quotedCols[i] = quoteIdent(col)
				placeholders[i] = fmt.Sprintf("$%d", i+1)
			}
			whereClause = fmt.Sprintf(" WHERE (%s) > (%s)",
				strings.Join(quotedCols, ", "), strings.Join(placeholders, ", "))
			args = append(args, c.lastKey...)
		}
	}

	limitPlaceholder := fmt.Sprintf("$%d", len(args)+1)
	query := fmt.Sprintf("SELECT * FROM %s%s%s LIMIT %s", c.fromRef, whereClause, orderClause, limitPlaceholder)
	args = append(args, c.batchSize)
	return query, args, nil
}
