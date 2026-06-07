package mysql

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/driver/remote"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

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
		countQ := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", ref.FromRef, where)
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
	query := fmt.Sprintf("SELECT * FROM %s%s%s LIMIT ? OFFSET ?", ref.FromRef, where, sortCol)
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

func orderClauseForBrowse(opts model.BrowseOptions, order string) (string, error) {
	if opts.Sort == "" {
		return "", nil
	}
	if !sqlutil.IsSafeQuotedIdentifier(opts.Sort) || strings.Contains(opts.Sort, "`") {
		return "", model.ErrInvalidRequest("invalid sort column")
	}
	return fmt.Sprintf(" ORDER BY %s %s", quoteIdent(opts.Sort), order), nil
}

func (d *Driver) buildBrowseWhere(ctx context.Context, ref tableRef, opts model.BrowseOptions) (string, []any, error) {
	schema, err := d.GetTableSchema(ctx, ref.Qualified)
	if err != nil {
		return "", nil, err
	}
	return buildWhereClause(opts.Filters, schema)
}

func buildWhereClause(filters []model.RowFilter, schema *model.TableSchema) (clause string, args []any, err error) {
	if len(filters) == 0 {
		return "", nil, nil
	}
	colSet := make(map[string]bool, len(schema.Columns))
	for _, c := range schema.Columns {
		colSet[c.Name] = true
	}

	var parts []string
	for _, f := range filters {
		if !sqlutil.IsSafeQuotedIdentifier(f.Column) || strings.Contains(f.Column, "`") || !colSet[f.Column] {
			return "", nil, model.ErrInvalidRequest(fmt.Sprintf("invalid filter column: %s", f.Column))
		}
		colRef := quoteIdent(f.Column)
		switch f.Operator {
		case model.FilterEq:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for eq")
			}
			parts = append(parts, colRef+" = ?")
			args = append(args, *f.Value)
		case model.FilterNe:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for ne")
			}
			parts = append(parts, colRef+" != ?")
			args = append(args, *f.Value)
		case model.FilterGt:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for gt")
			}
			parts = append(parts, colRef+" > ?")
			args = append(args, *f.Value)
		case model.FilterGte:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for gte")
			}
			parts = append(parts, colRef+" >= ?")
			args = append(args, *f.Value)
		case model.FilterLt:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for lt")
			}
			parts = append(parts, colRef+" < ?")
			args = append(args, *f.Value)
		case model.FilterLte:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for lte")
			}
			parts = append(parts, colRef+" <= ?")
			args = append(args, *f.Value)
		case model.FilterLike:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for like")
			}
			val := *f.Value
			if !strings.Contains(val, "%") && !strings.Contains(val, "_") {
				val = "%" + val + "%"
			}
			parts = append(parts, colRef+" LIKE ?")
			args = append(args, val)
		case model.FilterIsNull:
			parts = append(parts, colRef+" IS NULL")
		case model.FilterIsNotNull:
			parts = append(parts, colRef+" IS NOT NULL")
		case model.FilterIn:
			if len(f.Values) == 0 {
				return "", nil, model.ErrInvalidRequest("filter values required for in")
			}
			placeholders := make([]string, len(f.Values))
			for i, v := range f.Values {
				placeholders[i] = "?"
				args = append(args, v)
			}
			parts = append(parts, colRef+" IN ("+strings.Join(placeholders, ", ")+")")
		default:
			return "", nil, model.ErrInvalidRequest(fmt.Sprintf("unsupported filter operator: %s", f.Operator))
		}
	}
	return " WHERE " + strings.Join(parts, " AND "), args, nil
}
