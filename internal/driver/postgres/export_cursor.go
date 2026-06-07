package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
	"github.com/wzhejunqiu/data-nexus/internal/driver/remote"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type pgExportCursor struct {
	inner export.TableExportCursor
	db    *sql.DB
}

func (c *pgExportCursor) Columns() []model.ColumnMeta {
	return c.inner.Columns()
}

func (c *pgExportCursor) NextBatch(ctx context.Context) (*model.TableExportBatch, error) {
	return c.inner.NextBatch(ctx)
}

func (c *pgExportCursor) Close() error {
	err := c.inner.Close()
	if _, rbErr := c.db.ExecContext(context.Background(), "ROLLBACK"); rbErr != nil && err == nil {
		err = model.ErrSQL(rbErr.Error())
	}
	return err
}

func rollbackExportTx(ctx context.Context, db *sql.DB) {
	_, _ = db.ExecContext(ctx, "ROLLBACK")
}

func newExportCursor(d *Driver, ctx context.Context, tableName string, opts model.TableExportOptions) (export.TableExportCursor, error) {
	ref, err := d.parseTableRef(tableName)
	if err != nil {
		return nil, err
	}
	if _, err := d.lookupTableType(ctx, ref); err != nil {
		return nil, err
	}
	columns, err := d.exportColumnMeta(ctx, ref)
	if err != nil {
		return nil, err
	}
	if _, err := d.db.ExecContext(ctx, "BEGIN READ ONLY"); err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	selectSQL := fmt.Sprintf("SELECT * FROM %s", ref.FromRef)
	inner, err := export.NewStreamingCursor(ctx, d.db, selectSQL, columns, opts.NormalizedBatchSize(), remote.SerializeCellValue)
	if err != nil {
		rollbackExportTx(ctx, d.db)
		return nil, err
	}
	return &pgExportCursor{inner: inner, db: d.db}, nil
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
