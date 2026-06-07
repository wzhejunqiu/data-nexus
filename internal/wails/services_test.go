package wails_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/executionlog/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/logger"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/secrets"
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
	mgr := service.NewTestConnectionManager(store)
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	return mgr, service.NewQueryService(mgr, nil, nil), conn
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

func TestCannedQueryServiceWails(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewQueryStore(filepath.Join(dir, "queries.json"))
	if err != nil {
		t.Fatal(err)
	}
	svc := wailssvc.NewCannedQueryService(service.NewCannedQueryService(store), zap.NewNop())

	list, err := svc.ListCannedQueries()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 0 {
		t.Fatalf("expected empty list, got %+v", list.Items)
	}

	saved, err := svc.SaveCannedQuery(model.SaveCannedQueryRequest{Name: "Q1", SQL: "SELECT 1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteCannedQuery(saved.ID); err != nil {
		t.Fatal(err)
	}
}

func TestConfigServiceWails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	logMgr, err := logger.NewManager(config.DefaultConfig().Log, true)
	if err != nil {
		t.Fatal(err)
	}
	svc := wailssvc.NewConfigService(logMgr, zap.NewNop())

	cfg, err := svc.GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("unexpected config level %s", cfg.Log.Level)
	}
	path, err := svc.GetConfigPath()
	if err != nil || path == "" {
		t.Fatalf("unexpected config path: %q err=%v", path, err)
	}
	cfg.Log.Level = "warn"
	if err := svc.UpdateConfig(cfg); err != nil {
		t.Fatal(err)
	}
	reloaded, err := svc.GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Log.Level != "warn" {
		t.Fatalf("expected warn after update, got %s", reloaded.Log.Level)
	}
}

func TestConnectionServiceCreateAndRemove(t *testing.T) {
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
	svc := wailssvc.NewConnectionService(mgr, zap.NewNop())

	saved, err := svc.CreateConnection(model.ConnectRequest{FilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == "" {
		t.Fatal("expected saved connection id")
	}

	list, err := svc.ListConnections()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].Status != model.ConnectionStatusClosed {
		t.Fatalf("expected one closed connection, got %+v", list.Items)
	}

	opened, err := svc.OpenConnection(saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	if opened.ID != saved.ID {
		t.Fatalf("expected opened id %s, got %s", saved.ID, opened.ID)
	}

	list, err = svc.ListConnections()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].Status != model.ConnectionStatusOpen {
		t.Fatalf("expected open connection, got %+v", list.Items)
	}

	if err := svc.RemoveConnection(saved.ID); err != nil {
		t.Fatal(err)
	}
	list, err = svc.ListConnections()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 0 {
		t.Fatalf("expected empty list after remove, got %+v", list.Items)
	}
}

func TestConnectionServiceAttachDetach(t *testing.T) {
	mgr, _, conn := newWailsTestEnv(t)
	svc := wailssvc.NewConnectionService(mgr, zap.NewNop())

	otherPath := filepath.Join(t.TempDir(), "other.db")
	f, err := os.Create(otherPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	if err := svc.Attach(conn.ID, otherPath, "other"); err != nil {
		t.Fatal(err)
	}
	attached, err := svc.ListAttached(conn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(attached) != 1 || attached[0].Alias != "other" {
		t.Fatalf("unexpected attached: %+v", attached)
	}
	if err := svc.Detach(conn.ID, "other"); err != nil {
		t.Fatal(err)
	}
}

func TestSchemaServiceGetTableProfile(t *testing.T) {
	_, qs, conn := newWailsTestEnv(t)
	querySvc := wailssvc.NewQueryService(qs, zap.NewNop())
	schemaSvc := wailssvc.NewSchemaService(qs, zap.NewNop())

	_, err := querySvc.Execute(model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE t (id INTEGER PRIMARY KEY, v TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = querySvc.Execute(model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "INSERT INTO t (v) VALUES ('x')",
	})
	if err != nil {
		t.Fatal(err)
	}

	profile, err := schemaSvc.GetTableProfile(conn.ID, "t")
	if err != nil {
		t.Fatal(err)
	}
	if profile.TotalRows == nil || *profile.TotalRows != 1 {
		t.Fatalf("expected 1 row, got %v", profile.TotalRows)
	}
}

func TestSqlExecutionServiceWails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sql-global.db")
	store, err := sqlite.NewStore(&model.ExecutionLogSQLiteConfig{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	svc := wailssvc.NewSqlExecutionService(service.NewSqlExecutionService(store), zap.NewNop())
	if err := store.Insert(context.Background(), model.SqlExecutionRecord{
		ConnectionID: "conn-1",
		SQL:          "SELECT 1",
		Kind:         model.SqlExecutionResult,
		EffectRows:   1,
		DurationMs:   1,
		ExecutedAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}

	history, err := svc.ListQueryHistory("conn-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) == 0 {
		t.Fatal("expected query history")
	}
	executions, err := svc.ListSqlExecutions("conn-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(executions.Items) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(executions.Items))
	}
}

func TestTableServiceUpdateCellsBatch(t *testing.T) {
	_, qs, conn := newWailsTestEnv(t)
	querySvc := wailssvc.NewQueryService(qs, zap.NewNop())
	tableSvc := wailssvc.NewTableService(qs, zap.NewNop())

	_, err := querySvc.Execute(model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE cells (id INTEGER PRIMARY KEY, val TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = querySvc.Execute(model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "INSERT INTO cells (val) VALUES ('old')",
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := tableSvc.BrowseRows(model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "cells",
		Page:         1,
		PageSize:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	rowID := data.Rows[0]["id"]

	res, err := tableSvc.UpdateCellsBatch(model.UpdateCellsBatchRequest{
		ConnectionID: conn.ID,
		TableName:    "cells",
		Changes: []model.CellChange{
			{ColumnName: "val", PrimaryKey: map[string]any{"id": rowID}, NewValue: "new"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.UpdatedCount != 1 {
		t.Fatalf("expected 1 update, got %d", res.UpdatedCount)
	}
}

func TestConnectionServiceCreateRemoteConnection(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManagerWithSecrets(store, secrets.NewMockStore())
	svc := wailssvc.NewConnectionService(mgr, zap.NewNop())

	saved, err := svc.CreateRemoteConnection(model.RemoteConnectRequest{
		Type:     model.DriverTypePostgres,
		Name:     "pg",
		Password: "secret",
		Postgres: &model.PostgresConfig{
			Host: "127.0.0.1", Port: 5432, Database: "db", User: "u",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Name != "pg" {
		t.Fatalf("unexpected name %s", saved.Name)
	}
}

func TestConnectionServiceOpenFromFile(t *testing.T) {
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
	svc := wailssvc.NewConnectionService(mgr, zap.NewNop())

	conn, err := svc.OpenConnectionFromFile(model.ConnectRequest{FilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	if conn.ID == "" {
		t.Fatal("expected connection id")
	}
}

func TestConnectionServiceUpdateSQLiteSettings(t *testing.T) {
	mgr, _, conn := newWailsTestEnv(t)
	svc := wailssvc.NewConnectionService(mgr, zap.NewNop())

	updated, err := svc.UpdateConnectionSQLiteSettings(conn.ID, model.SQLiteSettingsUpdate{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Config.SQLite == nil || !updated.Config.SQLite.ReadOnly {
		t.Fatalf("expected read-only, got %+v", updated.Config.SQLite)
	}
}

func TestDialogServiceOpenCSVFile(t *testing.T) {
	rt := &mockRuntime{filePath: "/tmp/data.csv"}
	svc := wailssvc.NewDialogServiceWithRuntime(zap.NewNop(), rt)
	svc.SetContext(context.Background())

	path, err := svc.OpenCSVFile()
	if err != nil {
		t.Fatal(err)
	}
	if path != "/tmp/data.csv" {
		t.Fatalf("unexpected path %s", path)
	}
}

func TestDialogServiceSaveFile(t *testing.T) {
	rt := &mockRuntime{savePath: "/tmp/save.csv"}
	svc := wailssvc.NewDialogServiceWithRuntime(zap.NewNop(), rt)
	svc.SetContext(context.Background())

	path, err := svc.SaveFile("out.csv", []model.FileFilter{{DisplayName: "CSV", Pattern: "*.csv"}})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/tmp/save.csv" {
		t.Fatalf("unexpected path %s", path)
	}
}

func TestSchemaServiceDetectFTSTable(t *testing.T) {
	_, qs, conn := newWailsTestEnv(t)
	querySvc := wailssvc.NewQueryService(qs, zap.NewNop())
	schemaSvc := wailssvc.NewSchemaService(qs, zap.NewNop())

	_, err := querySvc.Execute(model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE VIRTUAL TABLE docs USING fts5(title)",
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := schemaSvc.DetectFTSTable(conn.ID, "docs")
	if err != nil {
		t.Fatal(err)
	}
	if info == nil || !info.Enabled {
		t.Fatalf("expected FTS enabled, got %+v", info)
	}
}
