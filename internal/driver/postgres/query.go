package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/driver/remote"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

const maxRowsDefault = model.MaxQueryRows

func (d *Driver) QueryRows(ctx context.Context, sqlText string, params []any, maxRows int) (*model.QueryResult, error) {
	if maxRows <= 0 {
		maxRows = maxRowsDefault
	}
	if maxRows > maxRowsDefault {
		maxRows = maxRowsDefault
	}
	start := time.Now()
	rows, err := d.db.QueryContext(ctx, sqlText, params...)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	meta := make([]model.ColumnMeta, len(columns))
	for i, c := range columns {
		meta[i] = model.ColumnMeta{Name: c, DataType: colTypes[i].DatabaseTypeName()}
	}

	var resultRows []map[string]any
	count := 0
	for rows.Next() {
		if count >= maxRows {
			return nil, model.ErrResultTooLarge()
		}
		values := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			row[col] = remote.SerializeCellValue(values[i])
		}
		resultRows = append(resultRows, row)
		count++
	}
	if resultRows == nil {
		resultRows = []map[string]any{}
	}
	return &model.QueryResult{
		Columns:   meta,
		Rows:      resultRows,
		RowCount:  count,
		Truncated: false,
		Duration:  time.Since(start),
	}, rows.Err()
}

func (d *Driver) Exec(ctx context.Context, sqlText string, params []any) (*model.ExecResult, error) {
	if d.readOnly {
		return nil, model.ErrReadOnly()
	}
	start := time.Now()
	res, err := d.db.ExecContext(ctx, sqlText, params...)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "read-only") {
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

func (d *Driver) ClassifySQL(ctx context.Context, sqlText string) (model.StatementKind, error) {
	if d.db == nil {
		return "", model.ErrConnectionNotFound("")
	}
	return remote.ClassifySQL(sqlText)
}
