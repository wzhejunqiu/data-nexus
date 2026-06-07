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

func TestLineEndingAndEncodeContent(t *testing.T) {
	lf := model.CSVFormatOptions{LineEnding: "lf"}
	if LineEnding(lf) != "\n" {
		t.Fatalf("expected LF, got %q", LineEnding(lf))
	}
	if LineEnding(model.DefaultCSVFormat()) != "\r\n" {
		t.Fatal("expected CRLF by default")
	}

	body := "a\nb\r\nc"
	if got := EncodeContent(body, lf); got != "a\nb\nc" {
		t.Fatalf("expected normalized LF content, got %q", got)
	}
	crlf := model.DefaultCSVFormat()
	if got := EncodeContent("a\nb", crlf); got != "a\r\nb" {
		t.Fatalf("expected CRLF content, got %q", got)
	}
}
