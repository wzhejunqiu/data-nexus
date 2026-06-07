package remote

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestSerializeCellValue(t *testing.T) {
	if SerializeCellValue(nil) != nil {
		t.Fatal("expected nil")
	}
	blob := SerializeCellValue([]byte{1, 2, 3})
	m, ok := blob.(map[string]any)
	if !ok || m["type"] != "blob" || m["size"] != 3 {
		t.Fatalf("unexpected blob: %#v", blob)
	}
	if SerializeCellValue("text") != "text" {
		t.Fatal("expected plain value")
	}
}

func TestClassifySQL(t *testing.T) {
	cases := []struct {
		sql  string
		kind model.StatementKind
		err  bool
	}{
		{"", "", true},
		{"  ", "", true},
		{"SELECT 1", model.StatementQuery, false},
		{"WITH cte AS (SELECT 1) SELECT * FROM cte", model.StatementQuery, false},
		{"SHOW TABLES", model.StatementQuery, false},
		{"INSERT INTO t VALUES (1)", model.StatementWrite, false},
	}
	for _, tc := range cases {
		kind, err := ClassifySQL(tc.sql)
		if tc.err {
			if err == nil {
				t.Fatalf("expected error for %q", tc.sql)
			}
			continue
		}
		if err != nil || kind != tc.kind {
			t.Fatalf("sql %q: kind=%v err=%v", tc.sql, kind, err)
		}
	}
}

func TestQuoteDouble(t *testing.T) {
	if got := QuoteDouble(`a"b`); got != `"a""b"` {
		t.Fatalf("got %q", got)
	}
}

func TestQuoteBacktick(t *testing.T) {
	if got := QuoteBacktick("a`b"); got != "`a``b`" {
		t.Fatalf("got %q", got)
	}
}
