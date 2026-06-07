package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	_ "modernc.org/sqlite"
)

const maxRowsDefault = model.MaxQueryRows

var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type Driver struct {
	db       *sql.DB
	readOnly bool
}

func New() *Driver {
	return &Driver{}
}

func (d *Driver) Type() model.DriverType {
	return model.DriverTypeSQLite
}

func (d *Driver) ReadOnly() bool {
	return d.readOnly
}

func (d *Driver) Connect(ctx context.Context, cfg model.DriverConfig) error {
	if cfg.SQLite == nil {
		return model.ErrInvalidRequest("sqlite config required")
	}
	if d.db != nil {
		_ = d.Close()
	}
	d.readOnly = cfg.SQLite.ReadOnly
	dsn := fmt.Sprintf("file:%s", cfg.SQLite.FilePath)
	mode := "mode=rwc"
	if d.readOnly {
		mode = "mode=ro"
	}
	dsn = fmt.Sprintf("%s?%s", dsn, mode)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return model.ErrConnectionFailed(err.Error())
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		if strings.Contains(strings.ToLower(err.Error()), "locked") {
			return model.ErrDatabaseLocked()
		}
		return model.ErrConnectionFailed(err.Error())
	}
	d.db = db
	return nil
}

func (d *Driver) Close() error {
	if d.db == nil {
		return nil
	}
	err := d.db.Close()
	d.db = nil
	return err
}

func (d *Driver) Ping(ctx context.Context) error {
	if d.db == nil {
		return model.ErrConnectionNotFound("")
	}
	return d.db.PingContext(ctx)
}

func (d *Driver) ListTables(ctx context.Context) ([]model.TableInfo, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT name, type FROM sqlite_master
		WHERE type IN ('table', 'view')
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY name`)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var items []model.TableInfo
	for rows.Next() {
		var name, typ string
		if err := rows.Scan(&name, &typ); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		t := model.TableTypeTable
		if typ == "view" {
			t = model.TableTypeView
		}
		items = append(items, model.TableInfo{Name: name, Type: t})
	}
	return items, rows.Err()
}

func (d *Driver) GetTableSchema(ctx context.Context, tableName string) (*model.TableSchema, error) {
	if !identRe.MatchString(tableName) {
		return nil, model.ErrTableNotFound(tableName)
	}
	tableType, err := d.lookupTableType(ctx, tableName)
	if err != nil {
		return nil, err
	}
	cols, err := d.loadColumns(ctx, tableName)
	if err != nil {
		return nil, err
	}
	indexes, err := d.loadIndexes(ctx, tableName)
	if err != nil {
		return nil, err
	}
	return &model.TableSchema{
		Name:    tableName,
		Type:    tableType,
		Columns: cols,
		Indexes: indexes,
	}, nil
}

func (d *Driver) lookupTableType(ctx context.Context, tableName string) (model.TableType, error) {
	var typ string
	err := d.db.QueryRowContext(ctx,
		`SELECT type FROM sqlite_master WHERE name = ? AND type IN ('table','view')`, tableName).Scan(&typ)
	if err == sql.ErrNoRows {
		return "", model.ErrTableNotFound(tableName)
	}
	if err != nil {
		return "", model.ErrSQL(err.Error())
	}
	if typ == "view" {
		return model.TableTypeView, nil
	}
	return model.TableTypeTable, nil
}

func (d *Driver) loadColumns(ctx context.Context, tableName string) ([]model.ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%q)", tableName)
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var cols []model.ColumnInfo
	for rows.Next() {
		var cid, notnull, pk int
		var name, colType string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notnull, &dflt, &pk); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		var def *string
		if dflt.Valid {
			def = &dflt.String
		}
		native := colType
		cols = append(cols, model.ColumnInfo{
			Name:         name,
			DataType:     strings.ToUpper(colType),
			NativeType:   &native,
			Nullable:     notnull == 0,
			PrimaryKey:   pk > 0,
			DefaultValue: def,
			Position:     cid + 1,
		})
	}
	return cols, rows.Err()
}

func (d *Driver) loadIndexes(ctx context.Context, tableName string) ([]model.IndexInfo, error) {
	query := fmt.Sprintf("PRAGMA index_list(%q)", tableName)
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var indexes []model.IndexInfo
	for rows.Next() {
		var seq, unique, origin, partial int
		var name string
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		cols, err := d.loadIndexColumns(ctx, name)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, model.IndexInfo{
			Name:    name,
			Columns: cols,
			Unique:  unique == 1,
			Primary: origin == 1,
		})
	}
	return indexes, rows.Err()
}

func (d *Driver) loadIndexColumns(ctx context.Context, indexName string) ([]string, error) {
	query := fmt.Sprintf("PRAGMA index_info(%q)", indexName)
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var cols []string
	for rows.Next() {
		var seqno, cid int
		var name string
		if err := rows.Scan(&seqno, &cid, &name); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

func (d *Driver) BrowseTable(ctx context.Context, tableName string, opts model.BrowseOptions) (*model.PaginatedTableData, error) {
	if !identRe.MatchString(tableName) {
		return nil, model.ErrTableNotFound(tableName)
	}
	if _, err := d.lookupTableType(ctx, tableName); err != nil {
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
	if pageSize > 200 {
		pageSize = 200
	}

	var total int64
	if err := d.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %q", tableName)).Scan(&total); err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	order := "ASC"
	if opts.Order == model.SortDesc {
		order = "DESC"
	}
	sortCol := ""
	if opts.Sort != "" {
		if !identRe.MatchString(opts.Sort) {
			return nil, model.ErrInvalidRequest("invalid sort column")
		}
		sortCol = fmt.Sprintf(" ORDER BY %q %s", opts.Sort, order)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM %q%s LIMIT ? OFFSET ?", tableName, sortCol)
	rows, err := d.db.QueryContext(ctx, query, pageSize, offset)
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
			row[col] = SerializeCellValue(values[i])
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
	truncated := false
	count := 0
	for rows.Next() {
		if count >= maxRows {
			truncated = true
			break
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
			row[col] = SerializeCellValue(values[i])
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
		Truncated: truncated,
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

func SerializeCellValue(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []byte:
		return map[string]any{"type": "blob", "size": len(val)}
	default:
		return val
	}
}
