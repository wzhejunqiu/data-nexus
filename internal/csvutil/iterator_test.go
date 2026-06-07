package csvutil

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestReadPreviewTotalRowCount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte("id,name\n1,a\n2,b\n3,c\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	preview, err := ReadPreview(path, 2, model.DefaultCSVFormat())
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Rows) != 2 {
		t.Fatalf("expected 2 preview rows, got %d", len(preview.Rows))
	}
	if preview.RowCount != 3 {
		t.Fatalf("expected total row count 3, got %d", preview.RowCount)
	}
}

func TestIteratorReadAllRows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte("id\n1\n2\n3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	it, err := OpenIterator(path, model.DefaultCSVFormat())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = it.Close() }()

	total := 0
	for {
		batch, err := it.ReadBatch(2)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		total += len(batch)
	}
	if total != 3 {
		t.Fatalf("expected 3 rows, got %d", total)
	}
}

func TestIteratorHeaders(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte("id,name\n1,a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	it, err := OpenIterator(path, model.DefaultCSVFormat())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = it.Close() }()

	headers := it.Headers()
	if len(headers) != 2 || headers[0] != "id" || headers[1] != "name" {
		t.Fatalf("unexpected headers: %v", headers)
	}
	// copy-on-read
	headers[0] = "mutated"
	if it.Headers()[0] != "id" {
		t.Fatal("Headers should return a copy")
	}
}

func TestIteratorEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.csv")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	it, err := OpenIterator(path, model.DefaultCSVFormat())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = it.Close() }()
	if len(it.Headers()) != 0 {
		t.Fatalf("expected empty headers, got %v", it.Headers())
	}
}

func TestIteratorNoHeaderRow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte("1,a\n2,b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fmt := model.DefaultCSVFormat()
	fmt.HasHeader = false
	it, err := OpenIterator(path, fmt)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = it.Close() }()
	if it.Headers()[0] != "column_1" {
		t.Fatalf("unexpected headers: %v", it.Headers())
	}
}
