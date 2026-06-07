package remote

import (
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

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

func ClassifySQL(sqlText string) (model.StatementKind, error) {
	trimmed := strings.TrimSpace(sqlText)
	if trimmed == "" {
		return "", model.ErrInvalidRequest("sql is required")
	}
	upper := strings.ToUpper(trimmed)
	switch {
	case strings.HasPrefix(upper, "SELECT"),
		strings.HasPrefix(upper, "WITH"),
		strings.HasPrefix(upper, "SHOW"),
		strings.HasPrefix(upper, "EXPLAIN"),
		strings.HasPrefix(upper, "DESCRIBE"),
		strings.HasPrefix(upper, "PRAGMA"):
		return model.StatementQuery, nil
	default:
		return model.StatementWrite, nil
	}
}

func QuoteDouble(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func QuoteBacktick(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
