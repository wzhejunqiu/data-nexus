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

func TestQueryServiceReadOnlyBlocksWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ro.db")
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
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{
		FilePath: path,
		ReadOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	qs := service.NewQueryService(mgr)
	_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE t(id INTEGER)",
	})
	if err == nil {
		t.Fatal("expected read-only error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "READ_ONLY" {
		t.Fatalf("expected READ_ONLY, got %v", err)
	}
}

func TestQueryServiceSelectReturnsResult(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
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

	qs := service.NewQueryService(mgr)
	res, err := qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "SELECT 1 AS n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != "result" || res.RowCount != 1 {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestQueryServiceDescTableReturnsSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
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

	qs := service.NewQueryService(mgr)
	_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE orders (id INTEGER, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "DESC orders",
	})
	if err != nil {
		t.Fatalf("DESC orders: %v", err)
	}
	if res.Kind != "result" || res.RowCount == 0 {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestQueryServiceWithDeleteUsesExecPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
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

	qs := service.NewQueryService(mgr)
	_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE orders (id INTEGER)",
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "WITH c AS (SELECT 1) DELETE FROM orders WHERE id = 1",
	})
	if err != nil {
		t.Fatalf("WITH DELETE: %v", err)
	}
	if res.Kind != "exec" {
		t.Fatalf("expected exec kind, got %+v", res)
	}
}

func TestQueryServiceClassifySQL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
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

	qs := service.NewQueryService(mgr)
	kind, err := qs.ClassifySQL(context.Background(), conn.ID, "CREATE TABLE z(id INTEGER)")
	if err != nil {
		t.Fatal(err)
	}
	if kind != model.StatementWrite {
		t.Fatalf("got %q want write", kind)
	}
}

func TestQueryServiceInvalidDescReturnsSQLError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
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

	qs := service.NewQueryService(mgr)
	_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "desc SELECT * from orders limit 10;",
	})
	if err == nil {
		t.Fatal("expected SQL error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "SQL_ERROR" {
		t.Fatalf("expected SQL_ERROR, got %v", err)
	}
}
