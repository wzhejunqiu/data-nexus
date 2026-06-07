package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

// Tx wraps a database transaction for import and batch writes.
type Tx struct {
	tx *sql.Tx
}

func (d *Driver) BeginTx(ctx context.Context) (*Tx, error) {
	if d.db == nil {
		return nil, model.ErrConnectionNotFound("")
	}
	if d.readOnly {
		return nil, model.ErrReadOnly()
	}
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	return &Tx{tx: tx}, nil
}

func (t *Tx) Exec(ctx context.Context, sqlText string, params []any) (*model.ExecResult, error) {
	start := time.Now()
	res, err := t.tx.ExecContext(ctx, sqlText, params...)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "readonly") {
			return nil, model.ErrReadOnly()
		}
		return nil, model.ErrSQL(err.Error())
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return &model.ExecResult{
		RowsAffected: affected,
		LastInsertID: lastID,
		Duration:     time.Since(start),
	}, nil
}

func (t *Tx) QueryRows(ctx context.Context, sqlText string, params []any, maxRows int) (*model.QueryResult, error) {
	if maxRows <= 0 {
		maxRows = maxRowsDefault
	}
	if maxRows > maxRowsDefault {
		maxRows = maxRowsDefault
	}
	start := time.Now()
	rows, err := t.tx.QueryContext(ctx, sqlText, params...)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	meta := make([]model.ColumnMeta, len(cols))
	for i, name := range cols {
		meta[i] = model.ColumnMeta{Name: name, DataType: colTypes[i].DatabaseTypeName()}
	}

	var resultRows []map[string]any
	count := 0
	for rows.Next() {
		if count >= maxRows {
			break
		}
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = SerializeCellValue(values[i])
		}
		resultRows = append(resultRows, row)
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	if resultRows == nil {
		resultRows = []map[string]any{}
	}
	return &model.QueryResult{
		Columns:   meta,
		Rows:      resultRows,
		RowCount:  count,
		Truncated: count >= maxRows,
		Duration:  time.Since(start),
	}, nil
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}
