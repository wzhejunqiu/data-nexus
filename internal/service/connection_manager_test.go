package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/secrets"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestConnectionStoreUpsertSamePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "connections.json")
	store, err := service.NewConnectionStore(path)
	if err != nil {
		t.Fatal(err)
	}

	req := model.ConnectRequest{FilePath: filepath.Join(dir, "app.db"), ReadOnly: false}
	f, err := os.Create(req.FilePath)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	first, err := store.UpsertSQLite(req)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.UpsertSQLite(req)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected same id, got %s vs %s", first.ID, second.ID)
	}
}

func TestConnectionManagerMultipleOpen(t *testing.T) {
	dir := t.TempDir()
	db1 := filepath.Join(dir, "a.db")
	db2 := filepath.Join(dir, "b.db")
	for _, p := range []string{db1, db2} {
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
	}

	storePath := filepath.Join(dir, "connections.json")
	store, err := service.NewConnectionStore(storePath)
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)

	c1, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: db1})
	if err != nil {
		t.Fatal(err)
	}
	c2, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: db2})
	if err != nil {
		t.Fatal(err)
	}
	if c1.ID == c2.ID {
		t.Fatal("expected different connection ids")
	}

	list := mgr.ListConnections()
	open := 0
	for _, item := range list.Items {
		if item.Status == model.ConnectionStatusOpen {
			open++
		}
	}
	if open != 2 {
		t.Fatalf("expected 2 open connections, got %d", open)
	}

	if err := mgr.CloseConnection(context.Background(), c1.ID); err != nil {
		t.Fatal(err)
	}
	mgr.CloseAll()
}

func TestConnectionManagerPersistOpenConnections(t *testing.T) {
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
	if err := mgr.PersistOpenConnections(); err != nil {
		t.Fatal(err)
	}
	ids := store.OpenConnectionIDs()
	if len(ids) != 1 || ids[0] != conn.ID {
		t.Fatalf("unexpected open ids: %v", ids)
	}
}

func TestConnectionManagerRemoveOpen(t *testing.T) {
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
	if err := mgr.RemoveConnection(context.Background(), conn.ID); err != nil {
		t.Fatal(err)
	}
	list := mgr.ListConnections()
	if len(list.Items) != 0 {
		t.Fatalf("expected empty list, got %+v", list.Items)
	}
}

func TestConnectionManagerCreateWithoutOpen(t *testing.T) {
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
	saved, err := mgr.CreateConnection(context.Background(), model.ConnectRequest{FilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	list := mgr.ListConnections()
	if len(list.Items) != 1 || list.Items[0].Status != model.ConnectionStatusClosed {
		t.Fatalf("unexpected list: %+v", list)
	}
	if _, err := mgr.OpenConnection(context.Background(), saved.ID); err != nil {
		t.Fatal(err)
	}
}

func TestConnectionManagerUpdateReadOnly(t *testing.T) {
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

	_, err = mgr.UpdateConnectionSQLiteSettings(context.Background(), conn.ID, model.SQLiteSettingsUpdate{
		ReadOnly: true,
	})
	if err == nil {
		t.Fatal("expected CONNECTION_OPEN when updating open connection")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "CONNECTION_OPEN" {
		t.Fatalf("expected CONNECTION_OPEN, got %v", err)
	}

	if err := mgr.CloseConnection(context.Background(), conn.ID); err != nil {
		t.Fatal(err)
	}
	updated, err := mgr.UpdateConnectionSQLiteSettings(context.Background(), conn.ID, model.SQLiteSettingsUpdate{
		ReadOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Config.SQLite.ReadOnly {
		t.Fatal("expected read-only config")
	}
}

func TestOpenConnectionFromFile_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)

	_, err = mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_PATH" {
		t.Fatalf("expected INVALID_PATH, got %v", err)
	}
}

func TestOpenConnectionFromFile_Directory(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)

	_, err = mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: dir})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_PATH" {
		t.Fatalf("expected INVALID_PATH, got %v", err)
	}
}

func TestOpenConnectionFromFile_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)

	_, err = mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{
		FilePath: filepath.Join(dir, "missing.db"),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_PATH" {
		t.Fatalf("expected INVALID_PATH, got %v", err)
	}
}

func TestOpenConnectionFromFile_RelativePath(t *testing.T) {
	dir := t.TempDir()
	dbName := "app.db"
	f, err := os.Create(filepath.Join(dir, dbName))
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)

	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: dbName})
	if err != nil {
		t.Fatal(err)
	}
	if conn.ID == "" {
		t.Fatal("expected connection id")
	}
}

func TestOpenConnection_AlreadyOpen(t *testing.T) {
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

	_, err = mgr.OpenConnection(context.Background(), conn.ID)
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "CONNECTION_ALREADY_OPEN" {
		t.Fatalf("expected CONNECTION_ALREADY_OPEN, got %v", err)
	}
}

func TestCloseConnection_NotOpen(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)

	err = mgr.CloseConnection(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "CONNECTION_NOT_FOUND" {
		t.Fatalf("expected CONNECTION_NOT_FOUND, got %v", err)
	}
}

func TestRestoreConnectionsOnStartup_Disabled(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)
	if err := mgr.RestoreConnectionsOnStartup(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRestoreConnectionsOnStartup_RestoresOpenConnection(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "app.db")
	f, err := os.Create(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	storePath := filepath.Join(dir, "connections.json")
	store, err := service.NewConnectionStore(storePath)
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.SetRestoreOpenOnStartup(true); err != nil {
		t.Fatal(err)
	}
	if err := mgr.PersistOpenConnections(); err != nil {
		t.Fatal(err)
	}
	mgr.CloseAll()

	store2, err := service.NewConnectionStore(storePath)
	if err != nil {
		t.Fatal(err)
	}
	mgr2 := service.NewTestConnectionManager(store2)
	if err := mgr2.RestoreConnectionsOnStartup(context.Background()); err != nil {
		t.Fatal(err)
	}
	list := mgr2.ListConnections()
	if len(list.Items) != 1 || list.Items[0].Status != model.ConnectionStatusOpen {
		t.Fatalf("expected restored open connection, got %+v", list.Items)
	}
	if list.Items[0].ID != conn.ID {
		t.Fatalf("expected id %s, got %s", conn.ID, list.Items[0].ID)
	}
}

func TestRestoreConnectionsOnStartup_MissingFileLogsWarning(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "gone.db")
	storePath := filepath.Join(dir, "connections.json")

	store, err := service.NewConnectionStore(storePath)
	if err != nil {
		t.Fatal(err)
	}
	item, err := store.UpsertSQLite(model.ConnectRequest{FilePath: missing})
	if err != nil {
		t.Fatal(err)
	}
	store.SetRestoreOpenOnStartup(true)
	store.SetOpenConnectionIDs([]string{item.ID})
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	core, logs := observer.New(zap.WarnLevel)
	mgr := service.NewConnectionManager(store, secrets.NewMockStore(), zap.New(core))
	if err := mgr.RestoreConnectionsOnStartup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if logs.FilterMessageSnippet("failed to restore connection").Len() == 0 {
		t.Fatal("expected warn log for failed restore")
	}
}

func TestConnectionManagerAttachDetach(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.db")
	otherPath := filepath.Join(dir, "other.db")
	for _, p := range []string{mainPath, otherPath} {
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
	}

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: mainPath})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	if err := mgr.AttachDatabase(ctx, conn.ID, otherPath, "other"); err != nil {
		t.Fatal(err)
	}
	attached, err := mgr.ListAttachedDatabases(ctx, conn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(attached) != 1 || attached[0].Alias != "other" {
		t.Fatalf("unexpected attached: %+v", attached)
	}
	if err := mgr.DetachDatabase(ctx, conn.ID, "other"); err != nil {
		t.Fatal(err)
	}
	attached, err = mgr.ListAttachedDatabases(ctx, conn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(attached) != 0 {
		t.Fatalf("expected no attached databases, got %+v", attached)
	}
}

func TestConnectionManagerAttachValidation(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)
	ctx := context.Background()

	for _, err := range []error{
		mgr.AttachDatabase(ctx, "", "/tmp/x.db", "x"),
		mgr.AttachDatabase(ctx, "id", "", "x"),
		mgr.AttachDatabase(ctx, "id", "/tmp/x.db", ""),
	} {
		if err == nil {
			t.Fatal("expected validation error")
		}
	}
}

func TestConnectionManagerTestConnectionInvalidRequest(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)
	ctx := context.Background()

	if err := mgr.TestConnection(ctx, model.TestConnectionRequest{Type: model.DriverTypePostgres}); err == nil {
		t.Fatal("expected validation error")
	}
	if err := mgr.TestConnection(ctx, model.TestConnectionRequest{
		Type:     model.DriverTypePostgres,
		Postgres: &model.PostgresConfig{Host: "", Database: "db", User: "u"},
	}); err == nil {
		t.Fatal("expected host required")
	}
	if err := mgr.TestConnection(ctx, model.TestConnectionRequest{
		Type:     model.DriverTypePostgres,
		Postgres: &model.PostgresConfig{Host: "localhost", Database: "", User: "u"},
	}); err == nil {
		t.Fatal("expected postgres database required")
	}
	if _, err := mgr.CreateRemoteConnection(ctx, model.RemoteConnectRequest{
		Type:     model.DriverTypeMySQL,
		Password: "secret",
		MySQL:    &model.MySQLConfig{Host: "localhost", User: "root"},
	}); err != nil {
		t.Fatalf("expected mysql without database to pass validation, got %v", err)
	}
}

func TestConnectionManagerUpdateRemoteSettings(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mockSecrets := secrets.NewMockStore()
	mgr := service.NewTestConnectionManagerWithSecrets(store, mockSecrets)
	ctx := context.Background()

	saved, err := mgr.CreateRemoteConnection(ctx, model.RemoteConnectRequest{
		Type: model.DriverTypePostgres,
		Name: "pg",
		Postgres: &model.PostgresConfig{
			Host: "127.0.0.1", Port: 5432, Database: "db", User: "u",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := mgr.UpdateConnectionPostgresSettings(ctx, saved.ID, model.PostgresSettingsUpdate{
		Host: "10.0.0.1", Port: 5433,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Config.Postgres.Host != "10.0.0.1" {
		t.Fatalf("unexpected host %s", updated.Config.Postgres.Host)
	}

	mysqlSaved, err := mgr.CreateRemoteConnection(ctx, model.RemoteConnectRequest{
		Type: model.DriverTypeMySQL,
		Name: "mysql",
		MySQL: &model.MySQLConfig{
			Host: "127.0.0.1", Port: 3306, Database: "db", User: "u",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	mysqlUpdated, err := mgr.UpdateConnectionMySQLSettings(ctx, mysqlSaved.ID, model.MySQLSettingsUpdate{
		Host: "10.0.0.2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if mysqlUpdated.Config.MySQL.Host != "10.0.0.2" {
		t.Fatalf("unexpected mysql host %s", mysqlUpdated.Config.MySQL.Host)
	}
}
