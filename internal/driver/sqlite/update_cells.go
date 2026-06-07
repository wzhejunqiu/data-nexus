package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func (d *Driver) UpdateCells(ctx context.Context, tableName string, changes []model.CellChange, schema *model.TableSchema) (int, error) {
	if d.readOnly {
		return 0, model.ErrReadOnly()
	}
	if schema.Type == model.TableTypeView {
		return 0, model.ErrInvalidRequest("cannot edit view")
	}
	pkCols := primaryKeyColumns(schema)
	useRowid := false
	if len(pkCols) == 0 {
		withoutRowID, err := d.isWithoutRowID(ctx, tableName)
		if err != nil {
			return 0, err
		}
		if withoutRowID {
			return 0, model.ErrInvalidRequest("table has no primary key")
		}
		useRowid = true
	}
	colSet := columnNameSet(schema)

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, model.ErrSQL(err.Error())
	}
	defer func() { _ = tx.Rollback() }()

	updated := 0
	seen := make(map[string]struct{})
	for _, ch := range changes {
		if ch.ColumnName == "rowid" {
			return 0, model.ErrInvalidRequest("cannot edit rowid column")
		}
		if !identRe.MatchString(ch.ColumnName) || !colSet[ch.ColumnName] {
			return 0, model.ErrInvalidRequest("invalid column: " + ch.ColumnName)
		}
		if isBlobColumn(schema, ch.ColumnName) {
			return 0, model.ErrInvalidRequest("cannot edit BLOB column")
		}
		key := cellChangeKey(ch)
		if _, dup := seen[key]; dup {
			return 0, model.ErrInvalidRequest("duplicate cell change")
		}
		seen[key] = struct{}{}

		var whereParts []string
		var args []any
		args = append(args, sqlValue(ch.NewValue))
		if useRowid {
			rowid, ok := ch.PrimaryKey["rowid"]
			if !ok {
				return 0, model.ErrInvalidRequest("missing rowid")
			}
			whereParts = []string{"rowid = ?"}
			args = append(args, sqlValue(rowid))
		} else {
			for _, pk := range pkCols {
				if _, ok := ch.PrimaryKey[pk]; !ok {
					return 0, model.ErrInvalidRequest("missing primary key column: " + pk)
				}
			}
			whereParts = make([]string, len(pkCols))
			for i, pk := range pkCols {
				whereParts[i] = fmt.Sprintf("%q = ?", pk)
				args = append(args, sqlValue(ch.PrimaryKey[pk]))
			}
		}

		setClause := fmt.Sprintf("%q = ?", ch.ColumnName)
		query := fmt.Sprintf("UPDATE %q SET %s WHERE %s", tableName, setClause, strings.Join(whereParts, " AND "))
		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, model.ErrSQL(err.Error())
		}
		affected, _ := res.RowsAffected()
		updated += int(affected)
	}
	if err := tx.Commit(); err != nil {
		return 0, model.ErrSQL(err.Error())
	}
	return updated, nil
}

func primaryKeyColumns(schema *model.TableSchema) []string {
	var cols []string
	for _, c := range schema.Columns {
		if c.PrimaryKey {
			cols = append(cols, c.Name)
		}
	}
	return cols
}

func columnNameSet(schema *model.TableSchema) map[string]bool {
	m := make(map[string]bool, len(schema.Columns))
	for _, c := range schema.Columns {
		m[c.Name] = true
	}
	return m
}

func isBlobColumn(schema *model.TableSchema, name string) bool {
	for _, c := range schema.Columns {
		if c.Name == name {
			return strings.Contains(strings.ToUpper(c.DataType), "BLOB")
		}
	}
	return false
}

func cellChangeKey(ch model.CellChange) string {
	parts := make([]string, 0, len(ch.PrimaryKey)+1)
	parts = append(parts, ch.ColumnName)
	for k, v := range ch.PrimaryKey {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, "|")
}

func sqlValue(v any) any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		if t, ok := m["type"].(string); ok && t == "blob" {
			return nil
		}
	}
	return v
}
