package remote

import (
	"fmt"
	"strings"
)

// BuildLexicographicKeysetWhere builds a lexicographic keyset WHERE clause for
// composite primary keys ordered ASC. It is compatible with MySQL 5.7+ and other
// dialects that do not support row-value constructors.
//
// quoteCol quotes a column identifier; placeholder is the per-value placeholder
// (e.g. "?" for MySQL). lastKey holds the last seen key values in column order.
// Returns the clause without a leading " WHERE " and the bind arguments.
func BuildLexicographicKeysetWhere(columns []string, quoteCol func(string) string, placeholder string, lastKey []any) (string, []any) {
	if len(columns) == 0 || len(lastKey) == 0 {
		return "", nil
	}
	if len(columns) == 1 {
		return fmt.Sprintf("%s > %s", quoteCol(columns[0]), placeholder), []any{lastKey[0]}
	}

	var branches []string
	var args []any
	for i := range columns {
		var parts []string
		for j := 0; j < i; j++ {
			parts = append(parts, fmt.Sprintf("%s = %s", quoteCol(columns[j]), placeholder))
			args = append(args, lastKey[j])
		}
		parts = append(parts, fmt.Sprintf("%s > %s", quoteCol(columns[i]), placeholder))
		args = append(args, lastKey[i])
		branches = append(branches, "("+strings.Join(parts, " AND ")+")")
	}
	return strings.Join(branches, " OR "), args
}
