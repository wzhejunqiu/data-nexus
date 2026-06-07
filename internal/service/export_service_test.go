package service_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

func setupExportDB(t *testing.T, rowCount int) (*service.QueryService, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "export.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewConnectionManager(store, zap.NewNop())
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	qs := service.NewQueryService(mgr, nil, nil)
	_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}
	if rowCount > 0 {
		_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
			ConnectionID: conn.ID,
			SQL: fmt.Sprintf(`
				WITH RECURSIVE cnt(x) AS (
					SELECT 1 UNION ALL SELECT x + 1 FROM cnt WHERE x < %d
				)
				INSERT INTO items(name) SELECT 'row-' || x FROM cnt`, rowCount),
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return qs, conn.ID
}

func TestExportServiceExportTableToFile(t *testing.T) {
	qs, connID := setupExportDB(t, 5)
	exportSvc := service.NewExportService(qs)

	out := filepath.Join(t.TempDir(), "items.csv")
	err := exportSvc.ExportTableToFile(context.Background(), model.ExportTableCSVRequest{
		ConnectionID: connID,
		TableName:    "items",
		Format:       model.DefaultCSVFormat(),
	}, out, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "id") || !strings.Contains(content, "name") {
		t.Fatalf("expected header in csv: %q", content)
	}
	if strings.Count(content, "row-") != 5 {
		t.Fatalf("expected 5 data rows, got: %q", content)
	}
}

func TestExportServiceExportTableBeyondQueryLimit(t *testing.T) {
	qs, connID := setupExportDB(t, model.MaxQueryRows+50)
	exportSvc := service.NewExportService(qs)

	out := filepath.Join(t.TempDir(), "large.csv")
	err := exportSvc.ExportTableToFile(context.Background(), model.ExportTableCSVRequest{
		ConnectionID: connID,
		TableName:    "items",
		Format:       model.DefaultCSVFormat(),
	}, out, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), "row-") != model.MaxQueryRows+50 {
		t.Fatalf("expected %d rows exported", model.MaxQueryRows+50)
	}
}

func TestExportServiceExportSelectedColumns(t *testing.T) {
	qs, connID := setupExportDB(t, 3)
	exportSvc := service.NewExportService(qs)

	out := filepath.Join(t.TempDir(), "items-name-only.csv")
	err := exportSvc.ExportTableToFile(context.Background(), model.ExportTableCSVRequest{
		ConnectionID: connID,
		TableName:    "items",
		Format:       model.DefaultCSVFormat(),
		Columns:      []string{"name"},
	}, out, nil)
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	if strings.Contains(text, "\nid,") || strings.HasPrefix(text, "id,") {
		t.Fatalf("id column should not be exported: %q", text)
	}
	if !strings.Contains(text, "name") || strings.Count(text, "row-") != 3 {
		t.Fatalf("expected name column with 3 rows: %q", text)
	}
}

func TestExportServiceExportCancelled(t *testing.T) {
	qs, connID := setupExportDB(t, 5000)
	exportSvc := service.NewExportService(qs)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := filepath.Join(t.TempDir(), "cancel.csv")
	done := make(chan error, 1)
	go func() {
		done <- exportSvc.ExportTableToFile(ctx, model.ExportTableCSVRequest{
			ConnectionID: connID,
			TableName:    "items",
			Format:       model.DefaultCSVFormat(),
		}, out, func(exported int) {
			if exported >= 1000 {
				cancel()
			}
		})
	}()

	err := <-done
	if err == nil {
		t.Fatal("expected export cancelled")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "EXPORT_CANCELLED" {
		t.Fatalf("expected EXPORT_CANCELLED, got %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if _, statErr := os.Stat(out); statErr == nil {
		t.Fatal("partial export file should be removed")
	}
}
