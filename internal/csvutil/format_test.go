package csvutil

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestFormatCellNullValue(t *testing.T) {
	if got := formatCell(nil, ""); got != "" {
		t.Fatalf("expected empty nullValue, got %q", got)
	}
	if got := formatCell(nil, "\\N"); got != "\\N" {
		t.Fatalf("expected \\N, got %q", got)
	}
	if got := formatCell("hello", "\\N"); got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}
}

func TestWriteRowsNullValue(t *testing.T) {
	var buf bytes.Buffer
	opts := model.DefaultCSVFormat()
	opts.NullValue = "NULL"
	opts.HasHeader = false
	rows := []map[string]any{
		{"id": int64(1), "name": nil},
		{"id": int64(2), "name": "ok"},
	}
	if err := WriteRows(&buf, []string{"id", "name"}, rows, opts); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "1,NULL") {
		t.Fatalf("expected NULL in output, got %q", out)
	}
	if !strings.Contains(out, "2,ok") {
		t.Fatalf("expected ok in output, got %q", out)
	}
}
