package export

import (
	"sort"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

// ResolveStableRowKey picks a deterministic ORDER BY key for full-table export.
// UNIQUE index fallback is intentionally omitted (Phase 2); prefer explicit error over guessed order.
func ResolveStableRowKey(schema *model.TableSchema, dialect model.DriverType, withoutRowID bool) (model.StableRowKey, error) {
	for _, idx := range schema.Indexes {
		if idx.Primary && len(idx.Columns) > 0 {
			return model.StableRowKey{
				Columns: append([]string(nil), idx.Columns...),
				Source:  "primary_key",
			}, nil
		}
	}
	var pk []model.ColumnInfo
	for _, c := range schema.Columns {
		if c.PrimaryKey {
			pk = append(pk, c)
		}
	}
	sort.Slice(pk, func(i, j int) bool { return pk[i].Position < pk[j].Position })
	if len(pk) > 0 {
		cols := make([]string, len(pk))
		for i, c := range pk {
			cols[i] = c.Name
		}
		return model.StableRowKey{Columns: cols, Source: "primary_key"}, nil
	}
	if dialect == model.DriverTypeSQLite {
		if withoutRowID {
			return model.StableRowKey{}, model.ErrExportNoStableKey(schema.Name)
		}
		return model.StableRowKey{Columns: []string{"rowid"}, Source: "rowid"}, nil
	}
	return model.StableRowKey{}, model.ErrExportNoStableKey(schema.Name)
}
