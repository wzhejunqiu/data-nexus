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
	"go.uber.org/zap/zaptest/observer"
)

func newWailsTestEnv(t *testing.T) (*service.ConnectionManager, *service.QueryService, *model.Connection) {
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
	mgr := service.NewConnectionManager(store, zap.NewNop())
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	return mgr, service.NewQueryService(mgr), conn
}

func TestConnectionServiceListAndOpen(t *testing.T) {
	mgr, _, conn := newWailsTestEnv(t)
	svc := wailssvc.NewConnectionService(mgr, zap.NewNop())

	list, err := svc.ListConnections()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].ID != conn.ID {
		t.Fatalf("unexpected list: %+v", list.Items)
	}
}

func TestConnectionServiceCloseAndRename(t *testing.T) {
	mgr, _, conn := newWailsTestEnv(t)
	svc := wailssvc.NewConnectionService(mgr, zap.NewNop())

	if err := svc.CloseConnection(conn.ID); err != nil {
		t.Fatal(err)
	}
	renamed, err := svc.RenameConnection(conn.ID, "Renamed")
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Name != "Renamed" {
		t.Fatalf("unexpected name %s", renamed.Name)
	}
}

func TestConnectionServiceRestoreSettings(t *testing.T) {
	mgr, _, _ := newWailsTestEnv(t)
	svc := wailssvc.NewConnectionService(mgr, zap.NewNop())

	enabled, err := svc.GetRestoreOpenOnStartup()
	if err != nil || enabled {
		t.Fatalf("expected false by default, enabled=%v err=%v", enabled, err)
	}
	if err := svc.SetRestoreOpenOnStartup(true); err != nil {
		t.Fatal(err)
	}
	enabled, err = svc.GetRestoreOpenOnStartup()
	if err != nil || !enabled {
		t.Fatalf("expected true after set, enabled=%v err=%v", enabled, err)
	}
}

func TestQueryServiceExecuteAndClassify(t *testing.T) {
	_, qs, conn := newWailsTestEnv(t)
	svc := wailssvc.NewQueryService(qs, zap.NewNop())

	res, err := svc.Execute(model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "SELECT 1 AS n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != "result" {
		t.Fatalf("unexpected kind %s", res.Kind)
	}

	kind, err := svc.ClassifySQL(conn.ID, "CREATE TABLE t(id INTEGER)")
	if err != nil {
		t.Fatal(err)
	}
	if kind != string(model.StatementWrite) {
		t.Fatalf("got %q want write", kind)
	}
}

func TestSchemaAndTableServices(t *testing.T) {
	_, qs, conn := newWailsTestEnv(t)
	schemaSvc := wailssvc.NewSchemaService(qs, zap.NewNop())
	tableSvc := wailssvc.NewTableService(qs, zap.NewNop())
	querySvc := wailssvc.NewQueryService(qs, zap.NewNop())

	_, err := querySvc.Execute(model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = querySvc.Execute(model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "INSERT INTO items (name) VALUES ('a')",
	})
	if err != nil {
		t.Fatal(err)
	}

	tables, err := schemaSvc.ListTables(conn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(tables.Items) != 1 {
		t.Fatalf("unexpected tables: %+v", tables.Items)
	}

	tableSchema, err := schemaSvc.GetTableSchema(conn.ID, "items")
	if err != nil {
		t.Fatal(err)
	}
	if len(tableSchema.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(tableSchema.Columns))
	}

	data, err := tableSvc.BrowseRows(model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "items",
		Page:         1,
		PageSize:     50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 1 {
		t.Fatalf("expected 1 row, got %d", data.Pagination.TotalRows)
	}
}

func TestCallLogsAppError(t *testing.T) {
	mgr, _, _ := newWailsTestEnv(t)
	core, logs := observer.New(zap.InfoLevel)
	svc := wailssvc.NewConnectionService(mgr, zap.New(core))

	_, err := svc.OpenConnection("missing-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if logs.FilterMessageSnippet("service call failed").Len() == 0 {
		t.Fatal("expected failure log")
	}
	if logs.FilterField(zap.String("error_code", "SAVED_NOT_FOUND")).Len() == 0 {
		t.Fatal("expected error_code field in log")
	}
}

func TestAppServiceNilContextErrors(t *testing.T) {
	svc := wailssvc.NewAppService(zap.NewNop())

	for _, err := range []error{
		svc.SetWindowTitle("test"),
		svc.ShowAbout(),
		svc.EmitThemeChange("dark"),
		svc.EmitLanguageChange("en"),
	} {
		if err == nil {
			t.Fatal("expected error")
		}
		appErr, ok := err.(*model.AppError)
		if !ok || appErr.Code != "INTERNAL_ERROR" {
			t.Fatalf("expected INTERNAL_ERROR, got %v", err)
		}
	}
}

func TestAppServiceActiveConnection(t *testing.T) {
	svc := wailssvc.NewAppService(zap.NewNop())
	if svc.ActiveConnectionID() != "" {
		t.Fatal("expected empty active connection")
	}
	if err := svc.SetActiveConnection("conn-1"); err != nil {
		t.Fatal(err)
	}
	if svc.ActiveConnectionID() != "conn-1" {
		t.Fatalf("unexpected active id %s", svc.ActiveConnectionID())
	}
}
