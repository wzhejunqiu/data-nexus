package csvutil

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestNewRowWriterAndClose(t *testing.T) {
	var buf bytes.Buffer
	opts := model.DefaultCSVFormat()
	opts.HasHeader = true
	opts.NullValue = "NULL"

	rw, err := NewRowWriter(&buf, []string{"id", "name"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := rw.WriteRows([]map[string]any{
		{"id": int64(1), "name": nil},
		{"id": int64(2), "name": "ok"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := rw.Close(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "id,name") || !strings.Contains(out, "1,NULL") || !strings.Contains(out, "2,ok") {
		t.Fatalf("unexpected csv output: %q", out)
	}
}

func TestOpenExportFileCRLFAndUTF8BOM(t *testing.T) {
	dir := t.TempDir()

	crlfPath := filepath.Join(dir, "out.csv")
	opts := model.DefaultCSVFormat()
	opts.LineEnding = "crlf"
	f, w, err := OpenExportFile(crlfPath, opts)
	if err != nil {
		t.Fatal(err)
	}
	rw, err := NewRowWriter(w, []string{"a"}, model.CSVFormatOptions{HasHeader: false})
	if err != nil {
		t.Fatal(err)
	}
	if err := rw.WriteRows([]map[string]any{{"a": "x"}}); err != nil {
		t.Fatal(err)
	}
	if err := rw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(crlfPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("x\r\n")) {
		t.Fatalf("expected CRLF line ending, got %q", data)
	}

	bomPath := filepath.Join(dir, "bom.csv")
	bomOpts := model.DefaultCSVFormat()
	bomOpts.Encoding = "utf-8-bom"
	bomOpts.LineEnding = "lf"
	bomOpts.HasHeader = false
	f2, w2, err := OpenExportFile(bomPath, bomOpts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w2.Write([]byte("hi")); err != nil {
		t.Fatal(err)
	}
	if err := f2.Close(); err != nil {
		t.Fatal(err)
	}
	bomData, err := os.ReadFile(bomPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(bomData, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatal("expected UTF-8 BOM prefix")
	}
}

func TestValidateExportPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"empty", "", true},
		{"parent traversal", "../secret.csv", true},
		{"bad extension", filepath.Join(t.TempDir(), "data.json"), true},
		{"valid csv", filepath.Join(t.TempDir(), "data.csv"), false},
		{"valid tsv", filepath.Join(t.TempDir(), "data.tsv"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateExportPath(tc.path)
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestOpenExportFileUnsupportedEncoding(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.csv")
	opts := model.DefaultCSVFormat()
	opts.Encoding = "latin1"
	_, _, err := OpenExportFile(path, opts)
	if err == nil {
		t.Fatal("expected unsupported encoding error")
	}
}
