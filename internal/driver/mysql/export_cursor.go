package mysql

import (
	"context"
	"fmt"

	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
	"github.com/wzhejunqiu/data-nexus/internal/driver/remote"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func newExportCursor(d *Driver, ctx context.Context, tableName string, opts model.TableExportOptions) (export.TableExportCursor, error) {
	ref, err := d.parseTableRef(tableName)
	if err != nil {
		return nil, err
	}
	if _, err := d.lookupTableType(ctx, ref); err != nil {
		return nil, err
	}
	columns, err := d.exportColumnMeta(ctx, ref.FromRef)
	if err != nil {
		return nil, err
	}
	selectSQL := fmt.Sprintf("SELECT * FROM %s", ref.FromRef)
	return export.NewStreamingCursor(ctx, d.db, selectSQL, columns, opts.NormalizedBatchSize(), remote.SerializeCellValue)
}

func (d *Driver) exportColumnMeta(ctx context.Context, fromRef string) ([]model.ColumnMeta, error) {
	rows, err := d.db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s LIMIT 0", fromRef))
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
