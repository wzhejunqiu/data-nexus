package service_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
)

func TestImportServiceAppend(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE dest (id INTEGER PRIMARY KEY, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}

	csvPath := writeTempCSV(t, "name\nalice\nbob\n")
	res, err := is.ImportCSV(ctx, model.ImportCSVRequest{
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
	if res.RowsInserted != 2 {
		t.Fatalf("expected 2 inserts, got %d", res.RowsInserted)
	}
}

func TestImportServiceNullValue(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE dest (id INTEGER PRIMARY KEY, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}

	csvPath := writeTempCSV(t, "name\n\\N\nok\n")
	fmt := model.DefaultCSVFormat()
	fmt.NullValue = "\\N"
	res, err := is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		TargetTable:  "dest",
		Mode:         "append",
		ColumnMap:    map[string]string{"name": "name"},
		FilePath:     csvPath,
		Format:       fmt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsInserted != 2 {
		t.Fatalf("expected 2 inserts, got %d", res.RowsInserted)
	}

	data, err := qs.BrowseRows(ctx, model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "dest",
		Page:         1,
		PageSize:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Rows[0]["name"] != nil {
		t.Fatalf("expected NULL name, got %+v", data.Rows[0]["name"])
	}
}

func TestImportServiceUpdateRequiresPK(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE nopk (name TEXT); INSERT INTO nopk VALUES ('a')",
	})
	if err != nil {
		t.Fatal(err)
	}

	csvPath := writeTempCSV(t, "name\nb\n")
	_, err = is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		TargetTable:  "nopk",
		Mode:         "update",
		ColumnMap:    map[string]string{"name": "name"},
		FilePath:     csvPath,
		Format:       model.DefaultCSVFormat(),
	})
	if err == nil {
		t.Fatal("expected upsert keys error")
	}
	if !strings.Contains(err.Error(), "upsert keys") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImportServiceMalformedCSV(t *testing.T) {
	_, qs, _ := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	dir := t.TempDir()
	badPath := filepath.Join(dir, "bad.csv")
	if err := os.WriteFile(badPath, []byte("\"unclosed\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := is.ParseCSVPreview(ctx, model.ParseCSVPreviewRequest{
		FilePath: badPath,
		MaxRows:  5,
		Format:   model.DefaultCSVFormat(),
	})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
