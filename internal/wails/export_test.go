package wails_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	wailssvc "github.com/wzhejunqiu/data-nexus/internal/wails"
	"go.uber.org/zap"
)

func setupWailsExport(t *testing.T) (*wailssvc.ExportService, string, *mockRuntime) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "app.db")
	f, err := os.Create(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: dbPath})
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
	_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "INSERT INTO items (name) VALUES ('a')",
	})
	if err != nil {
		t.Fatal(err)
	}

	rt := &mockRuntime{}
	dialog := wailssvc.NewDialogServiceWithRuntime(zap.NewNop(), rt)
	dialog.SetContext(context.Background())
	svc := wailssvc.NewExportServiceWithRuntime(
		service.NewExportService(qs),
		dialog,
		zap.NewNop(),
		rt,
	)
	svc.SetContext(context.Background())
	return svc, conn.ID, rt
}

func TestExportServiceWithDefaultPath(t *testing.T) {
	svc, connID, _ := setupWailsExport(t)
	out := filepath.Join(t.TempDir(), "out.csv")
	path, err := svc.ExportTableCSV(model.ExportTableCSVRequest{
		ConnectionID: connID,
		TableName:    "items",
		DefaultPath:  out,
		Format:       model.DefaultCSVFormat(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != out {
		t.Fatalf("expected %s, got %s", out, path)
	}
}

func TestExportServiceCancelExport(t *testing.T) {
	svc, _, _ := setupWailsExport(t)
	if err := svc.CancelExportTableCSV(""); err == nil {
		t.Fatal("expected error for empty export id")
	}
	if err := svc.CancelExportTableCSV("missing"); err != nil {
		t.Fatal(err)
	}
}
