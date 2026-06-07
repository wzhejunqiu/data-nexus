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
	if !sqlutil.IsSafeQuotedIdentifier(tableName) {
		return nil, model.ErrTableNotFound(tableName)
	}
	info := &model.FTSInfo{Enabled: false, ContentTable: tableName}

	candidates := []string{
		tableName + "_fts",
		"fts_" + tableName,
		tableName + "_fts5",
		tableName + "_fts4",
	}
	for _, name := range candidates {
		var sqlText string
		err := d.db.QueryRowContext(ctx,
			`SELECT sql FROM sqlite_master WHERE name = ? AND type = 'table'`, name).Scan(&sqlText)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		upper := strings.ToUpper(sqlText)
		if strings.Contains(upper, "USING FTS5") || strings.Contains(upper, "USING FTS4") {
			info.Enabled = true
			info.FTSTableName = name
			return info, nil
		}
	}

	// Check if table itself is an FTS virtual table
	var sqlText string
	err := d.db.QueryRowContext(ctx,
		`SELECT sql FROM sqlite_master WHERE name = ? AND type = 'table'`, tableName).Scan(&sqlText)
	if err == nil {
		upper := strings.ToUpper(sqlText)
		if strings.Contains(upper, "USING FTS5") || strings.Contains(upper, "USING FTS4") {
			info.Enabled = true
			info.FTSTableName = tableName
		}
	}
	return info, nil
}

func ftsSearchSubquery(ftsTable, search string) (string, []any, error) {
	if !sqlutil.IsSafeQuotedIdentifier(ftsTable) {
		return "", nil, model.ErrInvalidRequest("invalid fts table name")
	}
	if strings.TrimSpace(search) == "" {
		return "", nil, nil
	}
	return fmt.Sprintf(` rowid IN (SELECT rowid FROM %q WHERE %q MATCH ?)`, ftsTable, ftsTable), []any{search}, nil
}
