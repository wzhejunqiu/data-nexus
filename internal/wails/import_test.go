package wails_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	wailssvc "github.com/wzhejunqiu/data-nexus/internal/wails"
	"go.uber.org/zap"
)

func TestImportServiceParseCSVPreview(t *testing.T) {
	_, qs, _ := newWailsTestEnv(t)
	svc := wailssvc.NewImportService(service.NewImportService(qs), zap.NewNop())

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(csvPath, []byte("name\nalice\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	preview, err := svc.ParseCSVPreview(model.ParseCSVPreviewRequest{
		FilePath: csvPath,
		Format:   model.DefaultCSVFormat(),
		MaxRows:  10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Rows) != 1 || preview.Headers[0] != "name" {
		t.Fatalf("unexpected preview: %+v", preview)
	}
}

func TestImportServiceImportCSV(t *testing.T) {
	_, qs, conn := newWailsTestEnv(t)
	querySvc := wailssvc.NewQueryService(qs, zap.NewNop())
	svc := wailssvc.NewImportService(service.NewImportService(qs), zap.NewNop())

	_, err := querySvc.Execute(model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE dest (id INTEGER PRIMARY KEY, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}

	csvPath := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(csvPath, []byte("name\nbob\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := svc.ImportCSV(model.ImportCSVRequest{
		ConnectionID: conn.ID,
		TargetTable:  "dest",
		Mode:         "append",
		ColumnMap:    map[string]string{"name": "name"},
		FilePath:     csvPath,
		Format:       model.DefaultCSVFormat(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsInserted != 1 {
		t.Fatalf("expected 1 insert, got %d", res.RowsInserted)
	}
}
