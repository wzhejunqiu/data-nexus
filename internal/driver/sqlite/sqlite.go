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

	type tableEntry struct {
		name string
		typ  string
	}
	var entries []tableEntry
	for rows.Next() {
		var name, typ string
		if err := rows.Scan(&name, &typ); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		entries = append(entries, tableEntry{name: name, typ: typ})
	}
	if err := rows.Err(); err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	_ = rows.Close()

	var items []model.TableInfo
	for _, entry := range entries {
		if entry.typ == "view" {
			items = append(items, model.TableInfo{Name: entry.name, Type: model.TableTypeView})
			continue
		}
		var rowCount int64
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %q", entry.name)
		if err := d.db.QueryRowContext(ctx, countQuery).Scan(&rowCount); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		rc := rowCount
		items = append(items, model.TableInfo{Name: entry.name, Type: model.TableTypeTable, RowCount: &rc})
	}
	return items, nil
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

	type indexEntry struct {
		name    string
		unique  bool
		primary bool
	}
	var entries []indexEntry
	for rows.Next() {
		var seq, unique, partial int
		var name string
		var originRaw any
		if err := rows.Scan(&seq, &name, &unique, &originRaw, &partial); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		entries = append(entries, indexEntry{
			name:    name,
			unique:  unique == 1,
			primary: indexOriginIsPrimary(originRaw),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	_ = rows.Close()

	var indexes []model.IndexInfo
	for _, entry := range entries {
		cols, err := d.loadIndexColumns(ctx, entry.name)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, model.IndexInfo{
			Name:    entry.name,
			Columns: cols,
			Unique:  entry.unique,
			Primary: entry.primary,
		})
	}
	return indexes, nil
}

func indexOriginIsPrimary(origin any) bool {
	switch v := origin.(type) {
	case int64:
		return v == 1
	case int:
		return v == 1
	case string:
		return v == "pk" || v == "u"
	case []byte:
		s := string(v)
		return s == "pk" || s == "u"
	default:
		return false
	}
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

func orderClauseForBrowse(opts model.BrowseOptions, order string) (string, error) {
	if opts.Sort == "" {
		return "", nil
	}
	if !identRe.MatchString(opts.Sort) {
		return "", model.ErrInvalidRequest("invalid sort column")
	}
	return fmt.Sprintf(" ORDER BY %q %s", opts.Sort, order), nil
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
	const maxPageSize = 200
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	var total int64
	totalPages := 1
	if !opts.SkipTotalCount {
		if err := d.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %q", tableName)).Scan(&total); err != nil {
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

	schema, err := d.GetTableSchema(ctx, tableName)
	if err != nil {
		return nil, err
	}
	pkCols := primaryKeyColumns(schema)
	selectFrom := fmt.Sprintf("SELECT * FROM %q", tableName)
	if len(pkCols) == 0 {
		withoutRowID, err := d.isWithoutRowID(ctx, tableName)
		if err != nil {
			return nil, err
		}
		if !withoutRowID {
			selectFrom = fmt.Sprintf("SELECT rowid, * FROM %q", tableName)
		}
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("%s%s LIMIT ? OFFSET ?", selectFrom, sortCol)
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
