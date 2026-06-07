package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func (d *Driver) OpenTableExport(ctx context.Context, tableName string, opts model.TableExportOptions) (export.TableExportCursor, error) {
	ref, err := parseTableRef(tableName)
	if err != nil {
		return nil, err
	}
	if _, err := d.lookupTableType(ctx, ref); err != nil {
		return nil, err
	}
	columns, err := d.exportColumnMeta(ctx, tableName)
	if err != nil {
		return nil, err
	}
	selectSQL := fmt.Sprintf("SELECT * FROM %q", tableName)
	return export.NewStreamingCursor(ctx, d.db, selectSQL, columns, opts.NormalizedBatchSize(), SerializeCellValue)
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

func (d *Driver) isWithoutRowID(ctx context.Context, ref tableRef) (bool, error) {
	var createSQL sql.NullString
	var err error
	if ref.Schema == "main" {
		err = d.db.QueryRowContext(ctx,
			`SELECT sql FROM sqlite_master WHERE name = ? AND type = 'table'`, ref.BareName).Scan(&createSQL)
	} else {
		q := fmt.Sprintf(`SELECT sql FROM %s.sqlite_master WHERE name = ? AND type = 'table'`, ref.Schema)
		err = d.db.QueryRowContext(ctx, q, ref.BareName).Scan(&createSQL)
	}
	if err == sql.ErrNoRows {
		return false, model.ErrTableNotFound(ref.Qualified)
	}
	if err != nil {
		return false, model.ErrSQL(err.Error())
	}
	if !createSQL.Valid {
		return false, nil
	}
	return strings.Contains(strings.ToUpper(createSQL.String), "WITHOUT ROWID"), nil
}
