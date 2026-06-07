package sqlutil

import (
	"fmt"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

// BuildWhereClause builds a parameterized WHERE clause from structured filters.
// Returns empty clause and nil args when filters is empty.
func BuildWhereClause(filters []model.RowFilter, schema *model.TableSchema) (clause string, args []any, err error) {
	if len(filters) == 0 {
		return "", nil, nil
	}
	colSet := make(map[string]bool, len(schema.Columns))
	for _, c := range schema.Columns {
		colSet[c.Name] = true
	}

	var parts []string
	for _, f := range filters {
		if !IsSafeQuotedIdentifier(f.Column) || !colSet[f.Column] {
			return "", nil, model.ErrInvalidRequest(fmt.Sprintf("invalid filter column: %s", f.Column))
		}
		colRef := fmt.Sprintf("%q", f.Column)
		switch f.Operator {
		case model.FilterEq:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for eq")
			}
			parts = append(parts, colRef+" = ?")
			args = append(args, *f.Value)
		case model.FilterNe:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for ne")
			}
			parts = append(parts, colRef+" != ?")
			args = append(args, *f.Value)
		case model.FilterGt:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for gt")
			}
			parts = append(parts, colRef+" > ?")
			args = append(args, *f.Value)
		case model.FilterGte:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for gte")
			}
			parts = append(parts, colRef+" >= ?")
			args = append(args, *f.Value)
		case model.FilterLt:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for lt")
			}
			parts = append(parts, colRef+" < ?")
			args = append(args, *f.Value)
		case model.FilterLte:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for lte")
			}
			parts = append(parts, colRef+" <= ?")
			args = append(args, *f.Value)
		case model.FilterLike:
			if f.Value == nil {
				return "", nil, model.ErrInvalidRequest("filter value required for like")
			}
			val := *f.Value
			if !strings.Contains(val, "%") && !strings.Contains(val, "_") {
				val = "%" + val + "%"
			}
			parts = append(parts, colRef+" LIKE ?")
			args = append(args, val)
		case model.FilterIsNull:
			parts = append(parts, colRef+" IS NULL")
		case model.FilterIsNotNull:
			parts = append(parts, colRef+" IS NOT NULL")
		case model.FilterIn:
			if len(f.Values) == 0 {
				return "", nil, model.ErrInvalidRequest("filter values required for in")
			}
			placeholders := make([]string, len(f.Values))
			for i, v := range f.Values {
				placeholders[i] = "?"
				args = append(args, v)
			}
			parts = append(parts, colRef+" IN ("+strings.Join(placeholders, ", ")+")")
		default:
			return "", nil, model.ErrInvalidRequest(fmt.Sprintf("unsupported filter operator: %s", f.Operator))
		}
	}
	return " WHERE " + strings.Join(parts, " AND "), args, nil
}
