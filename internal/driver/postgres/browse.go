package postgres

import (
	"context"
	"fmt"
	"math"

	"github.com/wzhejunqiu/data-nexus/internal/driver/remote"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

func orderClauseForBrowse(opts model.BrowseOptions, order string) (string, error) {
	if opts.Sort == "" {
		return "", nil
	}
	if !sqlutil.IsSafeQuotedIdentifier(opts.Sort) {
		return "", model.ErrInvalidRequest("invalid sort column")
	}
	return fmt.Sprintf(" ORDER BY %s %s", quoteIdent(opts.Sort), order), nil
}

func (d *Driver) buildBrowseWhere(ctx context.Context, ref tableRef, opts model.BrowseOptions) (string, []any, error) {
	schema, err := d.GetTableSchema(ctx, ref.Qualified)
	if err != nil {
		return "", nil, err
	}
	return sqlutil.BuildWhereClause(opts.Filters, schema)
}

func (d *Driver) BrowseTable(ctx context.Context, tableName string, opts model.BrowseOptions) (*model.PaginatedTableData, error) {
	ref, err := d.parseTableRef(tableName)
	if err != nil {
		return nil, err
	}
	if _, err := d.lookupTableType(ctx, ref); err != nil {
		return nil, err
	}
	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	const maxPageSize = 200
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	where, whereArgs, err := d.buildBrowseWhere(ctx, ref, opts)
	if err != nil {
		return nil, err
	}

	var total int64
	totalPages := 1
	if !opts.SkipTotalCount {
		countQ := rebindPlaceholders(fmt.Sprintf("SELECT COUNT(*) FROM %s%s", ref.FromRef, where))
		if err := d.db.QueryRowContext(ctx, countQ, whereArgs...).Scan(&total); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
		if totalPages == 0 {
			totalPages = 1
		}
	}

	order := "ASC"
	if opts.Order == model.SortDesc {
		order = "DESC"
	}
	sortCol, err := orderClauseForBrowse(opts, order)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	query := rebindPlaceholders(fmt.Sprintf(
		"SELECT * FROM %s%s%s LIMIT ? OFFSET ?",
		ref.FromRef, where, sortCol,
	))
	queryArgs := append(append([]any{}, whereArgs...), pageSize, offset)
	rows, err := d.db.QueryContext(ctx, query, queryArgs...)
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
	for rows.Next() {
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
	}
	if resultRows == nil {
		resultRows = []map[string]any{}
	}

	return &model.PaginatedTableData{
		Columns: meta,
		Rows:    resultRows,
		Pagination: model.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalRows:  total,
			TotalPages: totalPages,
		},
	}, rows.Err()
}
