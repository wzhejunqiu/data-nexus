package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

type sqliteExportCursor struct {
	db          *sql.DB
	tableName   string
	key         model.StableRowKey
	batchSize   int
	columns     []model.ColumnMeta
	lastKey     []any
	started     bool
	finished    bool
	useRowidSel bool // SELECT rowid, * when ordering by implicit rowid
}

func (d *Driver) OpenTableExport(ctx context.Context, tableName string, opts model.TableExportOptions) (export.TableExportCursor, error) {
	if !sqlutil.IsSafeQuotedIdentifier(tableName) {
		return nil, model.ErrTableNotFound(tableName)
	}
	if _, err := d.lookupTableType(ctx, tableName); err != nil {
		return nil, err
	}
	schema, err := d.GetTableSchema(ctx, tableName)
	if err != nil {
		return nil, err
	}
	withoutRowID, err := d.isWithoutRowID(ctx, tableName)
	if err != nil {
		return nil, err
	}
	key, err := export.ResolveStableRowKey(schema, model.DriverTypeSQLite, withoutRowID)
	if err != nil {
		return nil, err
	}

	columns, err := d.exportColumnMeta(ctx, tableName)
	if err != nil {
		return nil, err
	}

	return &sqliteExportCursor{
		db:          d.db,
		tableName:   tableName,
		key:         key,
		batchSize:   opts.NormalizedBatchSize(),
		columns:     columns,
		useRowidSel: key.Source == "rowid",
	}, nil
}

func (d *Driver) isWithoutRowID(ctx context.Context, tableName string) (bool, error) {
	var createSQL sql.NullString
	err := d.db.QueryRowContext(ctx,
		`SELECT sql FROM sqlite_master WHERE name = ? AND type = 'table'`, tableName).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return false, model.ErrTableNotFound(tableName)
	}
	if err != nil {
		return false, model.ErrSQL(err.Error())
	}
	if !createSQL.Valid {
		return false, nil
	}
	return strings.Contains(strings.ToUpper(createSQL.String), "WITHOUT ROWID"), nil
}

func (d *Driver) exportColumnMeta(ctx context.Context, tableName string) ([]model.ColumnMeta, error) {
	rows, err := d.db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %q LIMIT 0", tableName))
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

func (c *sqliteExportCursor) Columns() []model.ColumnMeta {
	return c.columns
}

func (c *sqliteExportCursor) Close() error {
	return nil
}

func (c *sqliteExportCursor) NextBatch(ctx context.Context) (*model.TableExportBatch, error) {
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
			if c.useRowidSel && i == 0 && col == "rowid" {
				lastKey = []any{values[i]}
				continue
			}
			row[col] = SerializeCellValue(values[i])
		}
		resultRows = append(resultRows, row)
		lastKey = c.extractKeyFromRow(row, lastKey)
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
		if len(lastKey) == 0 {
			lastKey = c.extractKeyFromRow(resultRows[len(resultRows)-1], nil)
		}
		c.lastKey = lastKey
	}
	if !hasMore {
		c.finished = true
	}

	return &model.TableExportBatch{Rows: resultRows, HasMore: hasMore}, nil
}

func (c *sqliteExportCursor) extractKeyFromRow(row map[string]any, fallback []any) []any {
	if c.useRowidSel && len(fallback) > 0 {
		return fallback
	}
	vals := make([]any, len(c.key.Columns))
	for i, col := range c.key.Columns {
		vals[i] = row[col]
	}
	return vals
}

func (c *sqliteExportCursor) buildKeysetQuery() (string, []any, error) {
	orderParts := make([]string, len(c.key.Columns))
	for i, col := range c.key.Columns {
		orderParts[i] = quoteKeyColumn(col) + " ASC"
	}
	orderClause := " ORDER BY " + strings.Join(orderParts, ", ")

	var whereClause string
	var args []any
	if c.started && len(c.lastKey) > 0 {
		if len(c.key.Columns) == 1 {
			whereClause = fmt.Sprintf(" WHERE %s > ?", quoteKeyColumn(c.key.Columns[0]))
			args = append(args, c.lastKey[0])
		} else {
			quotedCols := make([]string, len(c.key.Columns))
			placeholders := make([]string, len(c.key.Columns))
			for i, col := range c.key.Columns {
				quotedCols[i] = quoteKeyColumn(col)
				placeholders[i] = "?"
			}
			whereClause = fmt.Sprintf(" WHERE (%s) > (%s)",
				strings.Join(quotedCols, ", "), strings.Join(placeholders, ", "))
			args = append(args, c.lastKey...)
		}
	}

	selectFrom := fmt.Sprintf("SELECT * FROM %q", c.tableName)
	if c.useRowidSel {
		selectFrom = fmt.Sprintf("SELECT rowid, * FROM %q", c.tableName)
	}
	query := fmt.Sprintf("%s%s%s LIMIT ?", selectFrom, whereClause, orderClause)
	args = append(args, c.batchSize)
	return query, args, nil
}

func quoteKeyColumn(col string) string {
	if col == "rowid" {
		return "rowid"
	}
	return fmt.Sprintf("%q", col)
}
