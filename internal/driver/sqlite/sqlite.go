package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
	_ "modernc.org/sqlite"
)

const maxRowsDefault = model.MaxQueryRows

type Driver struct {
	db          *sql.DB
	readOnly    bool
	attachedDBs *attachState
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
	return d.applySQLitePragmas(ctx, cfg.SQLite)
}

func (d *Driver) applySQLitePragmas(ctx context.Context, cfg *model.SQLiteConfig) error {
	if cfg == nil || d.readOnly || !cfg.WAL {
		return nil
	}
	var mode string
	if err := d.db.QueryRowContext(ctx, "PRAGMA journal_mode=WAL").Scan(&mode); err != nil {
		return model.ErrSQL(err.Error())
	}
	if strings.EqualFold(mode, "wal") {
		return nil
	}
	return model.ErrSQL("failed to enable WAL journal mode")
}

func (d *Driver) Close() error {
	if d.db == nil {
		return nil
	}
	d.detachAll()
	err := d.db.Close()
	d.db = nil
	d.attachedDBs = nil
	return err
}

func (d *Driver) Ping(ctx context.Context) error {
	if d.db == nil {
		return model.ErrConnectionNotFound("")
	}
	return d.db.PingContext(ctx)
}

func (d *Driver) ListTables(ctx context.Context) ([]model.TableInfo, error) {
	mainSchema := "main"
	items, err := d.listTablesInSchema(ctx, mainSchema, &mainSchema)
	if err != nil {
		return nil, err
	}
	attached, err := d.ListAttached(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range attached {
		alias := a.Alias
		more, err := d.listTablesInSchema(ctx, alias, &alias)
		if err != nil {
			return nil, err
		}
		items = append(items, more...)
	}
	return items, nil
}

func (d *Driver) listTablesInSchema(ctx context.Context, schema string, schemaLabel *string) ([]model.TableInfo, error) {
	var masterRef string
	if schema == "main" {
		masterRef = "sqlite_master"
	} else {
		if !sqlutil.IsSafeQuotedIdentifier(schema) {
			return nil, model.ErrInvalidRequest("invalid schema")
		}
		masterRef = schema + ".sqlite_master"
	}
	query := fmt.Sprintf(`
		SELECT name, type FROM %s
		WHERE type IN ('table', 'view')
		  AND name NOT LIKE 'sqlite_%%'
		ORDER BY name`, masterRef)
	rows, err := d.db.QueryContext(ctx, query)
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
		info := model.TableInfo{Name: entry.name, Schema: schemaLabel}
		if entry.typ == "view" {
			info.Type = model.TableTypeView
			items = append(items, info)
			continue
		}
		var fromRef string
		if schema == "main" {
			fromRef = fmt.Sprintf("%q", entry.name)
		} else {
			fromRef = fmt.Sprintf("%s.%q", schema, entry.name)
		}
		var rowCount int64
		if err := d.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", fromRef)).Scan(&rowCount); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		rc := rowCount
		info.Type = model.TableTypeTable
		info.RowCount = &rc
		items = append(items, info)
	}
	return items, nil
}

func (d *Driver) GetTableSchema(ctx context.Context, tableName string) (*model.TableSchema, error) {
	ref, err := parseTableRef(tableName)
	if err != nil {
		return nil, err
	}
	tableType, err := d.lookupTableType(ctx, ref)
	if err != nil {
		return nil, err
	}
	cols, err := d.loadColumns(ctx, ref)
	if err != nil {
		return nil, err
	}
	indexes, err := d.loadIndexes(ctx, ref)
	if err != nil {
		return nil, err
	}
	var schemaPtr *string
	if ref.Schema != "main" {
		s := ref.Schema
		schemaPtr = &s
	}
	return &model.TableSchema{
		Name:    ref.BareName,
		Type:    tableType,
		Schema:  schemaPtr,
		Columns: cols,
		Indexes: indexes,
	}, nil
}

func (d *Driver) lookupTableType(ctx context.Context, ref tableRef) (model.TableType, error) {
	var typ string
	var err error
	if ref.Schema == "main" {
		err = d.db.QueryRowContext(ctx,
			`SELECT type FROM sqlite_master WHERE name = ? AND type IN ('table','view')`, ref.BareName).Scan(&typ)
	} else {
		q := fmt.Sprintf(`SELECT type FROM %s.sqlite_master WHERE name = ? AND type IN ('table','view')`, ref.Schema)
		err = d.db.QueryRowContext(ctx, q, ref.BareName).Scan(&typ)
	}
	if err == sql.ErrNoRows {
		return "", model.ErrTableNotFound(ref.Qualified)
	}
	if err != nil {
		return "", model.ErrSQL(err.Error())
	}
	if typ == "view" {
		return model.TableTypeView, nil
	}
	return model.TableTypeTable, nil
}

func pragmaTableRef(schema string) string {
	if schema == "main" {
		return ""
	}
	return schema + "."
}

func (d *Driver) loadColumns(ctx context.Context, ref tableRef) ([]model.ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA %stable_info(%q)", pragmaTableRef(ref.Schema), ref.BareName)
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

func (d *Driver) loadIndexes(ctx context.Context, ref tableRef) ([]model.IndexInfo, error) {
	query := fmt.Sprintf("PRAGMA %sindex_list(%q)", pragmaTableRef(ref.Schema), ref.BareName)
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

	indexes := make([]model.IndexInfo, 0)
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
	if !sqlutil.IsSafeQuotedIdentifier(opts.Sort) {
		return "", model.ErrInvalidRequest("invalid sort column")
	}
	return fmt.Sprintf(" ORDER BY %q %s", opts.Sort, order), nil
}

func (d *Driver) BrowseTable(ctx context.Context, tableName string, opts model.BrowseOptions) (*model.PaginatedTableData, error) {
	ref, err := parseTableRef(tableName)
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

	schema, err := d.GetTableSchema(ctx, ref.Qualified)
	if err != nil {
		return nil, err
	}
	pkCols := primaryKeyColumns(schema)
	selectFrom := fmt.Sprintf("SELECT * FROM %s", ref.FromRef)
	if len(pkCols) == 0 {
		withoutRowID, err := d.isWithoutRowID(ctx, ref)
		if err != nil {
			return nil, err
		}
		if !withoutRowID {
			selectFrom = fmt.Sprintf("SELECT rowid, * FROM %s", ref.FromRef)
		}
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("%s%s%s LIMIT ? OFFSET ?", selectFrom, where, sortCol)
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
