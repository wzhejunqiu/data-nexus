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

func TestConnectionStoreSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "connections.json")
	store, err := service.NewConnectionStore(path)
	if err != nil {
		t.Fatal(err)
	}
	store.SetRestoreOpenOnStartup(true)
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	store2, err := service.NewConnectionStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if !store2.RestoreOpenOnStartup() {
		t.Fatal("expected restore flag true")
	}
}

func TestConnectionStoreRenameRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "connections.json")
	dbPath := filepath.Join(dir, "app.db")
	f, err := os.Create(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(path)
	if err != nil {
		t.Fatal(err)
	}
	item, err := store.UpsertSQLite(model.ConnectRequest{FilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	renamed, err := store.Rename(item.ID, "renamed.db")
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Name != "renamed.db" {
		t.Fatalf("unexpected name %s", renamed.Name)
	}
	if err := store.Remove(item.ID); err != nil {
		t.Fatal(err)
	}
}

func TestConnectionManagerOpenMissingFile(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "connections.json")
	store, err := service.NewConnectionStore(storePath)
	if err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing.db")
	item, err := store.UpsertSQLite(model.ConnectRequest{FilePath: missing})
	if err != nil {
		t.Fatal(err)
	}
	_ = store.Save()

	mgr := service.NewTestConnectionManager(store)
	_, err = mgr.OpenConnection(context.Background(), item.ID)
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "CONNECTION_FAILED" {
		t.Fatalf("expected CONNECTION_FAILED, got %v", err)
	}
}

func TestConnectionManagerRestoreSettings(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)
	if mgr.GetRestoreOpenOnStartup() {
		t.Fatal("expected false by default")
	}
	if err := mgr.SetRestoreOpenOnStartup(true); err != nil {
		t.Fatal(err)
	}
	if !mgr.GetRestoreOpenOnStartup() {
		t.Fatal("expected true after set")
	}
}

func TestConnectionManagerRename(t *testing.T) {
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
	renamed, err := mgr.RenameConnection(conn.ID, "My DB")
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Name != "My DB" {
		t.Fatalf("unexpected name %s", renamed.Name)
	}
}

func TestConnectionStore_FindByFilePath(t *testing.T) {
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
	item, err := store.UpsertSQLite(model.ConnectRequest{FilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}

	found, ok := store.FindByFilePath(dbPath)
	if !ok || found.ID != item.ID {
		t.Fatalf("expected to find connection by path, got ok=%v item=%+v", ok, found)
	}
}

func TestConnectionStore_FindByID_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, ok := store.FindByID("missing")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestConnectionStore_FreshCatalogInit(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	tree, err := service.NewConnectionGroupService(store.Catalog(), service.NewTestConnectionManager(store)).GetSidebarTree()
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Groups) < 1 {
		t.Fatal("expected default group on fresh catalog")
	}
}

func TestConnectionStore_UpdateSQLiteSettings_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.UpdateSQLiteSettings("missing", model.SQLiteSettingsUpdate{ReadOnly: true})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "SAVED_NOT_FOUND" {
		t.Fatalf("expected SAVED_NOT_FOUND, got %v", err)
	}
}

func TestConnectionStoreUpsertRemote(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}

	req := model.RemoteConnectRequest{
		Type: model.DriverTypePostgres,
		Name: "prod-pg",
		Postgres: &model.PostgresConfig{
			Host:     "db.example.com",
			Port:     5432,
			Database: "app",
			User:     "admin",
			Schema:   "public",
		},
	}
	first, err := store.UpsertRemote(req)
	if err != nil {
		t.Fatal(err)
	}
	if first.Name != "prod-pg" || first.Type != model.DriverTypePostgres {
		t.Fatalf("unexpected first upsert: %+v", first)
	}

	req.Name = ""
	second, err := store.UpsertRemote(req)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected same remote connection id, got %s vs %s", second.ID, first.ID)
	}
	if second.Name != "admin@db.example.com/app" {
		t.Fatalf("expected derived display name on upsert, got %q", second.Name)
	}

	mysqlReq := model.RemoteConnectRequest{
		Type: model.DriverTypeMySQL,
		MySQL: &model.MySQLConfig{
			Host:     "mysql.example.com",
			Database: "shop",
			User:     "root",
		},
	}
	mysqlItem, err := store.UpsertRemote(mysqlReq)
	if err != nil {
		t.Fatal(err)
	}
	if mysqlItem.Name != "root@mysql.example.com/shop" {
		t.Fatalf("unexpected mysql display name %q", mysqlItem.Name)
	}

	pwReq := model.RemoteConnectRequest{
		Type:     model.DriverTypePostgres,
		Password: "secret",
		Postgres: &model.PostgresConfig{
			Host:     "secure.example.com",
			Port:     5432,
			Database: "app",
			User:     "admin",
		},
	}
	pwItem, err := store.UpsertRemote(pwReq)
	if err != nil {
		t.Fatal(err)
	}
	if pwItem.Config.Postgres != nil && pwItem.Config.Postgres.Password != "" {
		t.Fatal("password must not be stored in connection config")
	}
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	reloaded, ok := store.FindByID(pwItem.ID)
	if !ok {
		t.Fatal("expected saved connection")
	}
	if reloaded.Config.Postgres != nil && reloaded.Config.Postgres.Password != "" {
		t.Fatal("password must not be stored in connection config after reload")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"password"`) {
		t.Fatalf("catalog.db must not contain password field")
	}
}

func TestConnectionStoreUpdatePostgresSettings(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	item, err := store.UpsertRemote(model.RemoteConnectRequest{
		Type: model.DriverTypePostgres,
		Name: "pg",
		Postgres: &model.PostgresConfig{
			Host: "h1", Port: 5432, Database: "db", User: "u",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.UpdatePostgresSettings(item.ID, model.PostgresSettingsUpdate{
		Host: "h2", Port: 5433, Database: "db2", User: "u2", Schema: "app",
	})
	if err != nil {
		t.Fatal(err)
	}
	pg := updated.Config.Postgres
	if pg.Host != "h2" || pg.Port != 5433 || pg.Schema != "app" {
		t.Fatalf("unexpected postgres settings: %+v", pg)
	}
}

func TestConnectionStoreUpdateMySQLSettings(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	item, err := store.UpsertRemote(model.RemoteConnectRequest{
		Type: model.DriverTypeMySQL,
		Name: "mysql",
		MySQL: &model.MySQLConfig{
			Host: "h1", Port: 3306, Database: "db", User: "u",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.UpdateMySQLSettings(item.ID, model.MySQLSettingsUpdate{
		Host: "h2", Port: 3307, Database: "db2", User: "u2", TLS: true, TLSSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	my := updated.Config.MySQL
	if my.Host != "h2" || my.Port != 3307 || !my.TLS || !my.TLSSkipVerify {
		t.Fatalf("unexpected mysql settings: %+v", my)
	}
}

func TestConnectionStoreUpdateMySQLSettingsPreservesStorageEngine(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	item, err := store.UpsertRemote(model.RemoteConnectRequest{
		Type: model.DriverTypeMySQL,
		Name: "mysql",
		MySQL: &model.MySQLConfig{
			Host: "h1", Port: 3306, Database: "db", User: "u",
			DefaultStorageEngine: "InnoDB",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.UpdateMySQLSettings(item.ID, model.MySQLSettingsUpdate{
		TLS: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Config.MySQL.DefaultStorageEngine != "InnoDB" {
		t.Fatalf("expected storage engine preserved, got %q", updated.Config.MySQL.DefaultStorageEngine)
	}
}
