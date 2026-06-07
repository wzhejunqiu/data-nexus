package remote

import (
	"fmt"
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

func TestBuildLexicographicKeysetWhere(t *testing.T) {
	quote := QuoteBacktick
	cases := []struct {
		name    string
		cols    []string
		lastKey []any
		want    string
		wantArg []any
	}{
		{
			name:    "single column",
			cols:    []string{"id"},
			lastKey: []any{int64(10)},
			want:    "`id` > ?",
			wantArg: []any{int64(10)},
		},
		{
			name:    "two columns",
			cols:    []string{"k1", "k2"},
			lastKey: []any{int64(1), int64(2)},
			want:    "(`k1` > ?) OR (`k1` = ? AND `k2` > ?)",
			wantArg: []any{int64(1), int64(1), int64(2)},
		},
		{
			name:    "three columns",
			cols:    []string{"a", "b", "c"},
			lastKey: []any{"x", "y", "z"},
			want:    "(`a` > ?) OR (`a` = ? AND `b` > ?) OR (`a` = ? AND `b` = ? AND `c` > ?)",
			wantArg: []any{"x", "x", "y", "x", "y", "z"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, args := BuildLexicographicKeysetWhere(tc.cols, quote, "?", tc.lastKey)
			if got != tc.want {
				t.Fatalf("clause: got %q want %q", got, tc.want)
			}
			if len(args) != len(tc.wantArg) {
				t.Fatalf("args len: got %d want %d", len(args), len(tc.wantArg))
			}
			for i := range args {
				if fmt.Sprint(args[i]) != fmt.Sprint(tc.wantArg[i]) {
					t.Fatalf("args[%d]: got %v want %v", i, args[i], tc.wantArg[i])
				}
			}
		})
	}
}
