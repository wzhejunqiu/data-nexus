package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
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
	item, err := store.Upsert(model.ConnectRequest{FilePath: dbPath})
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
	item, err := store.Upsert(model.ConnectRequest{FilePath: missing})
	if err != nil {
		t.Fatal(err)
	}
	_ = store.Save()

	mgr := service.NewConnectionManager(store, zap.NewNop())
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
	mgr := service.NewConnectionManager(store, zap.NewNop())
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
	mgr := service.NewConnectionManager(store, zap.NewNop())
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
