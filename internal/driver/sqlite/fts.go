package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

func (d *Driver) DetectFTSTable(ctx context.Context, tableName string) (*model.FTSInfo, error) {
	ref, err := parseTableRef(tableName)
	if err != nil {
		return nil, err
	}
	info := &model.FTSInfo{
		Enabled:      false,
		Schema:       ref.Schema,
		ContentTable: ref.Qualified,
	}

	masterFrom := "sqlite_master"
	if ref.Schema != "main" {
		masterFrom = fmt.Sprintf("%s.sqlite_master", ref.Schema)
	}

	candidates := []string{
		ref.BareName + "_fts",
		"fts_" + ref.BareName,
		ref.BareName + "_fts5",
		ref.BareName + "_fts4",
	}
	for _, name := range candidates {
		ok, err := d.isFTSTable(ctx, masterFrom, name)
		if err != nil {
			return nil, err
		}
		if ok {
			info.Enabled = true
			info.FTSTableName = name
			return info, nil
		}
	}

	ok, err := d.isFTSTable(ctx, masterFrom, ref.BareName)
	if err != nil {
		return nil, err
	}
	if ok {
		info.Enabled = true
		info.FTSTableName = ref.BareName
	}
	return info, nil
}

func (d *Driver) isFTSTable(ctx context.Context, masterFrom, name string) (bool, error) {
	if !sqlutil.IsSafeQuotedIdentifier(name) {
		return false, nil
	}
	q := fmt.Sprintf(`SELECT sql FROM %s WHERE name = ? AND type = 'table'`, masterFrom)
	var sqlText string
	err := d.db.QueryRowContext(ctx, q, name).Scan(&sqlText)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, model.ErrSQL(err.Error())
	}
	upper := strings.ToUpper(sqlText)
	return strings.Contains(upper, "USING FTS5") || strings.Contains(upper, "USING FTS4"), nil
}

func ftsSearchSubquery(schema, ftsTableName, search string) (string, []any, error) {
	if !sqlutil.IsSafeQuotedIdentifier(ftsTableName) {
		return "", nil, model.ErrInvalidRequest("invalid fts table name")
	}
	if schema != "" && schema != "main" && !sqlutil.IsSafeQuotedIdentifier(schema) {
		return "", nil, model.ErrInvalidRequest("invalid fts schema")
	}
	if strings.TrimSpace(search) == "" {
		return "", nil, nil
	}
	var fromRef string
	if schema == "" || schema == "main" {
		fromRef = fmt.Sprintf("%q", ftsTableName)
	} else {
		fromRef = fmt.Sprintf("%s.%q", schema, ftsTableName)
	}
	matchRef := fmt.Sprintf("%q", ftsTableName)
	return fmt.Sprintf(` rowid IN (SELECT rowid FROM %s WHERE %s MATCH ?)`, fromRef, matchRef), []any{search}, nil
}
