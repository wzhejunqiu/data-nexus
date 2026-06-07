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

	first, err := store.Upsert(req)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Upsert(req)
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
	mgr := service.NewConnectionManager(store, zap.NewNop())

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
