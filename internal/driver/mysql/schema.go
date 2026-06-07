package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

type tableRef struct {
	Database  string
	BareName  string
	FromRef   string
	Qualified string
}

func (d *Driver) parseTableRef(tableName string) (tableRef, error) {
	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		db, bare := parts[0], parts[1]
		if !sqlutil.IsSafeQuotedIdentifier(db) || !sqlutil.IsSafeQuotedIdentifier(bare) || strings.Contains(db, "`") || strings.Contains(bare, "`") {
			return tableRef{}, model.ErrTableNotFound(tableName)
		}
		if db != d.database {
			return tableRef{}, model.ErrTableNotFound(tableName)
		}
		return tableRef{
			Database:  db,
			BareName:  bare,
			FromRef:   tableFromRef(db, bare),
			Qualified: tableName,
		}, nil
	}
	if !sqlutil.IsSafeQuotedIdentifier(tableName) || strings.Contains(tableName, "`") {
		return tableRef{}, model.ErrTableNotFound(tableName)
	}
	return tableRef{
		Database:  d.database,
		BareName:  tableName,
		FromRef:   quoteIdent(tableName),
		Qualified: tableName,
	}, nil
}

func (d *Driver) ListTables(ctx context.Context) ([]model.TableInfo, error) {
	query := `
		SELECT TABLE_NAME, TABLE_TYPE
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_TYPE IN ('BASE TABLE', 'VIEW')
		ORDER BY TABLE_NAME`
	rows, err := d.db.QueryContext(ctx, query, d.database)
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

	schemaLabel := d.database
	var items []model.TableInfo
	for _, entry := range entries {
		info := model.TableInfo{Name: entry.name, Schema: &schemaLabel}
		if entry.typ == "VIEW" {
			info.Type = model.TableTypeView
			items = append(items, info)
			continue
		}
		fromRef := tableFromRef(d.database, entry.name)
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
	ref, err := d.parseTableRef(tableName)
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
	schemaPtr := &ref.Database
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
	err := d.db.QueryRowContext(ctx,
		`SELECT TABLE_TYPE FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`,
		ref.Database, ref.BareName).Scan(&typ)
	if err == sql.ErrNoRows {
		return "", model.ErrTableNotFound(ref.Qualified)
	}
	if err != nil {
		return "", model.ErrSQL(err.Error())
	}
	if typ == "VIEW" {
		return model.TableTypeView, nil
	}
	return model.TableTypeTable, nil
}

func (d *Driver) loadColumns(ctx context.Context, ref tableRef) ([]model.ColumnInfo, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT, ORDINAL_POSITION
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`
	rows, err := d.db.QueryContext(ctx, query, ref.Database, ref.BareName)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var cols []model.ColumnInfo
	for rows.Next() {
		var name, dataType, columnType, nullable, columnKey string
		var ordinal int
		var dflt sql.NullString
		if err := rows.Scan(&name, &dataType, &columnType, &nullable, &columnKey, &dflt, &ordinal); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		var def *string
		if dflt.Valid {
			def = &dflt.String
		}
		native := columnType
		cols = append(cols, model.ColumnInfo{
			Name:         name,
			DataType:     strings.ToUpper(dataType),
			NativeType:   &native,
			Nullable:     strings.EqualFold(nullable, "YES"),
			PrimaryKey:   columnKey == "PRI",
			DefaultValue: def,
			Position:     ordinal,
		})
	}
	return cols, rows.Err()
}

func (d *Driver) loadIndexes(ctx context.Context, ref tableRef) ([]model.IndexInfo, error) {
	query := `
		SELECT INDEX_NAME, NON_UNIQUE, SEQ_IN_INDEX, COLUMN_NAME
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX`
	rows, err := d.db.QueryContext(ctx, query, ref.Database, ref.BareName)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	type indexEntry struct {
		name    string
		unique  bool
		primary bool
		cols    []string
	}
	byName := make(map[string]*indexEntry)
	var order []string
	for rows.Next() {
		var indexName, colName string
		var nonUnique int
		var seq int
		if err := rows.Scan(&indexName, &nonUnique, &seq, &colName); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		entry, ok := byName[indexName]
		if !ok {
			entry = &indexEntry{
				name:    indexName,
				unique:  nonUnique == 0,
				primary: indexName == "PRIMARY",
			}
			byName[indexName] = entry
			order = append(order, indexName)
		}
		entry.cols = append(entry.cols, colName)
	}
	if err := rows.Err(); err != nil {
		return nil, model.ErrSQL(err.Error())
	}

	indexes := make([]model.IndexInfo, 0)
	for _, name := range order {
		entry := byName[name]
		indexes = append(indexes, model.IndexInfo{
			Name:    entry.name,
			Columns: append([]string(nil), entry.cols...),
			Unique:  entry.unique,
			Primary: entry.primary,
		})
	}
	return indexes, nil
}
