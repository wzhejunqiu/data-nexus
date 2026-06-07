package sqlite

import (
	"context"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

func (d *Driver) buildBrowseWhere(ctx context.Context, ref tableRef, opts model.BrowseOptions) (string, []any, error) {
	schema, err := d.GetTableSchema(ctx, ref.Qualified)
	if err != nil {
		return "", nil, err
	}
	where, args, err := sqlutil.BuildWhereClause(opts.Filters, schema)
	if err != nil {
		return "", nil, err
	}

	search := strings.TrimSpace(opts.Search)
	if search != "" {
		fts, err := d.DetectFTSTable(ctx, ref.BareName)
		if err != nil {
			return "", nil, err
		}
		if fts.Enabled && fts.FTSTableName != "" {
			ftsClause, ftsArgs, err := ftsSearchSubquery(fts.FTSTableName, search)
			if err != nil {
				return "", nil, err
			}
			if ftsClause != "" {
				if where == "" {
					where = " WHERE " + ftsClause
				} else {
					where += " AND " + ftsClause
				}
				args = append(args, ftsArgs...)
			}
		}
	}
	return where, args, nil
}
