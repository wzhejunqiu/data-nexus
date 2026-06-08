package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func (d *Driver) ListTables(ctx context.Context, opts model.ListTablesOptions) ([]model.TableInfo, error) {
	if opts.Database != "" && opts.Database != d.database {
		return nil, nil
	}
	schema := d.resolveSchema(opts)
	rows, err := d.db.QueryContext(ctx, `
		SELECT table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = $1
		  AND table_type IN ('BASE TABLE', 'VIEW')
		ORDER BY table_name`, schema)
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

	schemaPtr := schemaLabel(schema)
	var items []model.TableInfo
	for _, entry := range entries {
		info := model.TableInfo{Name: entry.name, Schema: schemaPtr}
		if entry.typ == "VIEW" {
			info.Type = model.TableTypeView
			items = append(items, info)
			continue
		}
		fromRef := tableFromRef(schema, entry.name)
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
	pkCols, err := d.loadPrimaryKeyColumns(ctx, ref)
	if err != nil {
		return nil, err
	}
	pkSet := make(map[string]bool, len(pkCols))
	for _, c := range pkCols {
		pkSet[c] = true
	}
	for i := range cols {
		cols[i].PrimaryKey = pkSet[cols[i].Name]
	}
	indexes, err := d.loadIndexes(ctx, ref)
	if err != nil {
		return nil, err
	}
	return &model.TableSchema{
		Name:    ref.BareName,
		Type:    tableType,
		Schema:  schemaLabel(ref.Schema),
		Columns: cols,
		Indexes: indexes,
	}, nil
}

func (d *Driver) lookupTableType(ctx context.Context, ref tableRef) (model.TableType, error) {
	var typ string
	err := d.db.QueryRowContext(ctx, `
		SELECT table_type
		FROM information_schema.tables
		WHERE table_schema = $1 AND table_name = $2`, ref.Schema, ref.BareName).Scan(&typ)
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
	rows, err := d.db.QueryContext(ctx, `
		SELECT column_name, data_type, udt_name, is_nullable, column_default, ordinal_position
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position`, ref.Schema, ref.BareName)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var cols []model.ColumnInfo
	for rows.Next() {
		var name, dataType, udtName, nullable string
		var dflt sql.NullString
		var position int
		if err := rows.Scan(&name, &dataType, &udtName, &nullable, &dflt, &position); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		var def *string
		if dflt.Valid {
			def = &dflt.String
		}
		native := udtName
		cols = append(cols, model.ColumnInfo{
			Name:         name,
			DataType:     strings.ToUpper(dataType),
			NativeType:   &native,
			Nullable:     strings.EqualFold(nullable, "YES"),
			Position:     position,
			DefaultValue: def,
		})
	}
	return cols, rows.Err()
}

func (d *Driver) loadPrimaryKeyColumns(ctx context.Context, ref tableRef) ([]string, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.constraint_name = kcu.constraint_name
		 AND tc.table_schema = kcu.table_schema
		 AND tc.table_name = kcu.table_name
		WHERE tc.constraint_type = 'PRIMARY KEY'
		  AND tc.table_schema = $1
		  AND tc.table_name = $2
		ORDER BY kcu.ordinal_position`, ref.Schema, ref.BareName)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var cols []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

type indexRow struct {
	name    string
	column  string
	unique  bool
	primary bool
	ord     int
}

func (d *Driver) loadIndexes(ctx context.Context, ref tableRef) ([]model.IndexInfo, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT
			irel.relname AS index_name,
			idx.indisunique,
			idx.indisprimary,
			attr.attname AS column_name,
			keys.ordinality AS col_position
		FROM pg_class tbl
		JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
		JOIN pg_index idx ON idx.indrelid = tbl.oid
		JOIN pg_class irel ON irel.oid = idx.indexrelid
		JOIN LATERAL unnest(idx.indkey) WITH ORDINALITY AS keys(attnum, ordinality) ON true
		JOIN pg_attribute attr ON attr.attrelid = tbl.oid AND attr.attnum = keys.attnum
		WHERE ns.nspname = $1
		  AND tbl.relname = $2
		  AND tbl.relkind IN ('r', 'p')
		ORDER BY irel.relname, keys.ordinality`, ref.Schema, ref.BareName)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var entries []indexRow
	for rows.Next() {
		var entry indexRow
		if err := rows.Scan(&entry.name, &entry.unique, &entry.primary, &entry.column, &entry.ord); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	_ = rows.Close()

	byName := make(map[string]*model.IndexInfo)
	var order []string
	for _, entry := range entries {
		idx, ok := byName[entry.name]
		if !ok {
			idx = &model.IndexInfo{
				Name:    entry.name,
				Unique:  entry.unique,
				Primary: entry.primary,
			}
			byName[entry.name] = idx
			order = append(order, entry.name)
		}
		idx.Columns = append(idx.Columns, entry.column)
	}
	sort.Strings(order)
	indexes := make([]model.IndexInfo, 0, len(order))
	for _, name := range order {
		indexes = append(indexes, *byName[name])
	}
	return indexes, nil
}
